package main

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type online struct {
	clients  map[*client]bool
	entering chan *client
	leaving  chan *client
	sched    *time.Ticker
	refs     map[string]int
	dirty    bool
	mux      sync.RWMutex
}

func newOnline() *online {
	o := new(online)
	o.clients = make(map[*client]bool)
	o.entering = make(chan *client)
	o.leaving = make(chan *client)
	o.sched = time.NewTicker(time.Millisecond * 200)
	o.refs = make(map[string]int)
	go o.start()
	return o
}

func (o *online) enter(s *client) {
	o.entering <- s
}

func (o *online) leave(s *client) {
	o.leaving <- s
}

func (o *online) sync() msgWriter {
	o.mux.RLock()
	defer o.mux.RUnlock()
	msg := new(onlineUsersMsg)
	msg.users = make([]string, 0, len(o.refs))
	for id := range o.refs {
		msg.users = append(msg.users, id)
	}
	return msg
}

func (o *online) start() {
	defer o.sched.Stop()
	for {
		select {
		case <-o.sched.C:
			if o.dirty {
				o.dirty = false

				msg := o.sync()
				for s := range o.clients {
					s.msg <- msg
				}
			}
		case s := <-o.entering:
			o.clients[s] = true
			o.refs[s.id]++
			if o.refs[s.id] == 1 {
				o.dirty = true
			}
		case s := <-o.leaving:
			delete(o.clients, s)
			close(s.msg)

			o.refs[s.id]--
			if o.refs[s.id] == 0 {
				delete(o.refs, s.id)
				o.dirty = true
			}
		}
	}
}

type onlineUsersMsg struct {
	users []string
}

func (o *onlineUsersMsg) write(c *websocket.Conn, p *principal) error {
	return c.WriteJSON(o.users)
}
