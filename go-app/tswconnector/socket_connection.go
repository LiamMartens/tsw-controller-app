package tswconnector

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"
	"tsw_controller_app/chan_utils"
	"tsw_controller_app/logger"
	"tsw_controller_app/map_utils"
	"tsw_controller_app/pubsub_utils"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const SOCKET_CONNECTION_OUTGOING_QUEUE_BUFFER_SIZE = 32
const SOCKET_CONNECTION_PORT_RANGE_START = 63241
const SOCKET_CONNECTION_PORT_RANGE_END = 63242

type SocketConnection struct {
	WsUpgrader       *websocket.Upgrader
	Server           *http.Server
	OutgoingChannels *map_utils.LockMap[uuid.UUID, chan TSWConnector_Message]
	Subscribers      *pubsub_utils.PubSubSlice[TSWConnector_Message]
}

var _ TSWConnector = (*SocketConnection)(nil)

func (c *SocketConnection) WebsocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := c.WsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Logger.Error("[SocketConnection::WebsocketHandler] websocket upgrade error", "error", err.Error())
		return
	}
	defer conn.Close()

	conn_id := uuid.New()
	outgoing_channel := make(chan TSWConnector_Message, SOCKET_CONNECTION_OUTGOING_QUEUE_BUFFER_SIZE)
	c.OutgoingChannels.Set(conn_id, outgoing_channel)
	defer close(outgoing_channel)
	defer c.OutgoingChannels.Delete(conn_id)

	ctx_with_cancel, cancel_sender := context.WithCancel(r.Context())
	go func() {
		for {
			select {
			case <-ctx_with_cancel.Done():
				return
			case message := <-outgoing_channel:
				err := conn.WriteMessage(websocket.TextMessage, []byte(message.ToString()))
				if err != nil {
					cancel_sender()
					return
				}
			}
		}
	}()

	for {
		msg_type, msg, err := conn.ReadMessage()
		if err != nil {
			logger.Logger.Error("[ProfileRunner::WebsocketHandler] message read error", "error", err)
			return
		}

		if msg_type == websocket.CloseMessage {
			logger.Logger.Debug("[ProfileRunner::WebsocketHandler] received close message from client")
			break
		}

		if msg_type == websocket.TextMessage {
			socket_message := TSWConnector_Message_FromString(string(msg))
			logger.Logger.Debug("[ProfileRunner::WebsocketHandler] received message from client", "message", socket_message)
			c.Subscribers.EmitTimeout(time.Second, socket_message)
			go c.Forward(conn_id, socket_message)
		} else {
			logger.Logger.Debug("[ProfileRunner::WebsocketHandler] received unsupported message %d", "message_type", msg_type)
		}
	}

	cancel_sender()
}

func (c *SocketConnection) Subscribe() (chan TSWConnector_Message, func()) {
	return c.Subscribers.Subscribe()
}

func (c *SocketConnection) IsActive() bool {
	c.OutgoingChannels.Mutex.RLock()
	defer c.OutgoingChannels.Mutex.RUnlock()
	return len(c.OutgoingChannels.Map) > 0
}

func (c *SocketConnection) Stop() error {
	return c.Server.Close()
}

func (c *SocketConnection) Start() error {
	/* try to bind each port in the range; if successfull return nil */
	for port := SOCKET_CONNECTION_PORT_RANGE_START; port <= SOCKET_CONNECTION_PORT_RANGE_END; port++ {
		c.Server.Addr = fmt.Sprintf("0.0.0.0:%d", port)
		logger.Logger.Debug("[SocketConnection::start] Starting direct control server", "addr", c.Server.Addr)
		err := c.Server.ListenAndServe()
		if err == http.ErrServerClosed {
			/*
				server closed is the only acceptable error because this is a graceful shutdown;
				any other error should continue trying the next port until exhausted
			*/
			return nil
		}
		logger.Logger.Error("[SocketConnection::start] could not start direct control server", "addr", c.Server.Addr)
	}
	return fmt.Errorf("exhausted all port options")
}

func (c *SocketConnection) Send(m TSWConnector_Message) error {
	c.OutgoingChannels.ForEach(func(channel chan TSWConnector_Message, key uuid.UUID) bool {
		chan_utils.SendTimeout(channel, time.Second, m)
		return true
	})
	return nil
}

func (c *SocketConnection) Forward(from uuid.UUID, m TSWConnector_Message) error {
	c.OutgoingChannels.ForEach(func(channel chan TSWConnector_Message, key uuid.UUID) bool {
		if key != from {
			chan_utils.SendTimeout(channel, time.Second, m)
		}
		return true
	})
	return nil
}

func NewSocketConnection(ctx context.Context) *SocketConnection {
	mux := http.NewServeMux()
	server := &http.Server{
		BaseContext: func(l net.Listener) context.Context {
			return ctx
		},
		Addr:    fmt.Sprintf("0.0.0.0:%d", SOCKET_CONNECTION_PORT_RANGE_START),
		Handler: mux,
	}
	conn := SocketConnection{
		WsUpgrader: &websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		Server:           server,
		OutgoingChannels: map_utils.NewLockMap[uuid.UUID, chan TSWConnector_Message](),
		Subscribers:      pubsub_utils.NewPubSubSlice[TSWConnector_Message](),
	}
	mux.HandleFunc("/", conn.WebsocketHandler)
	return &conn
}
