package pve

import (
	"bytes"
	"crypto/des"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/buger/goterm"
	"github.com/gorilla/websocket"
	"github.com/hilaoyu/go-utils/utilBuf"
	"github.com/hilaoyu/go-utils/utilLogger"
	"github.com/hilaoyu/go-utils/utils"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	DefaultUserAgent = "hilaoyu-go-pve-client"
)

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Otp      string `json:"otp,omitempty"`
	Path     string `json:"path,omitempty"`
	Privs    string `json:"privs,omitempty"`
	Realm    string `json:"realm,omitempty"`
}

type Session struct {
	Username            string `json:"username"`
	CsrfPreventionToken string `json:"CSRFPreventionToken,omitempty"`
	ClusterName         string `json:"clustername,omitempty"`
	Ticket              string `json:"ticket,omitempty"`
}

type Version struct {
	Release string `json:"release"`
	RepoID  string `json:"repoid"`
	Version string `json:"version"`
}

var (
	bufCopy          = utilBuf.NewBufCopy()
	ErrNotAuthorized = errors.New("not authorized to access endpoint")
	ErrTimeout       = errors.New("the operation has timed out")
)

func IsNotAuthorized(err error) bool {
	return err == ErrNotAuthorized
}

func IsTimeout(err error) bool {
	return err == ErrTimeout
}

type Client struct {
	httpClient  *http.Client
	userAgent   string
	baseURL     string
	token       string
	credentials *Credentials
	version     *Version
	session     *Session
	logger      *utilLogger.Logger
}

func NewClient(baseURL string, opts ...Option) *Client {
	c := &Client{
		baseURL:   baseURL,
		userAgent: DefaultUserAgent,
		logger:    utilLogger.NewLogger(),
	}

	c.logger.AddConsoleWriter()

	for _, o := range opts {
		o(c)
	}

	if c.httpClient == nil {
		c.httpClient = http.DefaultClient
	}

	return c
}

func (c *Client) Login(username, password string) error {
	_, err := c.Ticket(&Credentials{
		Username: username,
		Password: password,
	})

	return err
}

func (c *Client) APIToken(tokenID, secret string) {
	c.token = fmt.Sprintf("%s=%s", tokenID, secret)
}

func (c *Client) Ticket(credentials *Credentials) (*Session, error) {
	return c.session, c.Post("/access/ticket", credentials, &c.session)
}

func (c *Client) Version() (*Version, error) {
	return c.version, c.Get("/version", &c.version)
}

func (c *Client) Req(method, path string, data []byte, v interface{}) error {
	if strings.HasPrefix(path, "/") {
		path = c.baseURL + path
	}

	c.logger.InfoF("SEND: %s - %s", method, path)

	var body io.Reader
	if data != nil {
		if path != (c.baseURL + "/access/ticket") {
			// dont show passwords in the logs
			if len(data) < 2048 {
				c.logger.DebugF("DATA: %s", string(data))
			} else {
				c.logger.DebugF("DATA: %s", "truncated due to length")
			}
		}

		body = bytes.NewBuffer(data)
	}

	req, err := http.NewRequest(method, path, body)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Add("Content-Type", "application/json")
	}

	c.authHeaders(&req.Header)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusUnauthorized {
		if c.credentials != nil && c.session == nil {
			// credentials passed but no session started, try a login and retry the request
			if _, err := c.Ticket(c.credentials); err != nil {
				return err
			}
			return c.Req(method, path, data, v)
		}
		return ErrNotAuthorized
	}

	return c.handleResponse(res, v)

}

func (c *Client) Get(p string, v interface{}) error {
	return c.Req(http.MethodGet, p, nil, v)
}

func (c *Client) Post(p string, d interface{}, v interface{}) error {
	var data []byte
	if d != nil {
		var err error
		data, err = json.Marshal(d)
		if err != nil {
			return err
		}
	}

	return c.Req(http.MethodPost, p, data, v)
}

