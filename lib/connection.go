package lib

import (
	"log"
	"sync"
	"github.com/gorilla/websocket"
	"net/http"
)

//////////////////////
//CONNECTIONS
/////////////////////

type connection struct {
	// Buffered channel of outbound messages.
	send chan []byte

	// The hub.
	h *hub
}

func (c *connection) reader(wg *sync.WaitGroup, wsConn *websocket.Conn) {
	defer wg.Done()
	for {
		_, message, err := wsConn.ReadMessage()
		if err != nil {
			break
		}
		c.h.broadcast <- message
	}
}

func (c *connection) writer(wg *sync.WaitGroup, wsConn *websocket.Conn) {
	defer wg.Done()
	for message := range c.send {
		err := wsConn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			break
		}
	}
}

var upgrader = &websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}

type wsHandler struct {
	h *hub
}

func (wsh wsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	/*//BAN USERS HERE!
	//Update ip's with syntax "HOST:PORT"
	var bannedIp []string
	bannedIp = append(bannedIp, "")
	/////////////
	//Finds user IP
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	log.Printf("%s has connected.", ip)

	//Checks to see if user IP matches slice of banned IP's
	for i := range bannedIp {
		if bannedIp[i] == ip {
			log.Printf("%s has been banned from the web server", ip)
			http.Error(w, "BANNED", http.StatusUnauthorized)
			return
		}
	}
	*/

	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("error upgrading %s", err)
		return
	}
	c := &connection{send: make(chan []byte, 256), h: wsh.h}
	c.h.addConnection(c)
	defer c.h.removeConnection(c)
	var wg sync.WaitGroup
	wg.Add(2)
	go c.writer(&wg, wsConn)
	go c.reader(&wg, wsConn)
	wg.Wait()
	wsConn.Close()
}
