package main

import (
	"flag"
	"log"
	"net/http"
	"html/template"

	"github.com/EthanG78/golang_chat/controller"
)

func Handler (tpl *template.Template) http.Handler{
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
		tpl.Execute(w, r)
	})
}

func main() {
	flag.Parse()
	tpl := template.Must(template.ParseFiles("template/main.html"))
	h := newHub()
	router := http.NewServeMux()
	router.Handle("/", Handler(tpl))
	router.Handle("/ws", wsHandler{h: h})
	log.Println("Serving on port 8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}