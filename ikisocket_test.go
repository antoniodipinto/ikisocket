package ikisocket

import (
	"context"
	"github.com/fasthttp/websocket"
	"github.com/gofiber/fiber/v2"
	fws "github.com/gofiber/websocket/v2"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
)

const numTestConn = 10
const numParallelTestConn = 5000

type HandlerMock struct {
	mock.Mock
	wg sync.WaitGroup
}

type WebsocketMock struct {
	mock.Mock
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

type testHandler struct {
	app *fiber.App
}

func (h *testHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	resp, err := h.app.Test(r)
	if err != nil {
		panic(err)
	}
	for name, values := range resp.Header {
		w.Header()[name] = values
	}
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
	_ = resp.Body.Close()
}

func TestParallelConnections(t *testing.T) {
	app := fiber.New()

	app.Use(upgradeMiddleware)

	wg := sync.WaitGroup{}
	On(EventConnect, func(payload *EventPayload) {
		// wg.Done()
		payload.Kws.Emit([]byte("response"))
		log.Println("hi")
	})
	On(EventMessage, func(payload *EventPayload) {
		log.Println(payload.Data)
	})
	On(EventClose, func(payload *EventPayload) {
		log.Println("Close")
	})
	On(EventDisconnect, func(payload *EventPayload) {
		log.Println("Disconnect")
	})

	app.Get("/", New(func(kws *Websocket) {
		// log.Println("conn")

	}))

	s := httptest.NewServer(&testHandler{
		app,
	})
	// s := httptest.NewServer(adaptor.FiberHandler(New(func(kws *Websocket) {
	// 	log.Println("test")
	//
	// 	On(EventConnect, func(payload *EventPayload) {
	// 		log.Println("jep")
	// 		payload.Kws.Emit([]byte("response"))
	// 		wg.Done()
	// 	})
	// })))

	wsURL := httpToWs(t, s.URL)

	defer s.Close()

	// req := httptest.NewRequest(fiber.MethodGet, "/", nil)

	// req := httptest.NewRequest(fiber.MethodGet, "/", nil)
	// req.Header.Set("Connection", "Upgrade")
	// req.Header.Set("Upgrade", "Websocket")
	// req.Header.Set("Sec-WebSocket-Key", "veQ+5bJcQAhyLAn+SnM5YA==")
	// req.Header.Set("Sec-WebSocket-Version", "13")
	for i := 0; i < numParallelTestConn; i++ {
		wg.Add(1)
		// resp, err := app.Test(req, -1)
		// require.Nil(t, err)
		// require.Equal(t, fiber.StatusSwitchingProtocols, resp.StatusCode)
		go func() {
			ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
			if err != nil {
				t.Fatal(err)
			}

			// if err := ws.WriteMessage(websocket.TextMessage, []byte("test")); err != nil {
			// 	t.Fatal(err)
			// }

			// time.Sleep(1 * time.Second)

			tp, m, err := ws.ReadMessage()
			if err != nil {
				t.Fatal(err)
			}
			require.Equal(t, TextMessage, tp)
			require.Equal(t, "response", m)
			wg.Done()

			if err := ws.Close(); err != nil {
				t.Fatal(err)
			}
		}()
	}
	wg.Wait()
	// time.Sleep(500 * time.Millisecond)
}

func httpToWs(t *testing.T, u string) string {
	wsURL, err := url.Parse(u)
	if err != nil {
		t.Fatal(err)
	}

	switch wsURL.Scheme {
	case "http":
		wsURL.Scheme = "ws"
	case "https":
		wsURL.Scheme = "wss"
	}

	return wsURL.String()
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
		mws.UUID = "80a80sdf809dsf"
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
		kws.UUID = uuid
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
		attributes: make(map[string]string),
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

func (s *WebsocketMock) SetAttribute(_ string, _ string) {
	panic("implement me")
}

func (s *WebsocketMock) GetAttribute(_ string) string {
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
	panic("implement me")
}

func (s *WebsocketMock) randomUUID() string {
	panic("implement me")
}

func (s *WebsocketMock) fireEvent(_ string, _ []byte, _ error) {
	panic("implement me")
}
