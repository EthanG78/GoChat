package main

import (
	"crypto/md5"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"text/template"
	"time"

	"github.com/EthanG78/golang_chat/lib"
	"github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

//////////////////////
//MAIN
/////////////////////

//Create a user
type user struct {
	UserName string
	Pass     string
}

//Establish variables
var dbUsers = map[string]user{}
var dbSessions = map[string]string{}
var tpl *template.Template

//Initialize template reader
func init() {
	tpl = template.Must(template.ParseGlob("templates/*"))
	dbUsers["Test"] = user{"Test", "eth787878"}
}

func homeHandler(tpl *template.Template) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tpl.Execute(w, r)
	})
}

//Redirects to error page
func forbidden(w http.ResponseWriter, req *http.Request) {
	tpl.ExecuteTemplate(w, "forbidden.gohtml", nil)
}

//Sign up function
func signUp(w http.ResponseWriter, req *http.Request) {
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
		if un == "" {
			time.Sleep(3000)
			http.Redirect(w, req, "/forbidden", http.StatusSeeOther)
			return

		} else if p == "" {
			time.Sleep(3000)
			http.Redirect(w, req, "/forbidden", http.StatusSeeOther)
			return
		}
		//Must declare password as a byte after error checking
		pss := []byte(p)
		password, err := bcrypt.GenerateFromPassword(pss, 0)
		if err != nil {
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

	tpl.ExecuteTemplate(w, "signup.gohtml", nil)
}

//Login function
func login(w http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodPost {
		un := req.FormValue("username")
		pass := req.FormValue("password")

		if un == "" {
			time.Sleep(3000)
			http.Redirect(w, req, "/forbidden", http.StatusSeeOther)
			return

		} else if pass == "" {
			time.Sleep(3000)
			http.Redirect(w, req, "/forbidden", http.StatusSeeOther)
			return
		}
		//Does this user exist?? Using comma ok idiom
		u, ok := dbUsers[un]
		if !ok {
			time.Sleep(3000)
			http.Redirect(w, req, "/forbidden", http.StatusSeeOther)
			return
		}
		//does the username/password combo match at all??
		//Compares bcrypt hash to user input!

		password := []byte(pass)
		hash := []byte(u.Pass)
		err := bcrypt.CompareHashAndPassword(hash, password)
		if err != nil {
			time.Sleep(3000)
			http.Redirect(w, req, "/forbidden", http.StatusSeeOther)
			return
		}

		//Create a session
		sID := uuid.NewV4()
		c := &http.Cookie{
			Name:     "session",
			Value:    sID.String(),
			HttpOnly: true,
		}

		http.SetCookie(w, c)
		dbSessions[c.Value] = un

		//Genertaing random token for validations
		h := md5.New()
		crutime := int64(-42)
		io.WriteString(h, strconv.FormatInt(crutime, 10))
		io.WriteString(h, "ganraomaxxxxxxxxx")
		Token := fmt.Sprintf("%x", h.Sum(nil))

		//Showing token for debugging
		log.Println(un, ":", Token)

		req.ParseForm()
		Token = req.Form.Get("token")
		if Token != "" {
			http.Redirect(w, req, "/chat", http.StatusSeeOther)
		} else {
			http.Error(w, "Error validating login token.", http.StatusForbidden)
		}
		return
	}

	tpl.ExecuteTemplate(w, "login.gohtml", nil)

}

//Handles site favicon
func faviconHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "styling/favicon.ico")
}

//Handles lassajous animated gif
func lassajousHandler(w http.ResponseWriter, r *http.Request) {
	lib.Lassajous(w)
}

//Handles the home page
func home(w http.ResponseWriter, r *http.Request) {
	tpl.ExecuteTemplate(w, "home.gohtml", nil)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		home(w, r)
	} else {
		log.Printf("rootHandler: Could not forward request for %s any further.", r.RequestURI)
	}
}

//MAIN
func main() {

	//Wrapping handlers

	flag.Parse()
	tpl := template.Must(template.ParseFiles("templates/chat.gohtml"))
	H := lib.NewHub()
	router := http.NewServeMux()
	router.Handle("/styling/", http.StripPrefix("/styling/", http.FileServer(http.Dir("styling/"))))
	router.HandleFunc("/favicon.ico", faviconHandler)
	router.HandleFunc("/", rootHandler)
	router.HandleFunc("/signup", signUp)
	router.HandleFunc("/login", login)
	router.HandleFunc("/forbidden", forbidden)
	router.HandleFunc("/lassajous", lassajousHandler)
	router.Handle("/chat", homeHandler(tpl))
	router.Handle("/ws", lib.WsHandler{H: H})
	log.Println("serving on port 8080")
	log.Println("Users:", dbUsers)
	//log.Println("Sessions: ", dbSessions)
	log.Fatal(http.ListenAndServe(":8080", router))
}

//TODO: Here is a comment, current build is not user friendly!!
//TODO: Build a home function where users can be redirected to and from login, signup and the chat
//TODO: Add redirecting links to go html files
