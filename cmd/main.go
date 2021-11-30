package main

import (
	"fmt"

	"reqreply/client"
)

func main() {
	c := client.New()

	go func() {
		res, _ := c.Request("Hi")
		fmt.Println("Hi", fmt.Sprintf("%+v", res))
	}()

	go func() {
		res, _ := c.Request("Hello")
		fmt.Println("Hello", fmt.Sprintf("%+v", res))
	}()

	go c.Respond(func(req *client.Request) client.Response {
		return client.Response{
			ID:      req.ID,
			Message: req.Message,
		}
	})

	<-make(chan int)
}
