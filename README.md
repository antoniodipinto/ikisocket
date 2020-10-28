
# WebSocket wrapper for [Fiber](https://github.com/gofiber/fiber) with events support
[![Go Report Card](https://goreportcard.com/badge/github.com/antoniodipinto/ikisocket)](https://goreportcard.com/report/github.com/antoniodipinto/ikisocket)
[![GoDoc](https://godoc.org/github.com/antoniodipinto/ikisocket?status.svg)](https://godoc.org/github.com/antoniodipinto/ikisocket)
[![GitHub license](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/antoniodipinto/ikisocket/blob/master/LICENSE)
### Based on [Fiber Websocket](https://github.com/gofiber/websocket) and inspired by [Socket.io](https://github.com/socketio/socket.io)

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
	ikisocket "github.com/antoniodipinto/ikisocket"
	"github.com/gofiber/fiber/v2"
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
    	app.Use(func(c *fiber.Ctx) {
    		c.Locals("user_id", c.Query("user_id"))
    		c.Next()
    	})
    
    	// Multiple event handling supported
    	ikisocket.On(ikisocket.EventConnect, func(ep *ikisocket.EventPayload) {
    		fmt.Println("fired connect 1")
    	})
    
    	ikisocket.On(ikisocket.EventConnect, func(ep *ikisocket.EventPayload) {
    		fmt.Println("fired connect 2")
    	})
    
    	ikisocket.On(ikisocket.EventMessage, func(ep *ikisocket.EventPayload) {
    		fmt.Println("fired message: " + string(ep.Data))
    	})
    
    	ikisocket.On(ikisocket.EventDisconnect, func(ep *ikisocket.EventPayload) {
    		fmt.Println("fired disconnect" + ep.Error.Error())
    	})
    
    	// Websocket route init
    	app.Get("/ws", ikisocket.New(func(kws *ikisocket.Websocket) {
    		// Retrieve user id from the middleware (optional)
    		userId := fmt.Sprintf("%v", kws.Locals("user_id"))
    
    		// Every websocket connection has an optional session key => value storage
    		kws.SetAttribute("user_id", userId)
    
    		// On connect event. Notify when comes a new connection
    		kws.OnConnect = func() {
    			// Add the connection to the list of the connected clients
    			// The UUID is generated randomly
    			clients[userId] = kws.UUID
    			//Broadcast to all the connected users the newcomer
    			kws.Broadcast([]byte("New user connected: "+userId+" and UUID: "+kws.UUID), true)
    			//Write welcome message
    			kws.Emit([]byte("Hello user: " + userId + " and UUID: " + kws.UUID))
    		}
    
    		// On message event
    		kws.OnMessage = func(data []byte) {
    
    			message := MessageObject{}
    			json.Unmarshal(data, &message)
    			// Emit the message directly to specified user
    			err := kws.EmitTo(clients[message.To], data)
    			if err != nil {
    				fmt.Println(err)
    			}
    		}
    	}))
    
    	ikisocket.On("close", func(payload *ikisocket.EventPayload) {
    		fmt.Println("fired close " + payload.SocketAttributes["user_id"])
    	})
    
    	// Start the application on port 3000
    	app.Listen(3000)
}
```
### Connect to the websocket

```
ws://localhost:3000/ws?user_id=54s5f18d1h8d1h8f
```
