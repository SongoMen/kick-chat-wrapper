package kickchatwrapper

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

const (
	APIURL = "wss://ws-us2.pusher.com/app/eb1d5f283081a78b932c?protocol=7&client=js&version=7.6.0&flash=false"
)

type Client struct {
	ws             *websocket.Conn
	joinedChannels []int
	quit           chan bool
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
	ID         string    `json:"id"`
	ChatroomID int       `json:"chatroom_id"`
	Content    string    `json:"content"`
	Type       string    `json:"type"`
	CreatedAt  time.Time `json:"created_at"`
	Sender     Sender    `json:"sender"`
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

func (client *Client) reconnect() error {
	client.ws.Close()

	ws, _, dialErr := websocket.DefaultDialer.Dial(APIURL, nil)
	if dialErr != nil {
		return dialErr
	}

	client.ws = ws

	for id := range client.joinedChannels {
		joinErr := client.JoinChannelByID(id)
		if joinErr != nil {
			return joinErr
		}
	}

	return nil
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
					if err.(*websocket.CloseError).Code == 4200 || err.(*websocket.CloseError).Code == 4201 {
						fmt.Println("Connection lost, Reconnecting...")
						reconnectErr := client.reconnect()
						if reconnectErr != nil {
							fmt.Println("Error reconnecting:", reconnectErr)
							time.Sleep(time.Second)
						}
					} else {
						fmt.Println("Error reading message:", err)
					}
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

	client.joinedChannels = append(client.joinedChannels, id)
	return nil
}

func (client *Client) Close() {
	client.quit <- true
	client.ws.Close()
}
