package controller

import (
	"log"
	"sync"
	"time"
)

type hub struct {
	//mutex that protects connections
	connectionMx sync.RWMutex

	//Registered connections
	connections map[*connection]struct{}

	//Messages from connections.
	broadcast chan []byte

	logMx sync.RWMutex
	log	[][]byte
}

func newHub() *hub{
	h := &hub{
		connectionsMx: sync.RWMutex{},
		broadcast: make(chan []byte),
		connections: make(map[*connection]struct{}),
	}

	go func() {
		for {
			msg := <-h.broadcast
			h.connectionsMx.RLock()
			for c := range h.connections {
				select {
				case c.send <- msg:
					//stop trying to send this connection after trying for 1 second
					//if we have tos top, it means that a reader dies to remove the connection also
				case <- time.After(1 * time.Second):
					log.Printf("shutting down connection %s", c)
					h.removeConnections(c)
				}
			}
			h.connectionMx.RUnlock()
		}
	}()
	return h
}
