package bot

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type Hub struct {
	clients map[int]*Client

	register chan *Client

	setAction chan *Client

	message chan *Client

	unregister chan *Client

	token string

	api string

	adminPass string

	Offset int
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[int]*Client),
		register:   make(chan *Client),
		setAction:  make(chan *Client),
		unregister: make(chan *Client),
		message:    make(chan *Client),
		token:      "bot1446985566:AAF73FkSr8xO9UhmY6IPdNWu3avue16H2SI",
		api:        "https://api.telegram.org",
		adminPass:  "ji128u*(WHJd898wyu9j24)",
		Offset:     0,
	}
}

func (h *Hub) sendMessage(Method string, Params map[string]string) {

	apiURL := h.api + "/" + h.token + "/" + Method + "?"
	for k, v := range Params {
		apiURL = apiURL + k + "=" + url.QueryEscape(v) + "&"
	}
	_, err := http.Get(apiURL)
	if err != nil {
		fmt.Println(err)
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			if _, ok := h.clients[client.ChatID]; ok {
				client.Message = "Бот уже зарегистрировал вас"
				h.clients[client.ChatID].Answer = ""
			} else {
				h.clients[client.ChatID] = client
				client.Message = "Бот успешно добавил вас в систему"
			}
			query := map[string]string{
				"chat_id":    strconv.Itoa(client.ChatID),
				"text":       client.Message,
				"parse_mode": "HTML",
			}
			go h.sendMessage("sendMessage", query)
		case client := <-h.message:
			if _, ok := h.clients[client.ChatID]; ok && client.Answer != "" {
				h.clients[client.ChatID].Answer = client.Answer
				if h.clients[client.ChatID].IsAdmin {
					switch client.Answer {
					case "users":
						client.Message = "Зарегистрированные юзеры: \n \n"
						for _, v := range h.clients {
							client.Message = client.Message +
								"<b>#" + v.Username + "</b>\n" +
								"chat_id: " + strconv.Itoa(v.ChatID) + "\n" +
								"admin: " + strconv.FormatBool(v.IsAdmin) + "\n \n"
						}
					}
				}
			}
			query := map[string]string{
				"chat_id":    strconv.Itoa(client.ChatID),
				"text":       client.Message,
				"parse_mode": "HTML",
			}
			go h.sendMessage("sendMessage", query)
		case client := <-h.setAction:
			if _, ok := h.clients[client.ChatID]; ok {
				switch h.clients[client.ChatID].Answer {
				case "admin":
					if h.clients[client.ChatID].IsAdmin {
						client.Message = "У вас уже есть админский доступ"
					} else {
						if client.Message != h.adminPass {
							client.Message = "Неправильный пароль"
						} else {
							h.clients[client.ChatID].IsAdmin = true
							client.Message = "У вас теперь есть админский доступ"
						}
					}
				case "broadcast":
					if h.clients[client.ChatID].IsAdmin {
						fromChatID := strconv.Itoa(client.ChatID)
						messageID := strconv.Itoa(client.MessageId)
						for k := range h.clients {
							chatID := strconv.Itoa(k)
							if chatID != fromChatID {
								query := map[string]string{
									"chat_id":      chatID,
									"from_chat_id": fromChatID,
									"message_id":   messageID,
								}
								go h.sendMessage("forwardMessage", query)
							}
						}
						client.Message = "Сообщение разослано"
					} else {
						client.Message = "У вас нет доступа"
					}
				default:
					client.Message = "Метод не найден"
				}
			} else {
				client.Message = "Вы должны добавить себя в систему /start"
			}
			query := map[string]string{
				"chat_id":    strconv.Itoa(client.ChatID),
				"text":       client.Message,
				"parse_mode": "HTML",
			}
			go h.sendMessage("sendMessage", query)
		case client := <-h.unregister:
			fmt.Println(client)
		}
	}
}
