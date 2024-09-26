package pve

import (
	"encoding/json"
	"fmt"
	"github.com/jinzhu/copier"
	"strings"
	"time"
)

const (
	TaskRunning = "running"
)

type Tasks []*Tasks
type Task struct {
	client       *Client
	UPID         string
	ID           string
	Type         string
	User         string
	Status       string
	Node         string
	PID          uint64 `json:",omitempty"`
	PStart       uint64 `json:",omitempty"`
	Saved        string `json:",omitempty"`
	ExitStatus   string `json:",omitempty"`
	IsCompleted  bool
	IsRunning    bool
	IsFailed     bool
	IsSuccessful bool
	StartTime    time.Time     `json:"-"`
	EndTime      time.Time     `json:"-"`
	Duration     time.Duration `json:"-"`
}

type TaskLog map[int]string

// line numbers in the response start a 1  but the start param indexes from 0 so converting to that
func (l *TaskLog) UnmarshalJSON(b []byte) error {
	var data []map[string]interface{}
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}

	log := make(map[int]string, len(data))
	for _, row := range data {
		if n, ok := row["n"]; ok {
			if t, ok := row["t"]; ok {
				log[int(n.(float64))-1] = t.(string)
			}
		}
	}

	return copier.Copy(l, TaskLog(log))
}

func (t *Task) UnmarshalJSON(b []byte) error {
	var tmp map[string]interface{}
	if err := json.Unmarshal(b, &tmp); err != nil {
		return err
	}

	type TempTask Task
	var task TempTask
	if err := json.Unmarshal(b, &task); err != nil {
		return err
	}

	if starttime, ok := tmp["starttime"]; ok {
		task.StartTime = time.Unix(int64(starttime.(float64)), 0)
	}

	if endtime, ok := tmp["endtime"]; ok {
		task.EndTime = time.Unix(int64(endtime.(float64)), 0)
	}

	if !task.StartTime.IsZero() && !task.EndTime.IsZero() {
		task.Duration = task.EndTime.Sub(task.StartTime)
	}

	c := Task(task)
	return copier.Copy(t, &c)
}

var DefaultWaitInterval = 1 * time.Second

func NewTask(upid string, client *Client) *Task {
	if upid == "" {
		return nil
	}

	task := &Task{
		UPID:   upid,
		client: client,
	}

	sp := strings.Split(string(task.UPID), ":")
	if len(sp) == 0 || len(sp) < 7 {
		return task
	}

	task.Node = sp[1]
	task.Type = sp[5]
	task.ID = sp[6]
	task.User = sp[7]

	return task
}

func (t *Task) Ping() error {
	tmp := NewTask(t.UPID, t.client)
	err := t.client.Get(fmt.Sprintf("/nodes/%s/tasks/%s/status", t.Node, t.UPID), t)
	if nil != err || nil == t {
		t = tmp
	}
	if nil == t.client {
		t.client = tmp.client
	}

	if "stopped" == t.Status {
		t.IsCompleted = true
	} else {
		t.IsRunning = true
	}
	if t.IsCompleted {
		if "OK" == t.ExitStatus {
			t.IsSuccessful = true
		} else {
			t.IsFailed = true
		}
	}
	return err
}

func (t *Task) Stop() error {
	return t.client.Delete(fmt.Sprintf("/nodes/%s/tasks/%s", t.Node, t.UPID), nil)
}

func (t *Task) Log(start, limit int) (l TaskLog, err error) {
	return l, t.client.Get(fmt.Sprintf("/nodes/%s/tasks/%s/log?start=%d&limit=%d", t.Node, t.UPID, start, limit), &l)
}
func (t *Task) LogTail(start int, watch chan string) error {
	for {
		t.client.logger.DebugF("tailing log for task %s", t.UPID)
		if err := t.Ping(); err != nil {
			return err
		}

		if t.Status != TaskRunning {
			t.client.logger.DebugF("task %s is no longer running, closing down watcher", t.UPID)
			close(watch)
			return nil
		}

		logs, err := t.Log(start, 50)
		if err != nil {
			return err
		}
		for _, ln := range logs {
			watch <- ln
		}
		start = start + len(logs)
		time.Sleep(2 * time.Second)
	}
}

func (t *Task) Watch(start int) (chan string, error) {
	t.client.logger.DebugF("starting watcher on %s", t.UPID)
	watch := make(chan string)

	log, err := t.Log(start, 50)
	if err != nil {
		return watch, err
	}

	for i := 0; i < 3; i++ {
		// retry 3 times if the log has no entries
		t.client.logger.DebugF("no logs for %s found, retrying %d of 3 times", t.UPID, i)
		if len(log) > 0 {
			break
		}
		time.Sleep(1 * time.Second)

		log, err = t.Log(start, 50)
		if err != nil {
			return watch, err
		}
	}

	if len(log) == 0 {
		return watch, fmt.Errorf("no logs available for %s", t.UPID)
	}

	go func() {
		t.client.logger.DebugF("logs found for task %s", t.UPID)
		for _, ln := range log {
			watch <- ln
		}
		t.client.logger.DebugF("watching task %s", t.UPID)
		err := t.LogTail(len(log), watch)
		if err != nil {
			t.client.logger.ErrorF("error watching logs: %s", err)
		}
	}()

	t.client.logger.DebugF("returning watcher for %s", t.UPID)
	return watch, nil
}

func (t *Task) WaitFor(seconds int) error {
	return t.Wait(DefaultWaitInterval, time.Duration(seconds)*time.Second)
}

func (t *Task) Wait(interval, max time.Duration) error {
	// ping it quick to fill in all the details we need in case they're not there
	t.Ping()
	t.client.logger.DebugF("waiting for %s, checking every %fs for %fs", t.UPID, interval.Seconds(), max.Seconds())

	timeout := time.After(max)
	for {
		select {
		case <-timeout:
			t.client.logger.DebugF("timed out waiting for task %s for %fs", t.UPID, max.Seconds())
			return ErrTimeout
		default:
			if err := t.Ping(); err != nil {
				return err
			}

			if t.Status != TaskRunning {
				t.client.logger.DebugF("task %s has completed with status %s", t.UPID, t.Status)
				return nil
			}
			t.client.logger.DebugF("waiting on task %s sleeping for %fs", t.UPID, interval.Seconds())
		}
		time.Sleep(interval)
	}
}

func (t *Task) WaitForCompleteStatus(timesNum int, stepSeconds ...int) (status bool, completed bool, err error) {
	step := 1
	if len(stepSeconds) > 0 && stepSeconds[0] > 1 {
		step = stepSeconds[0]
	}
	var timeout <-chan time.Time

	for {
		if timesNum > 0 {
			select {
			case <-timeout:
				return

			default:

			}
		}

		err = t.Ping()
		if nil != err {
			t.client.logger.DebugF("task %s ping error %+v", t.UPID, err)
			break
		}
		completed = t.IsCompleted

		if completed {
			status = t.IsSuccessful
			if !status {
				err = fmt.Errorf(t.ExitStatus)
			}
			return
		}

		time.Sleep(time.Duration(step) * time.Second)

	}
	return
}
func (t *Task) WaitForComplete(timesNum int, stepSeconds ...int) (err error) {

	_, completed, err := t.WaitForCompleteStatus(timesNum, stepSeconds...)
	if nil != err {
		return
	}
	if !completed {
		err = fmt.Errorf("task not completed")
		return
	}

	return
}
