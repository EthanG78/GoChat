package main

import (
	"flag"
	"log"
	"net/http"
	"text/template"
	"golang.org/x/crypto/bcrypt"
	"github.com/EthanG78/golang_chat/lib"
	"github.com/satori/go.uuid"

	"time"
)


//////////////////////
//MAIN
/////////////////////

type user struct {
	UserName string
	Pass     string
}

var dbUsers = map[string]user{}
var dbSessions = map[string]string{}
var tpl *template.Template

func init() {
	tpl = template.Must(template.ParseGlob("templates/*"))
	dbUsers["Test"] = user{"Test", "eth787878"}
}

func homeHandler(tpl *template.Template) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tpl.Execute(w, r)
	})
}

func forbidden(w http.ResponseWriter, req *http.Request)  {

	/*
	http.Error(w, "Please fill out the required fields, you will be redirected shortly", http.StatusForbidden)
	*/
	tpl.ExecuteTemplate(w, "forbidden.gohtml", nil)
}



func sign_up(w http.ResponseWriter, req *http.Request) {
	c, err := req.Cookie("session")
	if err != nil {
		sID := uuid.NewV4()
		c = &http.Cookie{
			Name:  "session",
			Value: sID.String(),
		}
		http.SetCookie(w, c)
	}

	//Check form submission
	var u user
	if req.Method == http.MethodPost {
		un := req.FormValue("username")
		p := req.FormValue("password")


		//Checking to see if user filled out required fields.
		if un == ""{
			time.Sleep(3000)
			http.Redirect(w, req, "/forbidden", http.StatusSeeOther)
			return

		}else if p == "" {
			time.Sleep(3000)
			http.Redirect(w, req, "/forbidden", http.StatusSeeOther)
			return
		}
		//Must declare password as a byte after error checking
		pss := []byte(p)
		password, err := bcrypt.GenerateFromPassword(pss,0)
		if err != nil{
			log.Fatalf("Error logging password for %s", un)
		}

		pass := string(password[:])
		c.Value = un
		u = user{un, pass}

		dbUsers[c.Value] = u
		http.Redirect(w, req, "/login", http.StatusSeeOther)

		log.Println(dbUsers)
		return
	}

	//Executes Template
	tpl.ExecuteTemplate(w, "signup.gohtml", nil)
}

func login(w http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodPost {
		un := req.FormValue("username")
		p := req.FormValue("password")

		if un == ""{
			time.Sleep(3000)
			http.Redirect(w, req, "/forbidden", http.StatusSeeOther)
			return

		}else if p == ""{
			time.Sleep(3000)
			http.Redirect(w, req, "/forbidden", http.StatusSeeOther)
			return
		}
		//Does this user exist?? Using comma ok idiom
		u, ok:= dbUsers[un]
		if !ok {
			time.Sleep(3000)
			http.Redirect(w, req, "/forbidden", http.StatusSeeOther)
			return
		}
		//does the username/password combo match at all??
		//Compares bcrypt hash to user input!
		pass := []byte(p)
		userPass := []byte(u.Pass)
		password := bcrypt.CompareHashAndPassword(userPass, pass)
		if password != nil{
			time.Sleep(3000)
			http.Redirect(w, req, "/forbidden", http.StatusSeeOther)
			return
		}

		/*if u.Pass != p {
			time.Sleep(3000)
			http.Redirect(w, req, "/forbidden", http.StatusSeeOther)
			return
		}*/

		//Create a session
		sID := uuid.NewV4()
		c := &http.Cookie{
			Name:     "session",
			Value:    sID.String(),
			HttpOnly: true,
		}
		http.SetCookie(w, c)
		dbSessions[c.Value] = un
		http.Redirect(w, req, "/chat", http.StatusSeeOther)
		return
	}

	tpl.ExecuteTemplate(w, "login.gohtml", nil)
}



func main() {

	flag.Parse()
	tpl := template.Must(template.ParseFiles("templates/chat.gohtml"))
	H := lib.NewHub()
	router := http.NewServeMux()
	router.HandleFunc("/", sign_up)
	router.HandleFunc("/login", login)
	router.HandleFunc("/forbidden", forbidden)
	router.Handle("/chat", homeHandler(tpl))
	router.Handle("/ws", lib.WsHandler{H:H})
	log.Println("serving on port 8080")
	log.Println("Users:", dbUsers)
	//log.Println("Sessions: ", dbSessions)
	log.Fatal(http.ListenAndServe(":8080", router))
}

//TODO: Here is a comment, current build is not user friendly!!
//TODO: Build a home function where users can be redirected to and from login, signup and the chat
//TODO: Add redirecting links to go html files
//TODO: Make chat.html into "go html"
