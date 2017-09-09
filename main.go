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
	"github.com/labstack/echo"
)

//User type referenced in DB
type user struct {
	Username 		string		`json:"username"`
	Pass     		string		`json:"pass"`
}

type TemplateRenderer struct{
	template *template.Template
}

//Database and template variables
var dbUsers = map[string]user{}
var dbSessions = map[string]string{}
var tpl *template.Template


///
//Initialize template reader as well as superuser
///
func init() {
	tpl = template.Must(template.ParseGlob("static/*"))
	//ADMIN USER: Admin:gochatadmin
	dbUsers["Admin"] = user{"Admin", "$2a$10$5xymUNPSZfAm.XztfVCqUuC3MYLTPJ.dbXhGFsAJGaqyXoHteR8TO"}
}

func (t* TemplateRenderer) Renderer (w io.Writer, name string, data interface{}, c echo.Context) error{
	if viewContext, isMap := data.(map[string]interface{}); isMap{
		viewContext["reverse"] = c.Echo().Reverse
	}
	return t.template.ExecuteTemplate(w, name, data)
}

func homeHandler(tpl *template.Template) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tpl.Execute(w, r)
	})
}

/////////////////////////
//Redirects to error page
/////////////////////////
func forbidden(w http.ResponseWriter, req *http.Request) {
	tpl.ExecuteTemplate(w, "forbidden.gohtml", nil)
}

//////////////////
//Sign up function
//////////////////
func signUp(w http.ResponseWriter, req *http.Request) {
	//Create cookie session
	c, err := req.Cookie("session")
	if err != nil {
		sID := uuid.NewV4()
		c = &http.Cookie{
			Name:  "session",
			Value: sID.String(),
		}
		http.SetCookie(w, c)
	}

	//Checks if the user login is correct
	var u user
	if req.Method == http.MethodPost {
		un := req.FormValue("username")
		p := req.FormValue("password")

		//If the user does not submit anything, they will be redirected
		if un == "" {
			time.Sleep(3000)
			http.Redirect(w, req, "/forbidden", http.StatusSeeOther)
			return

		} else if p == "" {
			time.Sleep(3000)
			http.Redirect(w, req, "/forbidden", http.StatusSeeOther)
			return
		}

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

////////////////
//Login function
////////////////
func login(w http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodPost {
		un := req.FormValue("username")
		pass := req.FormValue("password")

		//If the user does not submit anything, they will be redirected
		if un == "" {
			time.Sleep(3000)
			http.Redirect(w, req, "/forbidden", http.StatusSeeOther)
			return

		} else if pass == "" {
			time.Sleep(3000)
			http.Redirect(w, req, "/forbidden", http.StatusSeeOther)
			return
		}

		//Checking to see if this user does in fact exist within the DataBase
		u, ok := dbUsers[un]
		if !ok {
			time.Sleep(3000)
			http.Redirect(w, req, "/forbidden", http.StatusSeeOther)
			return
		}
		//does the username/password combo match at all?
		//Compares bcrypt hash to user input

		pass2check := []byte(pass)
		hash := []byte(u.Pass)
		err := bcrypt.CompareHashAndPassword(hash, pass2check)
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

		//Generating random token for validation
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

//////////////////////
//Handles site favicon
//////////////////////
func faviconHandler(w http.ResponseWriter, req *http.Request) {
	http.ServeFile(w, req, "styling/favicon.ico")
}

///////////////////////
//Handles the home page
///////////////////////
func home(w http.ResponseWriter, req *http.Request) {
	tpl.ExecuteTemplate(w, "home.gohtml", nil)
}

func rootHandler(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path == "/" {
		home(w, req)
	} else {
		log.Printf("rootHandler: Could not forward request for %s any further.", req.RequestURI)
	}
}

//////
//MAIN
//////
func main() {

	//Wrapping handlers
	flag.Parse()
	tpl := template.Must(template.ParseFiles("static/chat.gohtml"))
	H := lib.NewHub()
	router := http.NewServeMux()
	router.Handle("/styling/", http.StripPrefix("/styling/", http.FileServer(http.Dir("styling/"))))
	router.HandleFunc("/favicon.ico", faviconHandler)
	router.HandleFunc("/", rootHandler)
	router.HandleFunc("/signup", signUp)
	router.HandleFunc("/login", login)
	router.HandleFunc("/forbidden", forbidden)
	router.Handle("/chat", homeHandler(tpl))
	router.Handle("/ws", lib.WsHandler{H: H})
	log.Println("serving on port 8080")
	log.Println("Users:", dbUsers)
	//log.Println("Sessions: ", dbSessions)
	log.Fatal(http.ListenAndServe(":8080", router))
}

//TODO Current build is beta v1.0, it was released on 1/29/2017
//This version is not user friendly, this will change:)
