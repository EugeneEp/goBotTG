package bot

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// Client Объект юзер
type Client struct {
	ChatID    int
	Username  string
	Message   string
	MessageID int
	IsAdmin   bool
	Answer    string
}

// APIResponse Объект ответ telegram api
type APIResponse struct {
	Ok     bool `json:"ok"`
	Result []struct {
		UpdateID int `json:"update_id"`
		Message  struct {
			MessageID int `json:"message_id"`
			From      struct {
				ID           int    `json:"id"`
				IsBot        bool   `json:"is_bot"`
				FirstName    string `json:"first_name"`
				LastName     string `json:"last_name"`
				Username     string `json:"username"`
				LanguageCode string `json:"language_code"`
			} `json:"from"`
			Chat struct {
				ID        int    `json:"id"`
				FirstName string `json:"first_name"`
				LastName  string `json:"last_name"`
				Username  string `json:"username"`
				Type      string `json:"type"`
			} `json:"chat"`
			Date int    `json:"date"`
			Text string `json:"text"`
		} `json:"message"`
	} `json:"result"`
}

// ServeClient Слушаем обновления telegram api
func ServeClient(hub *Hub) {
	ticker := time.NewTicker(3 * time.Second)
	apiURL := hub.api + "/" + hub.token
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C: // Каждые 3 секунды проверяем обновления, отщищаем предыдущие
			resp, err := http.Get(apiURL + "/getUpdates?offset=" + strconv.Itoa(hub.Offset))
			if err != nil {
				fmt.Println(err)
				return
			}
			defer resp.Body.Close()
			var message APIResponse
			json.NewDecoder(resp.Body).Decode(&message)

			// Получаем тело ответа
			for _, v := range message.Result {
				client := &Client{
					ChatID:    v.Message.Chat.ID,
					Username:  v.Message.Chat.Username,
					Answer:    "",
					MessageID: v.Message.MessageID,
				}
				if _, ok := hub.Ban[client.ChatID]; ok {
					v.Message.Text = "/banned"
				}

				switch v.Message.Text {
				case "/start":
					hub.register <- client
				case "/help":
					client.Message = "<b>Доступные команды:</b>\n" +
						"\n" +
						"/start Зарегистрировать себя в боте\n" +
						"/help Доступные команды\n" +
						"/admin Дать себе админские права\n" +
						"\n" +
						"<b>Доступные команды для администратора:</b>\n" +
						"\n" +
						"/users Получить список пользователей зарегистрированных в системе\n" +
						"/broadcast Отправить массовое сообщение\n" +
						"/ban Забанить пользователя в системе\n"
					hub.message <- client
				case "/admin":
					client.Message = "Введите пароль"
					client.Answer = "admin"
					hub.message <- client
				case "/users":
					client.Answer = "users"
					hub.message <- client
				case "/broadcast":
					client.Message = "Введите сообщение"
					client.Answer = "broadcast"
					hub.message <- client
				case "/ban":
					client.Message = "Введите chat_id юзера, которого хотите забанить"
					client.Answer = "ban"
					hub.message <- client
				case "/unban":
					client.Message = "Введите chat_id юзера, которого хотите разбанить"
					client.Answer = "unban"
					hub.message <- client
				case "/banned":
					client.Message = "К сожалению, вы заблокированы"
					hub.message <- client
				default:
					client.Message = v.Message.Text
					hub.setAction <- client
				}
			}

			// Если количество сообщений больше 0,
			// записываем update_id последнего сообщения в канал,
			// для дальнейшей отчистки telegram api от предыдущих сообщений
			length := len(message.Result)
			if length > 0 {
				hub.Offset = message.Result[length-1].UpdateID + 1
			}

		}
	}
}
