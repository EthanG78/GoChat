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
	"github.com/satori/go.uuid"
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



	//BAN USERS HERE!
	//Update ip's with syntax "HOST:PORT"
	var bannedIp []string
	bannedIp = append(bannedIp, "")
	/////////////
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
	//////////

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

type user struct {
	UserName string
	Pass string

}

var dbUsers = map[string]user{}
var dbSessions = map[string]string{}
var tpl *template.Template

func init() {
	tpl = template.Must(template.ParseGlob("templates/*"))
	dbUsers["Test"] = user{"Test", "eth787878"}
}


//TODO: Finish this stupid function
func signup(w http.ResponseWriter, req *http.Request)  {
	c, err := req.Cookie("session")
	if err != nil{
		sID := uuid.NewV4()
		c = &http.Cookie{
			Name: "session",
			Value: sID.String(),
		}
		http.SetCookie(w, c)
	}

	//Check form submission
	var u user
	if req.Method == http.MethodPost{
		un := req.FormValue("username")
		p := req.FormValue("password")

		u = user{un, p}

		dbUsers[c.Value] = u
	}

	//Executes Template
	tpl.ExecuteTemplate(w, "signup.gohtml", nil)
}

func login(w http.ResponseWriter, req *http.Request)  {
	if req.Method == http.MethodPost{
		un := req.FormValue("username")
		p := req.FormValue("password")
		//Does this user exist?? Using comma ok idiom
		u, ok := dbUsers[un]
		if !ok{
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		//does the username/password combo match at all??
		if u.Pass != p{
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		//Create a session
		sID := uuid.NewV4()
		c := &http.Cookie{
			Name: "session",
			Value: sID.String(),
			HttpOnly: true,
		}
		http.SetCookie(w, c)
		dbSessions[c.Value] = un
		http.Redirect(w, req, "/chat", http.StatusSeeOther)
		return
	}

	tpl.ExecuteTemplate(w, "login.gohtml", nil)
}



func homeHandler(tpl *template.Template) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tpl.Execute(w, r)
	})
}

func main() {
	flag.Parse()
	tpl := template.Must(template.ParseFiles("templates/chat.html"))
	h := newHub()
	router := http.NewServeMux()
	router.HandleFunc("/signup", signup)
	router.HandleFunc("/", login)
	router.Handle("/chat", homeHandler(tpl))
	router.Handle("/ws", wsHandler{h: h} )
	log.Println("serving on port 8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}

//TODO: Here is a comment, current build is not user friendly!!
//TODO: Build a home function where users can be redirected to and from login, signup and the chat
//TODO: Add redirecting links to gohtml files
//TODO: Make chat.html into "gohtml"