func (c *Client) Upload(path string, fields map[string]string, file *os.File, v interface{}) error {
	if strings.HasPrefix(path, "/") {
		path = c.baseURL + path
	}

	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	for name, val := range fields {
		if err := w.WriteField(name, val); err != nil {
			return err
		}
	}

	if _, err := w.CreateFormFile("filename", filepath.Base(file.Name())); err != nil {
		return err
	}

	header := b.Len()
	if err := w.Close(); err != nil {
		return err
	}

	body := io.MultiReader(bytes.NewReader(b.Bytes()[:header]),
		file,
		bytes.NewReader(b.Bytes()[header:]))

	req, err := http.NewRequest(http.MethodPost, path, body)
	if err != nil {
		return err
	}

	fi, err := file.Stat()
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", w.FormDataContentType())
	req.ContentLength = int64(b.Len()) + fi.Size()
	c.authHeaders(&req.Header)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	return c.handleResponse(res, &v)
}

func (c *Client) Put(p string, d interface{}, v interface{}) error {
	var data []byte
	if d != nil {
		var err error
		data, err = json.Marshal(d)
		if err != nil {
			return err
		}
	}

	return c.Req(http.MethodPut, p, data, v)
}

func (c *Client) Delete(p string, v interface{}) error {
	return c.Req(http.MethodDelete, p, nil, v)
}

func (c *Client) authHeaders(header *http.Header) {
	header.Add("User-Agent", c.userAgent)
	header.Add("Accept", "application/json")
	if c.token != "" {
		header.Add("Authorization", "PVEAPIToken="+c.token)
	} else if c.session != nil {
		header.Add("Cookie", "PVEAuthCookie="+c.session.Ticket)
		header.Add("CSRFPreventionToken", c.session.CsrfPreventionToken)
	}
}

