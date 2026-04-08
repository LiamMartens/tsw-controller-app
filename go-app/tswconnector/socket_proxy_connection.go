package tswconnector

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
	"tsw_controller_app/chan_utils"
	"tsw_controller_app/logger"
	"tsw_controller_app/pubsub_utils"

	"github.com/gorilla/websocket"
)

const SOCKET_PROXY_CONNECTION_OUTGOING_QUEUE_BUFFER_SIZE = 32

var ErrCancelled = errors.New("cancelled")
var ErrUnknownMessage = errors.New("received unknown message type")
var ErrCloseMessage = errors.New("received close message")

type SocketProxyConnection_ConnectionResult struct {
	connection *websocket.Conn
	err        error
}

type SocketProxyConnection_MessageResult struct {
	message string
	err     error
}

type SocketProxyConnection struct {
	context         context.Context
	cancel          context.CancelFunc
	connection      *websocket.Conn
	ServerAddr      string
	OutgoingChannel chan TSWConnector_Message
	Subscribers     *pubsub_utils.PubSubSlice[TSWConnector_Message]
}

var _ TSWConnector = (*SocketProxyConnection)(nil)

func (c *SocketProxyConnection) dial() chan SocketProxyConnection_ConnectionResult {
	done := make(chan SocketProxyConnection_ConnectionResult, 1)
	go func() {
		dialer := websocket.Dialer{
			HandshakeTimeout: 5 * time.Second,
		}
		for port := SOCKET_CONNECTION_PORT_RANGE_START; port <= SOCKET_CONNECTION_PORT_RANGE_END; port++ {
			u := url.URL{Scheme: "ws", Host: fmt.Sprintf("%s:%d", c.ServerAddr, port), Path: "/"}
			conn, _, err := dialer.Dial(u.String(), nil)
			if err != nil {
				/*
					if the dial was unsuccessfull we can retry except if this is the last port in the range;
					if the last port has been reached we will also send a message to the channel to propagate the error
				*/
				logger.Logger.Error("[SocketProxyConnection::dial] could not connect to server", "error", err, "port", port)
				if port >= SOCKET_CONNECTION_PORT_RANGE_END {
					done <- SocketProxyConnection_ConnectionResult{connection: nil, err: err}
				}
			} else {
				/* if the dial was successfull then we can return the connection */
				done <- SocketProxyConnection_ConnectionResult{connection: conn, err: nil}
			}
		}
	}()
	return done
}

func (c *SocketProxyConnection) waitForMessage(conn *websocket.Conn) chan SocketProxyConnection_MessageResult {
	received := make(chan SocketProxyConnection_MessageResult, 1)
	go func() {
		msg_type, msg, err := conn.ReadMessage()
		if err != nil {
			logger.Logger.Error("[SocketProxyConnection::waitForMessage] failed to read from connection", "error", err)
			received <- SocketProxyConnection_MessageResult{message: "", err: err}
		} else if msg_type == websocket.TextMessage {
			received <- SocketProxyConnection_MessageResult{message: string(msg), err: nil}
		} else if msg_type == websocket.CloseMessage {
			received <- SocketProxyConnection_MessageResult{message: "", err: ErrCloseMessage}
		} else {
			received <- SocketProxyConnection_MessageResult{message: "", err: ErrUnknownMessage}
		}
	}()
	return received
}

func (c *SocketProxyConnection) Port() int {
	if c.connection == nil {
		return SOCKET_CONNECTION_PORT_RANGE_START
	}
	addr_split := strings.Split(c.connection.LocalAddr().String(), ":")
	port, _ := strconv.Atoi(addr_split[len(addr_split)-1])
	return port
}

func (c *SocketProxyConnection) IsActive() bool {
	return c.connection != nil
}

func (c *SocketProxyConnection) Subscribe() (chan TSWConnector_Message, func()) {
	return c.Subscribers.Subscribe()
}

func (c *SocketProxyConnection) Start() error {
	var closeexisting func() error
	closeexisting = func() error {
		if c.connection != nil {
			connection := c.connection
			c.connection = nil
			return connection.Close()
		}
		return nil
	}
	/* make sure the connection will always be closed at the end */
	defer closeexisting()

	for {
		closeexisting()
		select {
		case conn := <-c.dial():
			if conn.err != nil {
				<-time.After(5 * time.Second)
				continue
			}
			c.connection = conn.connection
		case <-c.context.Done():
			return nil
		}

		connection := c.connection
		sender_ctx, cancel_sender := context.WithCancel(c.context)
		go func() {
			for {
				select {
				case <-sender_ctx.Done():
					return
				case message := <-c.OutgoingChannel:
					err := connection.WriteMessage(websocket.TextMessage, []byte(message.ToString()))
					if err != nil {
						cancel_sender()
						return
					}
				}
			}
		}()

	read_loop:
		for {
			select {
			case msg := <-c.waitForMessage(connection):
				if msg.err != nil {
					break read_loop
				}
				socket_message := TSWConnector_Message_FromString(msg.message)
				c.Subscribers.EmitTimeout(time.Second, socket_message)
			case <-c.context.Done():
				return nil
			}
		}
	}
}

func (c *SocketProxyConnection) Stop() error {
	c.cancel()
	return nil
}

func (c *SocketProxyConnection) Send(m TSWConnector_Message) error {
	return chan_utils.SendTimeout(c.OutgoingChannel, time.Second, m)
}

func NewSocketProxyConnection(ctx context.Context, addr string) *SocketProxyConnection {
	child_ctx, child_cancel := context.WithCancel(ctx)
	return &SocketProxyConnection{
		context:         child_ctx,
		cancel:          child_cancel,
		connection:      nil,
		ServerAddr:      addr,
		OutgoingChannel: make(chan TSWConnector_Message, SOCKET_PROXY_CONNECTION_OUTGOING_QUEUE_BUFFER_SIZE),
		Subscribers:     pubsub_utils.NewPubSubSlice[TSWConnector_Message](),
	}
}
