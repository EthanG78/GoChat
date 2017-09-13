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
	"github.com/labstack/echo/middleware"
)

//User type referenced in DB
type User struct {
	Username 		string		`json:"username"`
	Pass     		[]byte		`json:"pass"`
}

//Database and template variables
var dbUsers = map[string]User{}
var dbSessions = map[string]string{}
var tpl *template.Template



func home (c echo.Context) error{
	return c.String(http.StatusOK, "home")
}

func forbidden (c echo.Context) error{
	return c.String(http.StatusUnauthorized, "Nice try buster, you are unauthorized!")
}

func signup (c echo.Context) error{
	cookie := &http.Cookie{}
	cookieValue := uuid.NewV4()

	cookie.Name = "session_id"
	cookie.Value = cookieValue.String()

	c.SetCookie(cookie)

	var u User
	if c.Request().Method == http.MethodPost{
		un := c.Request().FormValue("username")
		p := c.Request().FormValue("password")


		//TODO: Make an individual way of handling when users do not insert anything into the fields..
		if un == "" {
			time.Sleep(3000)
			c.Redirect(http.StatusUnauthorized, "/401")
		}
		if p == ""{
			time.Sleep(3000)
			c.Redirect(http.StatusUnauthorized, "/401")
		}
		pByte := []byte(p)
		finalP, err := bcrypt.GenerateFromPassword(pByte, 0)
		if err != nil{
			log.Fatalf("Error encrypting password: %v", err)
			//This is probably really bad, should find a better way to handle it lmao
		}

		cookie.Value = un

		u = User{un, finalP}

		dbUsers[cookie.Value] = u
		c.Redirect(http.StatusOK, "/login")


		//FOR DEBUGGING
		log.Println(dbUsers)

		return c.String(http.StatusOK, "you have successfully signed up!")

	}


	return c.String(http.StatusBadRequest, "You could not be signed up")

}


func homeHandler(tpl *template.Template) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tpl.Execute(w, r)
	})
}





//OLD FUNCTION

/*func signUp(w http.ResponseWriter, req *http.Request) {
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

	tpl.ExecuteTemplate(w, "signup.html", nil)
}*/

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

	tpl.ExecuteTemplate(w, "login.html", nil)

}


//////
//MAIN
//////
func main() {

	e := echo.New()


	e.File("/favicon.ico", "styling/favicon.ico")

	//TODO: Create endpoints for each WebPage
	//TODO: Use uuidV4 for cookie checker and finish middleware
	//TODO: Find a way to store cookies
	//TODO: Figure out how to redirect using echo
	//TODO: Assign groups, use logger, auth, server info and such
	//TODO: How can I store users without using a DB?????


	//GROUPS
	admin := e.Group("/admin")
	login := e.Group("/login")


	//MIDDLEWARE
	e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		Root: "static",
		Browse: false,
		Index: "home.html",
	}))

	admin.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: `[${time_rfc3339}] ${status} ${method} ${host}${path} ${latency}` + "\n",
	}))

	admin.Use(middleware.BasicAuth(func (username, password string, c echo.Context) (bool, error) {
		//placeholders for now
		if username == "Admin" && password == "admin"{
			return true, nil
		}

		return false, nil
	}))

	//ENDPOINTS
	e.GET("/", home)
	e.File("/", "static/home.html")
	e.GET("/401", forbidden)
	e.File("/401", "static/forbidden")
	e.GET("/signup", signup)
	e.File("/signup", "static/signup")

	//CREATE SERVER
	e.Logger.Fatal(e.Start(":8080"))




	//OLD CODE
	flag.Parse()
	tpl := template.Must(template.ParseFiles("static/chat.html"))
	H := lib.NewHub()
	router := http.NewServeMux()
	router.Handle("/styling/", http.StripPrefix("/styling/", http.FileServer(http.Dir("styling/"))))
	router.HandleFunc("/login", login)
	router.Handle("/chat", homeHandler(tpl))
	router.Handle("/ws", lib.WsHandler{H: H})
	log.Println("serving on port 8080")
	log.Println("Users:", dbUsers)
	//log.Println("Sessions: ", dbSessions)
	log.Fatal(http.ListenAndServe(":8080", router))
}

//TODO Current build is beta v1.0, it was released on 1/29/2017
//This version is not user friendly, this will change:)
