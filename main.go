package main

import (
	"flag"
	"log"
	"net/http"
	"text/template"
	"time"

	"github.com/EthanG78/golang_chat/lib"
	"github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

//Ethan Zaat is a cool dude;)

type User struct {
	Username 		string		`json:"username"`
	Pass     		[]byte		`json:"pass"`
}

var dbUsers = map[string]User{}

func home (c echo.Context) error{
	return c.String(http.StatusOK, "home")
}

func four_o_one (c echo.Context) error{
	return c.String(http.StatusUnauthorized, "Nice try buster, you are unauthorized!")
}

func sign_up (c echo.Context) error{
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

func login (c echo.Context) error{
	if c.Request().Method == http.MethodPost{
		un := c.Request().FormValue("username")
		p := c.Request().FormValue("password")

		if un == "" {
			time.Sleep(3000)
			c.Redirect(http.StatusUnauthorized, "/401")
		}
		if p == "" {
			time.Sleep(3000)
			c.Redirect(http.StatusUnauthorized, "/401")
		}

		u, ok := dbUsers[un]
		if !ok {
			time.Sleep(3000)
			c.Redirect(http.StatusUnauthorized, "/401")
		}

		inputPass := []byte(p)
		userPass := u.Pass
		err := bcrypt.CompareHashAndPassword(userPass, inputPass)
		if err != nil{
			time.Sleep(3000)
			c.Redirect(http.StatusUnauthorized, "/401")
			//This is in the case that they input the incorrect password
		}

		return c.String(http.StatusOK, "You have successfully logged in!")

	}

	return c.String(http.StatusBadRequest, "You could not log in")
}


func homeHandler(tpl *template.Template) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tpl.Execute(w, r)
	})
}



//////
//MAIN
//////
func main() {

	e := echo.New()


	e.File("/favicon.ico", "styling/favicon.ico")

	//TODO: Use uuidV4 for cookie checker and finish middleware
	//TODO: Find a way to store cookies
	//TODO: Assign groups, use logger, auth, server info and such
	//TODO: How can I store users without using a DB?????
	//TODO: Maybe generate cookie during login? Then ask for it within the chat!
	//TODO: I also really need to re-style the web pages... They are garbage


	//GROUPS
	admin := e.Group("/admin")


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
	e.GET("/401", four_o_one)
	e.File("/401", "static/forbidden.html")
	e.GET("/signup", sign_up)
	e.File("/signup", "static/signup.html")
	e.GET("/login", login)
	e.File("/login", "static/login.html")

	//CREATE SERVER
	e.Logger.Fatal(e.Start(":8080"))




	//OLD CODE
	flag.Parse()
	tpl := template.Must(template.ParseFiles("static/chat.html"))
	H := lib.NewHub()
	router := http.NewServeMux()
	router.Handle("/styling/", http.StripPrefix("/styling/", http.FileServer(http.Dir("styling/"))))
	router.Handle("/chat", homeHandler(tpl))
	router.Handle("/ws", lib.WsHandler{H: H})
}

//TODO Current build is beta v1.0, it was released on 1/29/2017
//This version is not user friendly, this will change:)
