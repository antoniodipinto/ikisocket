
# WebSocket wrapper for [Fiber v2](https://github.com/gofiber/fiber) with events support
[![Go Report Card](https://goreportcard.com/badge/github.com/antoniodipinto/ikisocket)](https://goreportcard.com/report/github.com/antoniodipinto/ikisocket)
[![GoDoc](https://godoc.org/github.com/antoniodipinto/ikisocket?status.svg)](https://godoc.org/github.com/antoniodipinto/ikisocket)
[![GitHub license](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/antoniodipinto/ikisocket/blob/master/LICENSE)
### Based on [Fiber Websocket](https://github.com/gofiber/websocket) and inspired by [Socket.io](https://github.com/socketio/socket.io)

### Upgrade to Fiber v2 details [here](https://github.com/antoniodipinto/ikisocket/issues/6) 


## Any bug?
Create ad issue following [this](https://github.com/antoniodipinto/ikisocket/blob/master/.github/ISSUE_TEMPLATE/bug_report.md) template


## Feature request?
Create ad issue following [this](https://github.com/antoniodipinto/ikisocket/blob/master/.github/ISSUE_TEMPLATE/feature_request.md) template



## ⚙️ Installation

```
go get -u github.com/antoniodipinto/ikisocket
```

## 📖 ️ Documentation

```go
// Initialize new ikisocket in the callback this will
// execute a callback that expects kws *Websocket Object
New(callback func(kws *Websocket)) func(*fiber.Ctx) error
```
---
```go
// Add listener callback for an event into the listeners list
On(event string, callback func(payload *EventPayload))
```
Supported events:

```go
// Supported event list
const (
	// Fired when a Text/Binary message is received
	EventMessage = "message"
	// More details here:
	// @url https://developer.mozilla.org/en-US/docs/Web/API/WebSockets_API/Writing_WebSocket_servers#Pings_and_Pongs_The_Heartbeat_of_WebSockets
	EventPing = "ping"
	EventPong = "pong"
	// Fired on disconnection
	// The error provided in disconnection event
	// is defined in RFC 6455, section 11.7.
	// @url https://github.com/gofiber/websocket/blob/cd4720c435de415b864d975a9ca23a47eaf081ef/websocket.go#L192
	EventDisconnect = "disconnect"
	// Fired on first connection
	EventConnect = "connect"
	// Fired when the connection is actively closed from the server
	EventClose = "close"
	// Fired when some error appears useful also for debugging websockets
	EventError = "error"
)
```
Event Payload object
```go
// Event Payload is the object that
// stores all the information about the event and
// the connection
type EventPayload struct {
	// The connection object
	Kws *Websocket
	// The name of the event
	Name string
	// Unique connection UUID
	SocketUUID string
	// Optional websocket attributes
	SocketAttributes map[string]string
	// Optional error when are fired events like
	// - Disconnect
	// - Error
	Error error
	// Data is used on Message and on Error event
	Data []byte
}
```
---


```go
// Set a specific attribute for the specific socket connection
func (kws *Websocket) SetAttribute(key string, attribute string)
```
---


```go
// Get a specific attribute from the socket attributes
func (kws *Websocket) GetAttribute(key string) string
```
---


```go
// Emit the message to a specific socket uuids list
func (kws *Websocket) EmitToList(uuids []string, message []byte) 
```
---

```go
// Emit to a specific socket connection
func (kws *Websocket) EmitTo(uuid string, message []byte) error
```
---


```go
// Broadcast to all the active connections
// except avoid broadcasting the message to itself
func (kws *Websocket) Broadcast(message []byte, except bool)
```
---


```go
// Emit/Write the message into the given connection
func (kws *Websocket) Emit(message []byte)
```
---


```go
// Actively close the connection from the server
func (kws *Websocket) Close() 
```
---

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

		// Unmarshal the json message
		// {
		//  "from": "<user-id>",
		//  "to": "<recipient-user-id>",
		//  "data": "hello"
		//}
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

	// On close event
	// This event is called when the server disconnects the user actively with .Close() method
	ikisocket.On(ikisocket.EventClose, func(ep *ikisocket.EventPayload) {
		// Remove the user from the local clients
		delete(clients, ep.SocketAttributes["user_id"])
		fmt.Println(fmt.Sprintf("Close event - User: %s", ep.SocketAttributes["user_id"]))
	})

	// On error event
	ikisocket.On(ikisocket.EventError, func(ep *ikisocket.EventPayload) {
		fmt.Println(fmt.Sprintf("Error event - User: %s", ep.SocketAttributes["user_id"]))
	})

	app.Get("/ws/:id", ikisocket.New(func(kws *ikisocket.Websocket) {
		
		// Retrieve the user id from endpoint
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
### Message object example

```
{
  "from": "<user-id>",
  "to": "<recipient-user-id>",
  "data": "hello"
}
```







