package main

import (
	"flag"
	"log"
	"net/http"
	"text/template"
	"golang.org/x/crypto/bcrypt"
	"github.com/EthanG78/golang_chat/lib"
	"github.com/satori/go.uuid"
	"math/rand"
	"time"
	"io"
	"image/color"
	"image/gif"
	"image"
	"math"
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
		pass := req.FormValue("password")

		if un == ""{
			time.Sleep(3000)
			http.Redirect(w, req, "/forbidden", http.StatusSeeOther)
			return

		}else if pass == ""{
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

		password := []byte(pass)
		hash := []byte(u.Pass)
		err := bcrypt.CompareHashAndPassword(hash, password)
		if err != nil{
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
		http.Redirect(w, req, "/chat", http.StatusSeeOther)
		return
	}

	tpl.ExecuteTemplate(w, "login.gohtml", nil)
}

func faviconHandler(w http.ResponseWriter, r *http.Request){
	http.ServeFile(w, r, "styling/logo/favicon.ico")
}

//SECRET ANIMATED GIF
var palette = []color.Color{color.White, color.Black}

const(
	whiteIndex = 0 //first color in palette
	blackIndex = 1 //next color in palette
)

func lassajous(out io.Writer)  {
	const(
		cycles = 5 //number of complex x oscillator revolutions
		res = 0.001 // angular resolution
		size = 100 //image canvas covers [-size..+size]
		nframes = 64 //number of animation frames
		delay = 8 // delay between frames in 10ms units

	)

	freq := rand.Float64()* 3.0
	anim := gif.GIF{LoopCount: nframes}
	phase := 0.0
	for i := 0; i < nframes; i++ {
		rect := image.Rect(0,0,2*size+1,2*size+1)
		img := image.NewPaletted(rect, palette)
		for t := 0.0; t < cycles*2*math.Pi; t += res{
			x := math.Sin(t)
			y := math.Sin(t*freq + phase)
			img.SetColorIndex(size+int(x*size+0.5), size+int(y*size+0.5), blackIndex)
		}
		phase += 0.1
		anim.Delay = append(anim.Delay, delay)
		anim.Image = append(anim.Image, img)
	}
	gif.EncodeAll(out, &anim) // ignoring encoding errors.

}

func lassajousHandler (w http.ResponseWriter, r *http.Request){
	lassajous(w)
}
func main() {

	flag.Parse()
	tpl := template.Must(template.ParseFiles("templates/chat.gohtml"))
	H := lib.NewHub()
	router := http.NewServeMux()
	router.HandleFunc("/favicon.ico", faviconHandler)
	router.HandleFunc("/", sign_up)
	router.HandleFunc("/login", login)
	router.HandleFunc("/forbidden", forbidden)
	router.HandleFunc("/lassajous", lassajousHandler)
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

