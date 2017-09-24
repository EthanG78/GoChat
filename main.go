package main

import (
	"flag"
	"log"
	"net/http"
	"text/template"
	"time"
	mw "github.com/EthanG78/golang_chat/middleware"
	"github.com/EthanG78/golang_chat/lib"
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
	var u User
	if c.Request().Method == http.MethodPost{
		un := c.Request().FormValue("username")
		p := c.Request().FormValue("password")


		//TODO: Make an individual way of handling when users do not insert anything into the fields..
		if un == "" {
			time.Sleep(3000)
			RedirectError := c.Redirect(http.StatusFound, "/401")
			//Error checking for testing
			if RedirectError != nil{
				log.Printf("Error: %v", RedirectError)
			}
		}
		if p == ""{
			time.Sleep(3000)
			RedirectError:= c.Redirect(http.StatusFound, "/401")
			//Error checking for testing
			if RedirectError != nil{
				log.Printf("Error: %v", RedirectError)
			}
		}
		pByte := []byte(p)
		finalP, err := bcrypt.GenerateFromPassword(pByte, 0)
		if err != nil{
			log.Fatalf("Error encrypting password: %v", err)
			//This is probably really bad, should find a better way to handle it lmao
		}

		u = User{un, finalP}

		dbUsers[un] = u
		RedirectError := c.Redirect(http.StatusFound, "/login")
		//Error checking for testing
		if RedirectError != nil{
			log.Printf("Error: %v", RedirectError)
		}


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
			RedirectError := c.Redirect(http.StatusFound, "/401")
			//Error checking for testing
			if RedirectError != nil{
				log.Printf("Error: %v", RedirectError)
			}
		}
		if p == "" {
			time.Sleep(3000)
			RedirectError := c.Redirect(http.StatusFound, "/401")
			//Error checking for testing
			if RedirectError != nil{
				log.Printf("Error: %v", RedirectError)
			}
		}

		u, ok := dbUsers[un]
		if !ok {
			time.Sleep(3000)
			RedirectError := c.Redirect(http.StatusFound, "/401")
			//Error checking for testing
			if RedirectError != nil{
				log.Printf("Error: %v", RedirectError)
			}
		}

		inputPass := []byte(p)
		userPass := u.Pass
		err := bcrypt.CompareHashAndPassword(userPass, inputPass)
		if err != nil{
			time.Sleep(3000)
			RedirectError := c.Redirect(http.StatusFound, "/401")
			//This is in the case that they input the incorrect password
			//Error checking for testing
			if RedirectError != nil{
				log.Printf("Error: %v", RedirectError)
			}
		}

		cookie := &http.Cookie{}
		cookie.Name = "session_id"
		cookie.Value = mw.CookieVal


		c.SetCookie(cookie)

		RedirectError := c.Redirect(http.StatusFound, "/chat")
		//Error checking for testing
		if RedirectError != nil{
			log.Printf("Error: %v", RedirectError)
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


	e.File("/favicon.ico", "static/styling/favicon.ico")

	//TODO: Find a way to store cookies
	//TODO: Assign groups, use logger, auth, server info and such
	//TODO: How can I store users without using a DB?????
	//TODO: Maybe generate cookie during login? Then ask for it within the chat!
	//TODO: I also really need to re-style the web pages... They are garbage


	//TODO: COOKIES IN LOGIN!!!!!! (please don't forget this)

	//TODO: I also have to use echo's websockets... That's going to be brutal


	//GROUPS
	admin := e.Group("/admin")


	//MIDDLEWARE
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
	e.POST("/signup", sign_up)
	e.File("/signup", "static/signup.html")
	e.POST("/login", login)
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
