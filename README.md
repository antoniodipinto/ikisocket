
# WebSocket wrapper for [Fiber v2](https://github.com/gofiber/fiber) with events support
[![Go Report Card](https://goreportcard.com/badge/github.com/antoniodipinto/ikisocket)](https://goreportcard.com/report/github.com/antoniodipinto/ikisocket)
[![GoDoc](https://godoc.org/github.com/antoniodipinto/ikisocket?status.svg)](https://godoc.org/github.com/antoniodipinto/ikisocket)
[![GitHub license](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/antoniodipinto/ikisocket/blob/master/LICENSE)
### Based on [Fiber Websocket](https://github.com/gofiber/websocket) and inspired by [Socket.io](https://github.com/socketio/socket.io)

### Upgrade to Fiber v2 details [here](https://github.com/antoniodipinto/ikisocket/issues/6) 

## ⚙️ Installation

```
go get -u github.com/antoniodipinto/ikisocket
```

## ⚡️ Basic chat example

```go
package main

import (
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/antoniodipinto/ikisocket"
	"log"
)

// Basic chat message object
type MessageObject struct {
	Data string `json:"data"`
	From string `json:"from"`
	To   string `json:"to"`
}

func main() {

	// The key for the map is message.to
	clients := make(map[string]string)

	// Start a new Fiber application
	app := fiber.New()

	// Setup the middleware to retrieve the data sent in first GET request
	app.Use(func(c *fiber.Ctx) error {
		// IsWebSocketUpgrade returns true if the client
		// requested upgrade to the WebSocket protocol.
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	// Multiple event handling supported
	ikisocket.On(ikisocket.EventConnect, func(ep *ikisocket.EventPayload) {
		fmt.Println(fmt.Sprintf("Connection event 1 - User: %s", ep.SocketAttributes["user_id"]))
	})

	ikisocket.On(ikisocket.EventConnect, func(ep *ikisocket.EventPayload) {
		fmt.Println(fmt.Sprintf("Connection event 2 - User: %s", ep.SocketAttributes["user_id"]))
	})

	// On message event
	ikisocket.On(ikisocket.EventMessage, func(ep *ikisocket.EventPayload) {

		fmt.Println(fmt.Sprintf("Message event - User: %s - Message: %s", ep.SocketAttributes["user_id"], string(ep.Data)))

		message := MessageObject{}

		err := json.Unmarshal(ep.Data, &message)
		if err != nil {
			fmt.Println(err)
			return
		}

		// Emit the message directly to specified user
		err = ep.Kws.EmitTo(clients[message.To], ep.Data)
		if err != nil {
			fmt.Println(err)
		}
	})

	// On disconnect event
	ikisocket.On(ikisocket.EventDisconnect, func(ep *ikisocket.EventPayload) {
		// Remove the user from the local clients
		delete(clients, ep.SocketAttributes["user_id"])
		fmt.Println(fmt.Sprintf("Disconnection event - User: %s", ep.SocketAttributes["user_id"]))
	})

	// On error event
	ikisocket.On(ikisocket.EventError, func(ep *ikisocket.EventPayload) {
		fmt.Println(fmt.Sprintf("Error event - User: %s", ep.SocketAttributes["user_id"]))
	})

	app.Get("/ws/:id", ikisocket.New(func(kws *ikisocket.Websocket) {

		userId := kws.Params("id")

		// Add the connection to the list of the connected clients
		// The UUID is generated randomly and is the key that allow
		// ikisocket to manage Emit/EmitTo/Broadcast
		clients[userId] = kws.UUID

		// Every websocket connection has an optional session key => value storage
		kws.SetAttribute("user_id", userId)

		//Broadcast to all the connected users the newcomer
		kws.Broadcast([]byte(fmt.Sprintf("New user connected: %s and UUID: %s", userId, kws.UUID)), true)
		//Write welcome message
		kws.Emit([]byte(fmt.Sprintf("Hello user: %s with UUID: %s", userId, kws.UUID)))
	}))

	log.Fatal(app.Listen(":3000"))
}
```
### Connect to the websocket

```
ws://localhost:3000/ws/<user-id>
```
