package ikisocket

import (
	"context"
	"github.com/fasthttp/websocket"
	"github.com/gofiber/fiber/v2"
	fws "github.com/gofiber/websocket/v2"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp/fasthttputil"
	"math/rand"
	"net"
	"sync"
	"testing"
	"time"
)

const numTestConn = 10
const numParallelTestConn = 5_000

type HandlerMock struct {
	mock.Mock
	wg sync.WaitGroup
}

type WebsocketMock struct {
	mock.Mock
	mu         sync.RWMutex
	wg         sync.WaitGroup
	ws         *websocket.Conn
	isAlive    bool
	queue      map[string]message
	attributes map[string]string
	UUID       string
	Locals     func(key string) interface{}
	Params     func(key string, defaultValue ...string) string
	Query      func(key string, defaultValue ...string) string
	Cookies    func(key string, defaultValue ...string) string
}

func (s *WebsocketMock) SetUUID(uuid string) {

	s.mu.Lock()
	defer s.mu.Unlock()

	if pool.contains(uuid) {
		panic(ErrorUUIDDuplication)
	}
	s.UUID = uuid
}

func (s *WebsocketMock) GetIntAttribute(key string) int {
	panic("implement me")
}

func (s *WebsocketMock) GetStringAttribute(key string) string {
	panic("implement me")
}

func (h *HandlerMock) OnCustomEvent(payload *EventPayload) {
	h.Called(payload)
	h.wg.Done()
}

func (s *WebsocketMock) Emit(message []byte) {
	s.Called(message)
	s.wg.Done()
}

func (s *WebsocketMock) IsAlive() bool {
	args := s.Called()
	return args.Bool(0)
}

func (s *WebsocketMock) GetUUID() string {
	return s.UUID
}

func TestParallelConnections(t *testing.T) {
	pool.reset()

	// create test server
	cfg := fiber.Config{
		DisableStartupMessage: true,
	}
	app := fiber.New(cfg)
	ln := fasthttputil.NewInmemoryListener()
	wg := sync.WaitGroup{}

	defer func() {
		_ = app.Shutdown()
		_ = ln.Close()
	}()

	// attach upgrade middleware
	app.Use(upgradeMiddleware)

	// send back response on correct message
	On(EventMessage, func(payload *EventPayload) {
		if string(payload.Data) == "test" {
			payload.Kws.Emit([]byte("response"))
		}
	})

	// create websocket endpoint
	app.Get("/", New(func(kws *Websocket) {
	}))

	// start server
	go func() {
		_ = app.Listener(ln)
	}()

	wsURL := "ws://" + ln.Addr().String()

	// create concurrent connections
	for i := 0; i < numParallelTestConn; i++ {
		wg.Add(1)
		go func() {
			dialer := &websocket.Dialer{
				NetDial: func(network, addr string) (net.Conn, error) {
					return ln.Dial()
				},
				HandshakeTimeout: 45 * time.Second,
			}
			ws, _, err := dialer.Dial(wsURL, nil)
			if err != nil {
				t.Error(err)
				return
			}

			if err := ws.WriteMessage(websocket.TextMessage, []byte("test")); err != nil {
				t.Error(err)
				return
			}

			tp, m, err := ws.ReadMessage()
			if err != nil {
				t.Error(err)
				return
			}
			require.Equal(t, TextMessage, tp)
			require.Equal(t, "response", string(m))
			wg.Done()

			if err := ws.Close(); err != nil {
				t.Error(err)
				return
			}
		}()
	}
	wg.Wait()
}

func TestGlobalFire(t *testing.T) {
	pool.reset()

	// simulate connections
	for i := 0; i < numTestConn; i++ {
		kws := createWS()
		pool.set(kws)
	}

	h := new(HandlerMock)
	// setup expectations
	h.On("OnCustomEvent", mock.Anything).Return(nil)

	// Moved before registration of the event
	// if after can cause: panic: sync: negative WaitGroup counter
	h.wg.Add(numTestConn)

	// register custom event handler
	On("customevent", h.OnCustomEvent)

	// fire global custom event on all connections
	Fire("customevent", []byte("test"))

	h.wg.Wait()

	h.AssertNumberOfCalls(t, "OnCustomEvent", numTestConn)
}

func TestGlobalBroadcast(t *testing.T) {
	pool.reset()

	for i := 0; i < numParallelTestConn; i++ {
		mws := new(WebsocketMock)
		mws.SetUUID(mws.createUUID())
		pool.set(mws)

		// setup expectations
		mws.On("Emit", mock.Anything).Return(nil)

		mws.wg.Add(1)
	}

	// send global broadcast to all connections
	Broadcast([]byte("test"))

	for _, mws := range pool.all() {
		mws.(*WebsocketMock).wg.Wait()
		mws.(*WebsocketMock).AssertNumberOfCalls(t, "Emit", 1)
	}

}

