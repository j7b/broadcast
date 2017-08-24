package main

import (
	"log"
	"net"
	"net/textproto"
	"strings"

	"github.com/j7b/broadcast"
)

var ss = broadcast.NewShared()

type String string

func (String) Broadcast() {}

func serve(c net.Conn, r broadcast.Receiver) {
	tp := textproto.NewConn(c)
	lines := make(chan String)
	go func() {
		defer close(lines)
		for {
			s, err := tp.ReadLine()
			if err != nil {
				return
			}
			lines <- String(c.RemoteAddr().String() + ` sez: ` + s)
		}
	}()
	ss.Send(String(`New connection from: ` + c.RemoteAddr().String()))
	defer ss.Send(String(`Lost connection from: ` + c.RemoteAddr().String()))
	for {
		select {
		case <-r.Chan():
			b := r.Receive()
			if strings.Contains(string(b.(String)), c.RemoteAddr().String()) {
				continue
			}
			if err := tp.PrintfLine("%s", b); err != nil {
				return
			}
		case l, ok := <-lines:
			if !ok {
				return
			}
			ss.Send(l)
		}
	}
}

func main() {
	l, err := net.Listen("tcp4", "127.0.0.1:6666")
	if err != nil {
		log.Fatal(err)
	}
	for {
		c, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go serve(c, ss.Receiver())
	}
}
