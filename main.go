package main

import (
	"flag"
	"log"
	"net/http"
	"text/template"
	"sync"
	"time"
	"net"

	"github.com/gorilla/websocket"


)

//////////////////////
//HUB
/////////////////////
type hub struct {
	// the mutex to protect connections
	connectionsMx sync.RWMutex

	// Registered connections.
	connections map[*connection]struct{}

	// Inbound messages from the connections.
	broadcast chan []byte

	logMx sync.RWMutex
	log   [][]byte
}

func newHub() *hub {
	h := &hub{
		connectionsMx: sync.RWMutex{},
		broadcast:     make(chan []byte),
		connections:   make(map[*connection]struct{}),
	}

	go func() {
		for {
			msg := <-h.broadcast
			h.connectionsMx.RLock()
			for c := range h.connections {
				select {
				case c.send <- msg:
				// stop trying to send to this connection after trying for 1 second.
				// if we have to stop, it means that a reader died so remove the connection also.
				case <-time.After(1 * time.Second):
					log.Printf("shutting down connection %s", c)
					h.removeConnection(c)
				}
			}
			h.connectionsMx.RUnlock()
		}
	}()
	return h
}

func (h *hub) addConnection(conn *connection) {
	h.connectionsMx.Lock()
	defer h.connectionsMx.Unlock()
	h.connections[conn] = struct{}{}
}

func (h *hub) removeConnection(conn *connection) {
	h.connectionsMx.Lock()
	defer h.connectionsMx.Unlock()
	if _, ok := h.connections[conn]; ok {
		delete(h.connections, conn)
		close(conn.send)
	}
}

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

	//Update ip's with syntax "HOST:PORT"
	var bannedIp []string
	bannedIp = append(bannedIp, "")

	//Finds user IP
	ip,_,_ := net.SplitHostPort(r.RemoteAddr)
	log.Printf("%s has connected.", ip)

	//Checks to see if user IP matches slice of banned IP's
	for i := range bannedIp{
		if bannedIp[i] == ip{
			log.Printf("%s has been banned from the web server", ip)
			http.Error(w, "BANNED", http.StatusUnauthorized)
			return
		}
	}


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


//////////////////////
//MAIN
/////////////////////


/*//Must use this syntax for adding banned ip's "host:port"
var bannedIp []string

func banUser (res http.ResponseWriter, req *http.Request){
	//Grabs clients IP address
	ip,_,_ := net.SplitHostPort(req.RemoteAddr)

	for i := range bannedIp{
		if i == ip {
			log.Printf("%s has been banned from the web server", ip)
			log.Fatal()

		}
	}


}*/

func homeHandler(tpl *template.Template) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tpl.Execute(w, r)
	})
}

func main() {
	flag.Parse()
	tpl := template.Must(template.ParseFiles("home.html"))
	h := newHub()
	router := http.NewServeMux()
	router.Handle("/", homeHandler(tpl))
	router.Handle("/ws", wsHandler{h: h} )
	log.Println("serving on port 8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
