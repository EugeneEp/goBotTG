package main

import (
	"goBotTG/bot"
	"net/http"
)

func main() {
	hub := bot.NewHub()     // Инициализируем объект с каналами
	go hub.Run()            // Запускаем хаб на прослушивание каналов
	go bot.ServeClient(hub) // Запускаем метод на прослушивание обновлений в telegram
	http.ListenAndServe(":8080", nil)
}
