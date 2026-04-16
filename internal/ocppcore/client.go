package ocppcore

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const clientWriteWait = 10 * time.Second

type Client struct {
	OCPPID      string
	Conn        *websocket.Conn
	ConnectedAt time.Time
	RemoteAddr  string
	mu          sync.Mutex
	closed      bool
}

func NewClient(ocppID string, conn *websocket.Conn, remoteAddr string) *Client {
	return &Client{
		OCPPID:      ocppID,
		Conn:        conn,
		ConnectedAt: time.Now(),
		RemoteAddr:  remoteAddr,
	}
}

func (c *Client) WriteMessage(messageType int, data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return websocket.ErrCloseSent
	}

	_ = c.Conn.SetWriteDeadline(time.Now().Add(clientWriteWait))
	return c.Conn.WriteMessage(messageType, data)
}

func (c *Client) WriteControl(messageType int, data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return websocket.ErrCloseSent
	}

	return c.Conn.WriteControl(messageType, data, time.Now().Add(clientWriteWait))
}

func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	c.closed = true
	_ = c.Conn.SetWriteDeadline(time.Now().Add(clientWriteWait))
	_ = c.Conn.WriteControl(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, "closing connection"),
		time.Now().Add(clientWriteWait),
	)

	return c.Conn.Close()
}

func (c *Client) IsClosed() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.closed
}