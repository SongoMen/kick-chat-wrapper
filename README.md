# Introduction
`kickchatwrapper` is a Go package that provides a wrapper for interacting with the Kick websocket. It allows you to subscribe to chat channels and receive chat messages.

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
