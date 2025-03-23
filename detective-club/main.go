package main

import (
	"detective-club/model"
	"log"
	"net/http"
)

func main() {
	fs := http.FileServer(http.Dir("./static/"))
	http.Handle("/detective-club/", http.StripPrefix("/detective-club/", fs))

	r := model.NewRoom()
	go r.Run()
	http.Handle("/detective-club/ws", r)

	err := http.ListenAndServe(":4000", nil)
	if err != nil {
		log.Fatal(err)
	}
}
