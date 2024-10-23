package chat

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/belalakhter-alphasquad/chat_template/internal/buffer"
	"github.com/belalakhter-alphasquad/chat_template/utils"
	"github.com/gorilla/websocket"
)

var chatService *Chat

const (
	Wait           = 60000 * time.Second
	maxMessageSize = 2048
)

type chat interface {
	Reader(client *Client)
	Writer()
}

type Chat struct {
	Clients    map[string]*Client
	Buffer     *buffer.Buffer
	Register   chan map[*Client][]byte
	UnRegister chan map[*Client][]byte
	Unicast    chan map[*Client][]byte
	BroadCast  chan []byte
}
type Client struct {
	conn *websocket.Conn
}
type Message struct {
	Type    int64  `json:"type"`
	Content string `json:"content"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func SetupChat(buf *buffer.Buffer) {
	if chatService != nil {
		chatService = nil
	}
	buf.BufferWorker()
	chatService = &Chat{
		Clients:    make(map[string]*Client),
		Buffer:     buf,
		Register:   make(chan map[*Client][]byte),
		UnRegister: make(chan map[*Client][]byte),
		Unicast:    make(chan map[*Client][]byte),
		BroadCast:  make(chan []byte),
	}
	go func() {
		chatService.Writer()
	}()

}

func UpgradeConnectionWs(w http.ResponseWriter, r *http.Request) {
	if chatService == nil {
		http.Error(w, "Websocket instance not live", int(http.StateClosed))
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		utils.LogMessage(err.Error(), 2)
		http.Error(w, "Failed to upgrade connection", http.StatusInternalServerError)
		return
	}
	params := r.URL.Query()
	user := params.Get("user")
	if user == "" {
		http.Error(w, "User not defined", int(http.StatusConflict))
		conn.Close()
	}
	client := &Client{conn: conn}
	chatService.Clients[user] = client

	go chatService.PumpMessages(client, chatService.Buffer.Messages)
	Notify := map[*Client][]byte{
		client: []byte(fmt.Sprintf("User has Joined %v", user)),
	}
	chatService.Register <- Notify

	go chatService.Reader(client, user)
}

func (c *Chat) Reader(client *Client, user string) {
	defer func() {
		client.conn.Close()
		Notify := map[*Client][]byte{
			client: []byte(fmt.Sprintf("User has left %v", user)),
		}
		chatService.UnRegister <- Notify
		delete(c.Clients, user)
	}()
	client.conn.SetReadLimit(maxMessageSize)
	client.conn.SetReadDeadline(time.Now().Add(Wait))
	for {
		var data Message
		_, msg, err := client.conn.ReadMessage()
		if err != nil {
			utils.LogMessage(err.Error(), 2)
			break
		}
		err = json.Unmarshal(msg, &data)
		if err != nil {
			utils.LogMessage(err.Error(), 1)
			continue
		}

		switch data.Type {
		case 1:

			chatService.BroadCast <- []byte(data.Content)
			c.Buffer.Pipe <- buffer.Messages{
				Msg:  data.Content,
				User: user,
			}

		case 2:

			uni := map[*Client][]byte{
				client: []byte(data.Content),
			}
			chatService.Unicast <- uni
		}

	}
}
func (c *Chat) Writer() {
	for {

		select {
		case data := <-chatService.BroadCast:
			go c.BroadCastWorker(data)
		case data := <-chatService.Unicast:
			for user, msg := range data {
				go c.UnicastWorker(user, msg)
			}
		case data := <-chatService.Register:
			for user, msg := range data {
				go c.UnicastWorker(user, msg)
			}
		case data := <-chatService.UnRegister:
			for user, msg := range data {
				go c.UnicastWorker(user, msg)
			}
		}
	}
}

func (c *Chat) BroadCastWorker(msg []byte) {

	for user, Conn := range c.Clients {
		message := map[string]string{
			"user": user,
			"msg":  string(msg),
		}
		data, err := json.Marshal(message)
		if err != nil {
			utils.LogMessage(err.Error(), 2)
		}
		Conn.conn.WriteMessage(websocket.TextMessage, data)

	}
}

func (c *Chat) UnicastWorker(user *Client, msg []byte) {
	message := map[string]string{
		"Notification": "Update",
		"msg":          string(msg),
	}
	data, err := json.Marshal(message)
	if err != nil {
		utils.LogMessage(err.Error(), 2)
	}
	user.conn.WriteMessage(websocket.TextMessage, data)

}

func (c *Chat) PumpMessages(user *Client, msgs []buffer.Messages) {
	for _, message := range msgs {

		data, err := json.Marshal(message)
		if err != nil {
			utils.LogMessage(err.Error(), 2)
		}

		user.conn.WriteMessage(websocket.TextMessage, data)
	}
}
