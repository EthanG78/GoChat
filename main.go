package main

import (
	"flag"
	"log"
	"net/http"
	"text/template"
	"time"

	"github.com/EthanG78/golang_chat/lib"
	mw "github.com/EthanG78/golang_chat/middleware"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

//Ethan Zaat is a cool dude;)

//Default User type
type User struct {
	Username string `json:"username"`
	Pass     []byte `json:"pass"`
}

var dbUsers = map[string]User{}
var dbSessions = map[string]string{}

func home(c echo.Context) error {
	return c.String(http.StatusOK, "home")
}

func four0one(c echo.Context) error {
	return c.String(http.StatusUnauthorized, "Nice try buster, you are unauthorized!")
}

func signUp(c echo.Context) error {
	var u User
	if c.Request().Method == http.MethodPost {
		un := c.Request().FormValue("username")
		p := c.Request().FormValue("password")

		//TODO: Make an individual way of handling when users do not insert anything into the fields..
		if un == "" {
			time.Sleep(3000)
			RedirectError := c.Redirect(http.StatusFound, "/401")
			//Error checking for testing
			if RedirectError != nil {
				log.Printf("Error: %v", RedirectError)
			}
		}
		if p == "" {
			time.Sleep(3000)
			RedirectError := c.Redirect(http.StatusFound, "/401")
			//Error checking for testing
			if RedirectError != nil {
				log.Printf("Error: %v", RedirectError)
			}
		}
		if _, ok := dbUsers[un]; ok {
			return c.String(http.StatusForbidden, "Username has alreaedy been taken")
		}

		sessionID := uuid.NewV4()
		cookie := &http.Cookie{}
		cookie.Name = "session_id"
		cookie.Value = sessionID.String()
		c.SetCookie(cookie)

		dbSessions[cookie.Value] = un

		finalP, err := bcrypt.GenerateFromPassword([]byte(p), 0)
		if err != nil {
			log.Fatalf("Error encrypting password: %v", err)
			//This is probably really bad, should find a better way to handle it lmao
		}

		u = User{un, finalP}

		dbUsers[un] = u
		RedirectError := c.Redirect(http.StatusFound, "/login")
		//Error checking for testing
		if RedirectError != nil {
			log.Printf("Error: %v", RedirectError)
		}

		//FOR DEBUGGING
		log.Println(dbUsers)
		log.Println(dbSessions)

		return c.String(http.StatusOK, "you have successfully signed up!")

	}

	return c.String(http.StatusBadRequest, "You could not be signed up")

}

func login(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		un := c.Request().FormValue("username")
		p := c.Request().FormValue("password")

		if un == "" {
			time.Sleep(3000)
			RedirectError := c.Redirect(http.StatusFound, "/401")
			//Error checking for testing
			if RedirectError != nil {
				log.Printf("Error: %v", RedirectError)
			}
		}
		if p == "" {
			time.Sleep(3000)
			RedirectError := c.Redirect(http.StatusFound, "/401")
			//Error checking for testing
			if RedirectError != nil {
				log.Printf("Error: %v", RedirectError)
			}
		}

		u, ok := dbUsers[un]
		if !ok {
			time.Sleep(3000)
			RedirectError := c.Redirect(http.StatusFound, "/401")
			//Error checking for testing
			if RedirectError != nil {
				log.Printf("Error: %v", RedirectError)
			}
		}

		err := bcrypt.CompareHashAndPassword(u.Pass, []byte(p))
		if err != nil {
			time.Sleep(3000)
			RedirectError := c.Redirect(http.StatusFound, "/401")
			//This is in the case that they input the incorrect password
			//Error checking for testing
			if RedirectError != nil {
				log.Printf("Error: %v", RedirectError)
			}
		}

		cookie := &http.Cookie{}
		sessionID := uuid.NewV4()
		cookie.Name = "session_id"
		cookie.Value = sessionID.String()
		c.SetCookie(cookie)

		dbSessions[cookie.Value] = un

		RedirectError := c.Redirect(http.StatusFound, "/chat/")
		//Error checking for testing
		if RedirectError != nil {
			log.Printf("Error: %v", RedirectError)
		}
		return c.String(http.StatusOK, "You have successfully logged in!")

	}

	return c.String(http.StatusBadRequest, "You could not log in")
}

func chatMain(c echo.Context) error {
	tpl := template.Must(template.ParseGlob("static/chat.html"))
	tpl.Execute(c.Response(), c.Request())
	return c.String(http.StatusOK, "")
}

func adminMain(c echo.Context) error {
	return c.String(http.StatusOK, "Welcome to the admin page.")
}

//////
//MAIN
//////
func main() {

	e := echo.New()
	flag.Parse()

	e.File("/favicon.ico", "static/styling/favicon.ico")

	//TODO: Find a way to store cookies
	//TODO: Assign groups, use logger, auth, server info and such
	//TODO: How can I store users without using a DB?????
	//TODO: I also really need to re-style the web pages... They are garbage

	//TODO: ADD SESSION COOKIES, so no random dude can access chat without logging in

	//TODO: I also have to use echo's websockets... That's going to be brutal

	//TODO: Create admin group and page

	//GROUPS
	admin := e.Group("/admin")
	chat := e.Group("/chat")

	//MIDDLEWARE
	e.Use(mw.ServerInfo)

	chat.Use(mw.CheckCookies)

	chat.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: `[${time_rfc3339}] ${status} ${method} ${host}${path} ${latency}` + "\n",
	}))

	admin.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: `[${time_rfc3339}] ${status} ${method} ${host}${path} ${latency}` + "\n",
	}))

	admin.Use(middleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
		if username == mw.AdminLogin && password == mw.AdminPassword {
			return true, nil
		}

		return false, nil
	}))

	//WEBSOCKETS
	e.GET("/ws", lib.Chat)

	//ENDPOINTS
	e.GET("/", home)
	e.File("/", "static/home.html")
	e.GET("/401", four0one)
	e.File("/401", "static/forbidden.html")
	e.POST("/signup", signUp)
	e.File("/signup", "static/signup.html")
	e.POST("/login", login)
	e.File("/login", "static/login.html")
	chat.GET("/", chatMain)
	admin.GET("/", adminMain)
	//admin.File("/", "static/admin.html")

	//CREATE SERVER
	e.Logger.Fatal(e.Start(":8080"))

}

//TODO: Current build is beta v1.0, it was released on 1/29/2017