func (c *Client) handleResponse(res *http.Response, v interface{}) error {
	if res.StatusCode == http.StatusInternalServerError ||
		res.StatusCode == http.StatusNotImplemented {
		return errors.New(res.Status)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	c.logger.DebugF("RECV: %d - %s", res.StatusCode, res.Status)

	if res.Request.URL.String() != (c.baseURL + "/access/ticket") {
		// dont show tokens out of the logs
		c.logger.DebugF("BODY: %s", string(body))
	}
	if res.StatusCode == http.StatusBadRequest {
		var errorskey map[string]json.RawMessage
		if err := json.Unmarshal(body, &errorskey); err != nil {
			return err
		}

		if body, ok := errorskey["errors"]; ok {
			return fmt.Errorf("bad request: %s - %s", res.Status, body)
		}

		return fmt.Errorf("bad request: %s - %s", res.Status, string(body))
	}

	// account for everything being in a data key
	var datakey map[string]json.RawMessage
	if err := json.Unmarshal(body, &datakey); err != nil {
		return err
	}

	if body, ok := datakey["data"]; ok {
		return json.Unmarshal(body, &v)
	}

	return json.Unmarshal(body, &v) // assume passed in type fully supports response
}

func (c *Client) VNCWebSocket(path string, vnc *VNC) (chan string, chan string, chan error, func() error, error) {
	if strings.HasPrefix(path, "/") {
		path = strings.Replace(c.baseURL, "https://", "wss://", 1) + path
	}

	var tlsConfig *tls.Config
	transport := c.httpClient.Transport.(*http.Transport)
	if transport != nil {
		tlsConfig = transport.TLSClientConfig
	}
	c.logger.DebugF("connecting to websocket: %s", path)
	dialer := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 30 * time.Second,
		TLSClientConfig:  tlsConfig,
	}

	dialerHeaders := http.Header{}
	c.authHeaders(&dialerHeaders)

	conn, _, err := dialer.Dial(path, dialerHeaders)

	if err != nil {
		return nil, nil, nil, nil, err
	}

	// start the session by sending user@realm:ticket
	if err := conn.WriteMessage(websocket.BinaryMessage, []byte(vnc.User+":"+vnc.Ticket+"\n")); err != nil {
		return nil, nil, nil, nil, err
	}

	// it sends back the same thing you just sent so catch it drop it
	_, msg, err := conn.ReadMessage()
	if err != nil || string(msg) != "OK" {
		if err := conn.Close(); err != nil {
			return nil, nil, nil, nil, fmt.Errorf("error closing websocket: %s", err.Error())
		}
		return nil, nil, nil, nil, fmt.Errorf("unable to establish websocket: %s", err.Error())
	}

	type size struct {
		height int
		width  int
	}
	// start the session by sending user@realm:ticket
	tsize := size{
		height: goterm.Height(),
		width:  goterm.Width(),
	}

	c.logger.DebugF("sending terminal size: %d x %d", tsize.height, tsize.width)
	if err := conn.WriteMessage(websocket.BinaryMessage, []byte(fmt.Sprintf("1:%d:%d:", tsize.height, tsize.width))); err != nil {
		return nil, nil, nil, nil, err
	}

	send := make(chan string)
	recv := make(chan string)
	errors := make(chan error)
	done := make(chan struct{})
	ticker := time.NewTicker(30 * time.Second)
	resize := make(chan size)

	go func(tsize size) {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				resized := size{
					height: goterm.Height(),
					width:  goterm.Width(),
				}
				if tsize.height != resized.height ||
					tsize.width != resized.width {
					tsize = resized
					resize <- resized
				}
			}
		}
	}(tsize)

	closer := func() error {
		close(done)
		time.Sleep(1 * time.Second)
		close(send)
		close(recv)
		close(errors)
		ticker.Stop()

		return conn.Close()
	}

	go func() {
		for {
			select {
			case <-done:
				return
			default:
				_, msg, err := conn.ReadMessage()
				if err != nil {
					if strings.Contains(err.Error(), "use of closed network connection") {
						return
					}
					if !websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
						return
					}
					errors <- err
				}
				recv <- string(msg)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-done:
				if err := conn.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
					errors <- err
				}
				return
			case <-ticker.C:
				c.logger.DebugF("sending wss keep alive")
				if err := conn.WriteMessage(websocket.BinaryMessage, []byte("2")); err != nil {
					errors <- err
				}
			case resized := <-resize:
				c.logger.DebugF("resizing terminal window: %d x %d", resized.height, resized.width)
				if err := conn.WriteMessage(websocket.BinaryMessage, []byte(fmt.Sprintf("1:%d:%d:", resized.height, resized.width))); err != nil {
					errors <- err
				}
			case msg := <-send:
				c.logger.DebugF("sending: %s", string(msg))
				m := []byte(msg)
				send := append([]byte(fmt.Sprintf("0:%d:", len(m))), m...)
				if err := conn.WriteMessage(websocket.BinaryMessage, send); err != nil {
					errors <- err
				}
				if err := conn.WriteMessage(websocket.BinaryMessage, []byte("0:1:\n")); err != nil {
					errors <- err
				}
			}
		}
	}()

	return send, recv, errors, closer, nil
}

