package bot

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// Hub объект, который будет хранить каналы и зарегистрированных пользователей
type Hub struct {
	clients map[int]*Client

	Ban map[int]*Client

	register chan *Client

	setAction chan *Client

	message chan *Client

	token string

	api string

	adminPass string

	Offset int
}

// NewHub метод, инициализирующий объект Hub
func NewHub() *Hub {
	return &Hub{
		clients:   make(map[int]*Client),
		Ban:       make(map[int]*Client),
		register:  make(chan *Client),
		setAction: make(chan *Client),
		message:   make(chan *Client),
		token:     "bot**********************************************", // Токен от вашего бота telegram
		api:       "https://api.telegram.org",
		adminPass: "***********************", // Пароль администратора
		Offset:    0,
	}
}

// Метод для запросов к telegram api
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

// Run ... бесконечный цикл на прослушивание каналов объекта Hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register: // Добавление юзера в систему
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
		case client := <-h.message: // Обработка сообщений от юзера
			if _, ok := h.clients[client.ChatID]; ok && client.Answer != "" {
				h.clients[client.ChatID].Answer = client.Answer
				if h.clients[client.ChatID].IsAdmin {
					switch client.Answer {
					case "users":
						client.Message = "<b>Зарегистрированные юзеры:</b> \n \n"
						for _, v := range h.clients {
							client.Message = client.Message +
								"<b>#" + v.Username + "</b>\n" +
								"chat_id: " + strconv.Itoa(v.ChatID) + "\n" +
								"admin: " + strconv.FormatBool(v.IsAdmin) + "\n \n"
						}
						client.Message = client.Message + "<b>Заблокированные юзеры:</b> \n \n"
						for _, v := range h.Ban {
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
		case client := <-h.setAction: // Выполнить действие на метод, вызванный юзером
			if _, ok := h.clients[client.ChatID]; ok {
				switch h.clients[client.ChatID].Answer {
				case "admin": // Дать админский доступ
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
				case "broadcast": // Транслировать свое сообщение всем юзерам
					if h.clients[client.ChatID].IsAdmin {
						fromChatID := strconv.Itoa(client.ChatID)
						messageID := strconv.Itoa(client.MessageID)
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
				case "ban": // Ограничить доступ юзеру к боту
					userID, err := strconv.Atoi(client.Message)
					if err != nil {
						client.Message = "chat_id введен не верно"
					} else {
						if _, ok := h.clients[userID]; ok {
							if _, ok := h.Ban[userID]; ok {
								client.Message = "Юзер уже в бане"
							} else {
								h.Ban[userID] = h.clients[userID]
								delete(h.clients, userID)
								client.Message = "Юзер внесен в список заблокированных"
							}
						} else {
							client.Message = "Юзер не найден в системе"
						}
					}
				case "unban": // Разблокировать доступ юзеру к боту
					userID, err := strconv.Atoi(client.Message)
					if err != nil {
						client.Message = "chat_id введен не верно"
					} else {
						if _, ok := h.Ban[userID]; ok {
							delete(h.Ban, userID)
							client.Message = "Юзер разблокирован"
						} else {
							client.Message = "Юзер не найден в бане"
						}
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
		}
	}
}