func TestGlobalEmitTo(t *testing.T) {
	pool.reset()

	aliveUUID := "80a80sdf809dsf"
	closedUUID := "las3dfj09808"

	alive := new(WebsocketMock)
	alive.UUID = aliveUUID
	pool.set(alive)

	closed := new(WebsocketMock)
	closed.UUID = closedUUID
	pool.set(closed)

	// setup expectations
	alive.On("Emit", mock.Anything).Return(nil)
	alive.On("IsAlive").Return(true)
	closed.On("IsAlive").Return(false)

	var err error
	err = EmitTo("non-existent", []byte("error"))
	require.Equal(t, ErrorInvalidConnection, err)

	err = EmitTo(closedUUID, []byte("error"))
	require.Equal(t, ErrorInvalidConnection, err)

	alive.wg.Add(1)

	// send global broadcast to all connections
	err = EmitTo(aliveUUID, []byte("test"))
	require.Nil(t, err)

	alive.wg.Wait()

	alive.AssertNumberOfCalls(t, "Emit", 1)
}

func TestGlobalEmitToList(t *testing.T) {
	pool.reset()

	uuids := []string{
		"80a80sdf809dsf",
		"las3dfj09808",
	}

	for _, uuid := range uuids {
		kws := new(WebsocketMock)
		kws.SetUUID(uuid)
		kws.On("Emit", mock.Anything).Return(nil)
		kws.On("IsAlive").Return(true)
		kws.wg.Add(1)
		pool.set(kws)
	}

	// send global broadcast to all connections
	EmitToList(uuids, []byte("test"))

	for _, kws := range pool.all() {
		kws.(*WebsocketMock).wg.Wait()
		kws.(*WebsocketMock).AssertNumberOfCalls(t, "Emit", 1)
	}
}

func TestWebsocket_GetIntAttribute(t *testing.T) {
	kws := &Websocket{
		attributes: make(map[string]interface{}),
	}

	// get unset attribute
	// Will return null without panicking

	// get non-int attribute
	// Will return 0 without panicking
	kws.SetAttribute("notInt", "")

	// get int attribute
	kws.SetAttribute("int", 3)
	v := kws.GetIntAttribute("int")
	require.Equal(t, 3, v)
}

func TestWebsocket_GetStringAttribute(t *testing.T) {
	kws := &Websocket{
		attributes: make(map[string]interface{}),
	}

	// get unset attribute

	// get non-string attribute
	kws.SetAttribute("notString", 3)

	// get string attribute
	kws.SetAttribute("str", "3")
	v := kws.GetStringAttribute("str")
	require.Equal(t, "3", v)
}

func assertPanic(t *testing.T, f func()) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	f()
}

func createWS() *Websocket {
	kws := &Websocket{
		ws: nil,
		Locals: func(key string) interface{} {
			return ""
		},
		Params: func(key string, defaultValue ...string) string {
			return ""
		},
		Query: func(key string, defaultValue ...string) string {
			return ""
		},
		Cookies: func(key string, defaultValue ...string) string {
			return ""
		},
		queue:      make(chan message),
		attributes: make(map[string]interface{}),
		isAlive:    true,
	}

	kws.UUID = kws.createUUID()

	return kws
}

func upgradeMiddleware(c *fiber.Ctx) error {
	// IsWebSocketUpgrade returns true if the client
	// requested upgrade to the WebSocket protocol.
	if fws.IsWebSocketUpgrade(c) {
		c.Locals("allowed", true)
		return c.Next()
	}
	return fiber.ErrUpgradeRequired
}

//
// needed but not used
//

func (s *WebsocketMock) SetAttribute(_ string, _ interface{}) {
	panic("implement me")
}

func (s *WebsocketMock) GetAttribute(_ string) interface{} {
	panic("implement me")
}

func (s *WebsocketMock) EmitToList(_ []string, _ []byte) {
	panic("implement me")
}

func (s *WebsocketMock) EmitTo(_ string, _ []byte) error {
	panic("implement me")
}

func (s *WebsocketMock) Broadcast(_ []byte, _ bool) {
	panic("implement me")
}

func (s *WebsocketMock) Fire(_ string, _ []byte) {
	panic("implement me")
}

func (s *WebsocketMock) Close() {
	panic("implement me")
}

func (s *WebsocketMock) pong(_ context.Context) {
	panic("implement me")
}

func (s *WebsocketMock) write(_ int, _ []byte) {
	panic("implement me")
}

func (s *WebsocketMock) run() {
	panic("implement me")
}

func (s *WebsocketMock) read(_ context.Context) {
	panic("implement me")
}

func (s *WebsocketMock) disconnected(_ error) {
	panic("implement me")
}

func (s *WebsocketMock) createUUID() string {
	return s.randomUUID()
}

func (s *WebsocketMock) randomUUID() string {
	length := 100
	seed := rand.New(rand.NewSource(time.Now().UnixNano()))
	charset := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghijklmnopqrstuvwxyz"

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seed.Intn(len(charset))]
	}

	return string(b)
}

func (s *WebsocketMock) fireEvent(_ string, _ []byte, _ error) {
	panic("implement me")
}
