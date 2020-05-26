package ikisocket

import (
	"errors"
	"github.com/gofiber/fiber"
	"github.com/gofiber/websocket"
	"math/rand"
	"time"
)

// Source @url:https://github.com/gorilla/websocket/blob/master/conn.go#L61
// The message types are defined in RFC 6455, section 11.8.
const (
	// TextMessage denotes a text data message. The text message payload is
	// interpreted as UTF-8 encoded text data.
	TextMessage = 1

	// BinaryMessage denotes a binary data message.
	BinaryMessage = 2

	// CloseMessage denotes a close control message. The optional message
	// payload contains a numeric code and text. Use the FormatCloseMessage
	// function to format a close message payload.
	CloseMessage = 8

	// PingMessage denotes a ping control message. The optional message payload
	// is UTF-8 encoded text.
	PingMessage = 9

	// PongMessage denotes a pong control message. The optional message payload
	// is UTF-8 encoded text.
	PongMessage = 10
)

const (
	EventMessage = "message"

	EventPing = "ping"

	EventDisconnect = "disconnect"

	EventConnect = "connect"

	EventError = "error"
)

type message struct {
	mType int

	data []byte
}

type EventPayload struct {
	Kws Websocket

	Name string

	SocketUUID string

	SocketAttributes map[string]string

	Error error

	Data []byte
}

type Websocket struct {
	ws *websocket.Conn

	isAlive bool

	queue []message

	attributes map[string]string

	UUID string

	Locals func(key string) interface{}

	OnConnect func()

	OnMessage func(data []byte)

	OnDisconnect func()
}

// Pool with the active connections
var pool = make(map[string]*Websocket)

var listeners = make(map[string][]func(payload *EventPayload))

func New(callback func(kws *Websocket)) func(*fiber.Ctx) {
	return websocket.New(func(c *websocket.Conn) {

		kws := &Websocket{
			ws: c,
			Locals: func(key string) interface{} {
				return c.Locals(key)
			},
			queue:      nil,
			attributes: make(map[string]string),
			isAlive:    true,
		}

		//Generate uuid
		kws.UUID = kws.newUUID()

		pool[kws.UUID] = kws

		callback(kws)

		if kws.OnConnect != nil {
			kws.fireEvent(EventConnect, nil, nil)
			kws.OnConnect()
		}

		kws.run()
	})
}

func (kws *Websocket) SetAttribute(key string, attribute string) {
	kws.attributes[key] = attribute
}

func (kws *Websocket) GetAttribute(key string) string {
	return kws.attributes[key]
}

func (kws *Websocket) EmitTo(uuid string, message []byte) error {

	if !isValidUUID(uuid) {
		return errors.New("invalid UUID")
	}

	if !pool[uuid].isAlive {
		err := errors.New("message cannot be delivered. Socket disconnected")
		kws.fireEvent(EventError, nil, err)
		return err
	}
	pool[uuid].Emit(message)
	return nil
}

func (kws *Websocket) Broadcast(message []byte, except bool) {

	for uuid, _ := range pool {
		if except && kws.UUID == uuid {
			continue
		}

		kws.EmitTo(uuid, message)
	}
}

func (kws *Websocket) Emit(message []byte) {
	kws.write(TextMessage, message)
}

// pong writes a control message to the client
func (kws *Websocket) pong() {
	for range time.Tick(5 * time.Second) {
		kws.write(PongMessage, []byte{})
	}
}

func (kws *Websocket) write(messageType int, messageBytes []byte) {
	kws.queue = append(kws.queue, message{
		mType: messageType,
		data:  messageBytes,
	})
}

func (kws *Websocket) run() {

	go kws.pong()
	go kws.read()

	for range time.Tick(1 * time.Millisecond) {
		if !kws.isAlive {
			break
		}
		if len(kws.queue) == 0 {
			continue
		}
		for _, message := range kws.queue {
			err := kws.ws.WriteMessage(message.mType, message.data)
			if err != nil {
				kws.disconnected(err)
			}
		}
		kws.queue = nil
	}
}
func (kws *Websocket) read() {
	for range time.Tick(10 * time.Millisecond) {
		if !kws.isAlive {
			break
		}
		mtype, msg, err := kws.ws.ReadMessage()

		if mtype == PingMessage {
			kws.fireEvent(EventPing, nil, nil)
			continue
		}

		if mtype == CloseMessage {
			kws.disconnected(nil)
			break
		}

		if err == nil {
			kws.fireEvent(EventMessage, msg, nil)

			if kws.OnMessage != nil {
				kws.OnMessage(msg)
			}
		} else {
			kws.disconnected(err)
		}
	}
}

// When the connection closes, disconnected method
// handle also the OnDisconnect() event
func (kws *Websocket) disconnected(err error) {
	kws.fireEvent(EventDisconnect, nil, err)
	kws.isAlive = false

	// Remove the socket from the pool
	delete(pool, kws.UUID)

	// Close the connection from the server side
	kws.ws.Close()

	if kws.OnDisconnect != nil {
		kws.OnDisconnect()
	}
}

func (kws *Websocket) newUUID() string {

	length := 32
	seed := rand.New(rand.NewSource(time.Now().UnixNano()))
	charset := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghijklmnopqrstuvwxyz"

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seed.Intn(len(charset))]
	}
	return string(b)
}

func (kws *Websocket) fireEvent(event string, data []byte, error error) {
	callbacks, ok := listeners[event]

	if ok {
		for _, callback := range callbacks {
			callback(&EventPayload{
				Kws:              *kws,
				Name:             event,
				SocketUUID:       kws.UUID,
				SocketAttributes: kws.attributes,
				Data:             data,
				Error:            error,
			})
		}
	}
}

func On(event string, callback func(payload *EventPayload)) {
	listeners[event] = append(listeners[event], callback)
}

func isValidUUID(uuid string) bool {
	_, ok := pool[uuid]

	return ok
}