func (c *Client) VNCProxyWebsocketServeHTTP(path string, vnc *VNC, w http.ResponseWriter, r *http.Request, responseHeader http.Header) (err error) {
	upgrader := websocket.Upgrader{}
	websocketServe, err := upgrader.Upgrade(w, r, responseHeader)
	if err != nil {
		err = fmt.Errorf("upgrade http to websocket err: %+v", err)
		return
	}

	if strings.HasPrefix(path, "/") {
		path = strings.Replace(c.baseURL, "https://", "wss://", 1) + path
	}

	var tlsConfig *tls.Config
	transport := c.httpClient.Transport.(*http.Transport)
	if transport != nil {
		tlsConfig = transport.TLSClientConfig
	}
	c.logger.DebugF("connecting to websocket: %s", path)
	dialer := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 30 * time.Second,
		TLSClientConfig:  tlsConfig,
	}

	dialerHeaders := http.Header{}
	c.authHeaders(&dialerHeaders)

	pveVncConn, _, err := dialer.Dial(path, dialerHeaders)

	if err != nil {
		err = fmt.Errorf("connect to pve err: %+v", err)
		websocketServe.Close()
		return
	}

	defer func() {
		pveVncConn.Close()
		websocketServe.Close()
	}()

	var msgType int
	var msg []byte
	msgType, msg, err = pveVncConn.ReadMessage()
	if nil != err {
		err = fmt.Errorf("read pve rfb message err: %+v", err)
		return
	}
	//fmt.Println("pveVncConn1", msgType, msg, err, string(msg))
	//"RFB 003.008\n"
	if err = websocketServe.WriteMessage(msgType, msg); err != nil {
		err = fmt.Errorf("write websocket rfb message err: %+v", err)
		return
	}

	msgType, msg, err = websocketServe.ReadMessage()
	//fmt.Println("websocketServe2", msgType, msg, err, string(msg))
	if nil != err {
		err = fmt.Errorf("read websocket rfb message err: %+v", err)
		return
	}
	//"RFB 003.008\n"
	if err = pveVncConn.WriteMessage(msgType, msg); err != nil {
		err = fmt.Errorf("write pve rfb message err: %+v", err)
		return
	}
	msgType, msg, err = pveVncConn.ReadMessage()
	//fmt.Println("pveVncConn3", msgType, msg, err, string(msg))
	if nil != err {
		err = fmt.Errorf("read websocket auth type message err: %+v", err)
		return
	}
	//[]uint8{1,2}  type 2 is need password
	if err = pveVncConn.WriteMessage(websocket.BinaryMessage, []uint8{2}); err != nil {
		err = fmt.Errorf("write pve auth type message err: %+v", err)
		return
	}

	msgType, msg, err = pveVncConn.ReadMessage()
	//fmt.Println("pveVncConn4", msgType, msg, err, string(msg))
	if nil != err {
		err = fmt.Errorf("read pve auth random key message err: %+v", err)
		return
	}
	//[]unit8{...}  len 16
	enPassword, err := VNCAuthPasswordEncrypt(vnc.Ticket, msg)
	//fmt.Println("enPassword", enPassword, err, string(enPassword))
	if err = pveVncConn.WriteMessage(websocket.BinaryMessage, enPassword); err != nil {
		err = fmt.Errorf("write pve auth password message err: %+v", err)
		return
	}
	//msgType, msg, err = pveVncConn.ReadMessage()
	//fmt.Println("pveVncConn5", msgType, msg, err, string(msg))

	//send websocket do not need password
	if err = websocketServe.WriteMessage(websocket.BinaryMessage, []uint8{1, 1}); err != nil {
		err = fmt.Errorf("write websocket auth type message err: %+v", err)
		return
	}
	msgType, msg, err = websocketServe.ReadMessage()
	//fmt.Println("websocketServe6", msgType, msg, err, string(msg))
	if nil != err {
		err = fmt.Errorf("read websocket auth type return message err: %+v", err)
		return
	}

	go func() {

		for {
			_, err = bufCopy.Copy(websocketServe.NetConn(), pveVncConn.NetConn())
			if err != nil {
				err = fmt.Errorf("buf copy pve to websocket err: %+v", err)
				pveVncConn.Close()
				websocketServe.Close()
				return
			}
		}
	}()

	for {
		_, err = bufCopy.Copy(pveVncConn.UnderlyingConn(), websocketServe.UnderlyingConn())
		if err != nil {
			err = fmt.Errorf("buf copy websocket to pve err: %+v", err)
			pveVncConn.Close()
			websocketServe.Close()
			return
		}
	}

	return
}

func VNCAuthPasswordEncrypt(key string, bytes []byte) ([]byte, error) {
	keyBytes := []byte{0, 0, 0, 0, 0, 0, 0, 0}

	if len(key) > 8 {
		key = key[:8]
	}

	for i := 0; i < len(key); i++ {
		keyBytes[i] = utils.ReverseBits(key[i])
	}

	block, err := des.NewCipher(keyBytes)

	if err != nil {
		return nil, err
	}

	result1 := make([]byte, 8)
	block.Encrypt(result1, bytes)
	result2 := make([]byte, 8)
	block.Encrypt(result2, bytes[8:])

	crypted := append(result1, result2...)

	return crypted, nil
}
