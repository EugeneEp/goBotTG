package main

import (
	"goBotTG/bot"
	"net/http"
)

func main() {
	hub := bot.NewHub()
	go hub.Run()
	go bot.ServeClient(hub)
	http.ListenAndServe(":8080", nil)
}
