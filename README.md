# Introduction
`kickchatwrapper` is a Go package that provides a wrapper for interacting with the KickChat API. It allows you to connect to a KickChat WebSocket server, subscribe to chat channels, receive chat messages, and perform other actions related to chat functionality.

# Installation
```
go get github.com/your-username/kickchatwrapper
```

# Usage

```
client, err := kickchatwrapper.NewClient()
if err != nil {
    // handle error
}

client.JoinChannelByID(231221)

messageChan := client.ListenForMessages()

go func() {
  for message := range messageChan {
    fmt.Printf("Received chat message: %+v\n", message)
  }
}()

// To close connection you can call
// client.Close()
```
# Notes
Right now it is only possible to join chat room using user ID because to be able to join it by username we would need to call their API but its protected by CloudFlare.
