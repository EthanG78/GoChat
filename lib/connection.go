package lib

import (
	"log"
	"sync"
	"github.com/gorilla/websocket"
	"net/http"
	"net"

	"github.com/labstack/echo"
)

//////////////////////
//CONNECTIONS
/////////////////////

type connection struct {
	// Buffered channel of outbound messages.
	send chan []byte

	// The hub.
	h *Hub
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

//Main handler
type WsHandler struct {
	H *Hub
}

func Chat(c echo.Context)error {

	H := NewHub()
	wsHandler := WsHandler{H:H}

	//Find users IP and display them
	ip,_,_ := net.SplitHostPort(c.Request().RemoteAddr)
	log.Printf("%s has connected", ip)


	wsConn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		log.Printf("error upgrading %s", err)
		return c.String(http.StatusInternalServerError, "Error upgrading web socket")
	}
	con := &connection{send: make(chan []byte, 256), h: wsHandler.H}
	con.h.addConnection(con)
	defer con.h.removeConnection(con)
	var wg sync.WaitGroup
	wg.Add(2)
	go con.writer(&wg, wsConn)
	go con.reader(&wg, wsConn)
	wg.Wait()
	wsConn.Close()

	return c.String(http.StatusOK, "")
}


/*func (wsh WsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	//Find users IP and display them
	ip,_,_ := net.SplitHostPort(r.RemoteAddr)
	log.Printf("%s has connected", ip)


	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("error upgrading %s", err)
		return
	}
	c := &connection{send: make(chan []byte, 256), h: wsh.H}
	c.h.addConnection(c)
	defer c.h.removeConnection(c)
	var wg sync.WaitGroup
	wg.Add(2)
	go c.writer(&wg, wsConn)
	go c.reader(&wg, wsConn)
	wg.Wait()
	wsConn.Close()
}*/
