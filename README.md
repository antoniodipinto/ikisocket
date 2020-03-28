# 🧬 WebSocket wrapper for [Fiber](https://github.com/gofiber/fiber) with events

Based on [Fiber Websocket](https://github.com/gofiber/websocket)

### Install

```
go get -u github.com/antoniodipinto/ikisocket
```

### Basic chat example

```go
package main

import (
	"encoding/json"
	"fmt"
	ikisocket "github.com/antoniodipinto/ikisocket"
	"github.com/gofiber/fiber"
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

	// Websocket route init
	app.Get("/ws", ikisocket.New(func(kws *ikisocket.Websocket) {
		// Retrieve user id from the middleware (optional)
		userId := fmt.Sprintf("%v", kws.Locals("user_id"))
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

		kws.OnDisconnect = func() {
			fmt.Println("User disconnected: " + userId)
		}
	}))

	// Start the application on port 3000
	app.Listen(3000)
}

```