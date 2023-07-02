package kickchatwrapper

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/gorilla/websocket"
)

const (
	APIURL = "wss://ws-us2.pusher.com/app/eb1d5f283081a78b932c?protocol=7&client=js&version=7.6.0&flash=false"
)

type Client struct {
	ws   *websocket.Conn
	quit chan bool
}

type PusherSubscribe struct {
	Event string `json:"event"`
	Data  struct {
		Channel string `json:"channel"`
		Auth    string `json:"auth"`
	} `json:"data"`
}

type ChatMessageEvent struct {
	Event   string `json:"event"`
	Data    string `json:"data"`
	Channel string `json:"channel"`
}

type ChatMessage struct {
	ID         string `json:"id"`
	ChatroomID int    `json:"chatroom_id"`
	Content    string `json:"content"`
	Type       string `json:"type"`
	CreatedAt  string `json:"created_at"`
	Sender     Sender `json:"sender"`
}

type Sender struct {
	ID       int      `json:"id"`
	Username string   `json:"username"`
	Slug     string   `json:"slug"`
	Identity Identity `json:"identity"`
}

type Identity struct {
	Color  string  `json:"color"`
	Badges []Badge `json:"badges"`
}

type Badge struct {
	Type  string `json:"type"`
	Text  string `json:"text"`
	Count int    `json:"count"`
}

func NewClient() (*Client, error) {
	ws, _, err := websocket.DefaultDialer.Dial(APIURL, nil)
	if err != nil {
		return &Client{}, err
	}

	client := &Client{
		ws:   ws,
		quit: make(chan bool),
	}
	return client, err
}

func (client *Client) ListenForMessages() <-chan ChatMessage {
	ch := make(chan ChatMessage)
	go func() {
		for {
			select {
			case <-client.quit:
				return
			default:
				_, msg, err := client.ws.ReadMessage()
				if err != nil {
					fmt.Println("Error reading message", err)
					continue
				}

				var chatMessageEvent ChatMessageEvent
				errMarshalEvent := json.Unmarshal([]byte(msg), &chatMessageEvent)
				if errMarshalEvent != nil {
					continue
				}

				var chatMessage ChatMessage
				errMarshalMessage := json.Unmarshal([]byte(chatMessageEvent.Data), &chatMessage)
				if errMarshalMessage != nil {
					continue
				}

				ch <- chatMessage
			}
		}
	}()
	return ch
}

func (client *Client) JoinChannelByID(id int) error {
	pusherSubscribe := PusherSubscribe{
		Event: "pusher:subscribe",
		Data: struct {
			Channel string `json:"channel"`
			Auth    string `json:"auth"`
		}{
			Channel: "chatrooms." + strconv.Itoa(id) + ".v2",
			Auth:    "",
		},
	}

	msg, marshalErr := json.Marshal(pusherSubscribe)
	if marshalErr != nil {
		return errors.New("Marshal error")
	}

	err := client.ws.WriteMessage(websocket.TextMessage, []byte(msg))
	if err != nil {
		return errors.New("Error joining channel")
	}
	return nil
}

func (client *Client) Close() {
	client.quit <- true
	client.ws.Close()
}
