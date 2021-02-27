package ikisocket_chat_room_example

import (
	"encoding/json"
	"fmt"
	"github.com/antoniodipinto/ikisocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"log"
	"math/rand"
	"time"
)

// Basic chat message object
type MessageObject struct {
	Data string `json:"data"`
	From string `json:"from"`
	Room string `json:"room"`
	To   string `json:"to"`
}

// Chat Room message object
type Room struct {
	Name  string   `json:"name"`
	UUID  string   `json:"uuid"`
	Users []string `json:"users"`
}

var clients map[string]string
var rooms map[string]*Room

func main() {

	// The key for the map is message.to
	clients = make(map[string]string)

	// Rooms will be kept in memory as map
	// for faster and easier handling
	rooms = make(map[string]*Room)

	// Start a new Fiber application
	app := fiber.New()

	// Get all the available rooms
	app.Get("/rooms", func(ctx *fiber.Ctx) error {
		return ctx.JSON(beautifyRoomsObject(rooms))
	})

	// Create a Room passing the name
	app.Post("/rooms/create", func(ctx *fiber.Ctx) error {
		type request struct {
			Name string `json:"name"`
		}

		// check for the incoming request body
		body := new(request)
		if err := ctx.BodyParser(&body); err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "cannot parse JSON",
			})
		}

		if body.Name == "" {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid room name",
			})
		}

		room := Room{
			Name:  body.Name,
			UUID:  generateRoomId(),
			Users: nil,
		}

		rooms[room.UUID] = &room
		return ctx.JSON(room)
	})

	// Delete a room by his ID
	app.Delete("/rooms/delete/:id", func(ctx *fiber.Ctx) error {
		roomId := ctx.Params("id")
		delete(rooms, roomId)
		return ctx.JSON(true)
	})

	// Joins a room passing:
	// - Room ID (as room)
	// - User ID (as user)
	app.Post("/rooms/join", func(ctx *fiber.Ctx) error {
		type request struct {
			Room string `json:"room"`
			User string `json:"user"`
		}

		// check for the incoming request body
		body := new(request)
		if err := ctx.BodyParser(&body); err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "cannot parse JSON",
			})
		}

		if body.User == "" || body.Room == "" {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid user or room",
			})
		}

		rooms[body.Room].Users = append(rooms[body.Room].Users, body.User)
		return ctx.JSON(true)
	})

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

	// Pull out in another function
	// all the ikisocket callbacks and listeners
	setupSocketListeners()

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

// random room id generator
func generateRoomId() string {
	length := 100
	seed := rand.New(rand.NewSource(time.Now().UnixNano()))
	charset := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghijklmnopqrstuvwxyz"

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seed.Intn(len(charset))]
	}

	return string(b)
}

// for beautify purposes we clean the rooms object to a list of rooms
func beautifyRoomsObject(rooms map[string]*Room) []*Room {
	var result []*Room

	for _, room := range rooms {
		result = append(result, room)
	}

	return result
}

// Setup all the ikisocket listeners
// pulled out main function
func setupSocketListeners() {

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
		//  "room": "<room-id>",
		//  "data": "hello"
		//}
		err := json.Unmarshal(ep.Data, &message)
		if err != nil {
			fmt.Println(err)
			return
		}

		// If the user is trying to send message
		// into a specific group, iterate over the
		// group user socket UUIDs
		if message.Room != "" {
			// Emit the message to all the room participants
			// iterating on all the uuids
			for _, userId := range rooms[message.Room].Users {
				_ = ep.Kws.EmitTo(clients[userId], ep.Data)
			}

			// Other way can be used EmitToList method
			// if you have a []string of ikisocket uuids
			//
			// ep.Kws.EmitToList(list, data)
			//
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

}
