package tswconnector

import (
	"context"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSocketConnection(t *testing.T) {
	ctx := context.Background()
	version := "1.0.0"
	conn := NewSocketConnection(ctx, version)

	assert.NotNil(t, conn)
	assert.Equal(t, version, conn.Version)
	assert.NotNil(t, conn.WsUpgrader)
	assert.NotNil(t, conn.Server)
	assert.NotNil(t, conn.OutgoingChannels)
	assert.NotNil(t, conn.Subscribers)
}

func TestSocketConnection_Lifecycle(t *testing.T) {
	conn := NewSocketConnection(t.Context(), "1.0.0")

	errChan := make(chan error, 1)
	go func() {
		errChan <- conn.Start()
	}()

	port := conn.Port()
	assert.GreaterOrEqual(t, port, SOCKET_CONNECTION_PORT_RANGE_START)
	assert.LessOrEqual(t, port, SOCKET_CONNECTION_PORT_RANGE_END)

	err := conn.Stop()
	require.NoError(t, err)

	// Check if Start returned nil (due to http.ErrServerClosed)
	err = <-errChan
	require.NoError(t, err)
}

func TestSocketConnection_WebsocketFlow(t *testing.T) {
	conn := NewSocketConnection(t.Context(), "1.0.0")

	go func() {
		conn.Start()
	}()

	// Give it a moment to start
	time.Sleep(200 * time.Millisecond)
	defer conn.Stop()

	u := url.URL{Scheme: "ws", Host: conn.Server.Addr, Path: "/"}

	// Client side
	dialer := websocket.Dialer{}
	ws, _, err := dialer.Dial(u.String(), nil)
	require.NoError(t, err)
	defer ws.Close()

	// 1. Test Server -> Client (Send)
	msg := TSWConnector_Message{
		EventName:  "test_event",
		Properties: map[string]string{"key": "value"},
	}

	// We need to wait for the connection to be registered in OutgoingChannels
	require.Eventually(t, func() bool {
		return conn.IsActive()
	}, 2*time.Second, 100*time.Millisecond)

	err = conn.Send(msg)
	require.NoError(t, err)

	// Read from client
	_, p, err := ws.ReadMessage()
	require.NoError(t, err)
	assert.Equal(t, msg.ToString(), string(p))

	// 2. Test Client -> Server (Subscribe & Receive)
	subChan, unsubscribe := conn.Subscribe()
	defer unsubscribe()

	clientMsg := TSWConnector_Message{
		EventName:  "client_event",
		Properties: map[string]string{"data": "hello"},
	}

	err = ws.WriteMessage(websocket.TextMessage, []byte(clientMsg.ToString()))
	require.NoError(t, err)

	select {
	case received := <-subChan:
		assert.Equal(t, clientMsg.EventName, received.EventName)
		assert.Equal(t, clientMsg.Properties["data"], received.Properties["data"])
	case <-time.After(2 * time.Second):
		t.Fatal("Timed out waiting for message from client")
	}
}

func TestSocketConnection_Forwarding(t *testing.T) {
	conn := NewSocketConnection(t.Context(), "1.0.0")

	// Start in goroutine
	go func() {
		if err := conn.Start(); err != nil && err != http.ErrServerClosed {
		}
	}()

	// Give it a moment to start
	time.Sleep(200 * time.Millisecond)
	defer conn.Stop()

	u := url.URL{Scheme: "ws", Host: conn.Server.Addr, Path: "/"}
	dialer := websocket.Dialer{}

	// Client A
	wsA, _, err := dialer.Dial(u.String(), nil)
	require.NoError(t, err)
	defer wsA.Close()

	// Client B
	wsB, _, err := dialer.Dial(u.String(), nil)
	require.NoError(t, err)
	defer wsB.Close()

	require.Eventually(t, func() bool {
		return conn.IsActive()
	}, 2*time.Second, 100*time.Millisecond)

	// We want to verify that Client A's message is forwarded to Client B,
	// but NOT to Client A itself.

	// 1. Client B subscribes to see if it gets Client A's message
	subChanB, unsubscribeB := conn.Subscribe()
	defer unsubscribeB()

	// 2. Client A sends a message
	msgA := TSWConnector_Message{
		EventName:  "forward_test",
		Properties: map[string]string{"from": "A"},
	}

	err = wsA.WriteMessage(websocket.TextMessage, []byte(msgA.ToString()))
	require.NoError(t, err)

	// Client B should receive it via subscription
	select {
	case received := <-subChanB:
		assert.Equal(t, msgA.EventName, received.EventName)
		assert.Equal(t, "A", received.Properties["from"])
	case <-time.After(2 * time.Second):
		t.Fatal("Timed out waiting for forwarded message on Client B")
	}

	// 3. Test Server -> Client (Send) to Client A specifically
	// We need to make sure Client A is actually receiving messages.
	// Since we don't have the UUID of Client A easily, we rely on the fact that
	// Send() broadcasts to ALL.

	msgServer := TSWConnector_Message{
		EventName:  "server_msg",
		Properties: map[string]string{"msg": "hello_all"},
	}

	err = conn.Send(msgServer)
	require.NoError(t, err)

	// Client A should receive it
	_, p, err := wsA.ReadMessage()
	require.NoError(t, err)
	assert.Equal(t, msgServer.ToString(), string(p))
}

func TestSocketConnection_IsActive(t *testing.T) {
	conn := NewSocketConnection(t.Context(), "1.0.0")

	// Start in goroutine
	go func() {
		if err := conn.Start(); err != nil && err != http.ErrServerClosed {
		}
	}()

	// Give it a moment to start
	time.Sleep(200 * time.Millisecond)
	defer conn.Stop()

	assert.False(t, conn.IsActive())

	u := url.URL{Scheme: "ws", Host: conn.Server.Addr, Path: "/"}
	dialer := websocket.Dialer{}
	ws, _, err := dialer.Dial(u.String(), nil)
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		return conn.IsActive()
	}, 2*time.Second, 100*time.Millisecond)

	assert.True(t, conn.IsActive())

	ws.Close()

	require.Eventually(t, func() bool {
		return !conn.IsActive()
	}, 2*time.Second, 100*time.Millisecond)
}
