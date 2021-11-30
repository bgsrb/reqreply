package main

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Request struct {
	ID      uuid.UUID   `json:"id"`
	Message interface{} `json:"message"`
	Done    chan bool
	Error   error
	Res     Response
}

type Response struct {
	ID      uuid.UUID   `json:"id"`
	Message interface{} `json:"message"`
}

type Client struct {
	mutex sync.Mutex
	reqs  map[uuid.UUID]*Request
}

func New() *Client {
	c := &Client{
		reqs: make(map[uuid.UUID]*Request, 1),
	}
	go c.read()
	return c
}

var request = make(chan *Request)
var response = make(chan Response)

func (c *Client) read() {
	var err error

	for err == nil {
		res := <-response

		if err != nil {
			err = fmt.Errorf("error reading message: %q", err)
			continue
		}
		req := c.reqs[res.ID]
		if req == nil {
			err = errors.New("no pending request found")
			continue
		}
		req.Done <- true
		req.Res = res
	}
	for _, req := range c.reqs {
		req.Error = err
		req.Done <- true
	}
}

func (c *Client) Request(payload interface{}) (interface{}, error) {
	id := uuid.New()
	req := &Request{ID: id, Message: payload, Done: make(chan bool)}

	c.mutex.Lock()
	c.reqs[id] = req
	c.mutex.Unlock()

	request <- req

	select {
	case <-req.Done:
	case <-time.After(20 * time.Second):
		req.Error = errors.New("request timeout")
	}

	delete(c.reqs, req.ID)

	if req.Error != nil {
		return nil, req.Error
	}
	return req.Res.Message, nil
}

func main() {
	client := New()

	go func() {
		res, _ := client.Request("Hi")
		fmt.Println("Hi", fmt.Sprintf("%+v", res))
	}()

	go func() {
		res, _ := client.Request("Hello")
		fmt.Println("Hello", fmt.Sprintf("%+v", res))
	}()

	go func() {
		for {
			req := <-request
			response <- Response{
				ID:      req.ID,
				Message: req.Message,
			}
		}
	}()

	<-make(chan int)
}
