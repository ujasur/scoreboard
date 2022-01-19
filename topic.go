package main

import "github.com/gorilla/websocket"

type msgWriter interface {
	write(c *websocket.Conn, p *principal) error
}

type client struct {
	id  string
	msg chan msgWriter
}

type topic interface {
	enter(c *client)
	leave(c *client)
	sync() msgWriter
}
