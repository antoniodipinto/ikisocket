package ikisocket

import (
	"github.com/gofiber/websocket/v2"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
)

const numTestConn = 10

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

func TestGlobalFire(t *testing.T) {
	reset()

	// simulate connections
	for i := 0; i < numTestConn; i++ {
		kws := createWS()
		pool[kws.UUID] = kws
	}

	h := new(HandlerMock)
	// setup expectations
	h.On("OnCustomEvent", mock.Anything).Return(nil)

	// register custom event handler
	On("customevent", h.OnCustomEvent)

	h.wg.Add(numTestConn)

	// fire global custom event on all connections
	Fire("customevent", []byte("test"))

	h.wg.Wait()

	h.AssertNumberOfCalls(t, "OnCustomEvent", numTestConn)
}

func TestGlobalBroadcast(t *testing.T) {
	reset()

	mws := new(WebsocketMock)
	mws.UUID = "80a80sdf809dsf"
	pool[mws.UUID] = mws

	// setup expectations
	mws.On("Emit", mock.Anything).Return(nil)

	mws.wg.Add(1)

	// send global broadcast to all connections
	Broadcast([]byte("test"))

	mws.wg.Wait()

	mws.AssertNumberOfCalls(t, "Emit", 1)
}

func TestGlobalEmitTo(t *testing.T) {
	reset()

	aliveUUID := "80a80sdf809dsf"
	closedUUID := "las3dfj09808"

	alive := new(WebsocketMock)
	alive.UUID = aliveUUID
	pool[alive.UUID] = alive

	closed := new(WebsocketMock)
	closed.UUID = closedUUID
	pool[closed.UUID] = closed

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
	reset()

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
		pool[kws.UUID] = kws
	}

	// send global broadcast to all connections
	EmitToList(uuids, []byte("test"))

	for _, kws := range pool {
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
		queue:      make(map[string]message),
		attributes: make(map[string]string),
		isAlive:    true,
	}

	kws.UUID = kws.createUUID()

	return kws
}

func reset() {
	pool = make(map[string]ws)
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

func (s *WebsocketMock) pong() {
	panic("implement me")
}

func (s *WebsocketMock) write(_ int, _ []byte) {
	panic("implement me")
}

func (s *WebsocketMock) run() {
	panic("implement me")
}

func (s *WebsocketMock) read() {
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
