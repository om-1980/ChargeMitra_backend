package ocppcore

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	OCPPID      string
	Conn        *websocket.Conn
	ConnectedAt time.Time
	RemoteAddr  string
	mu          sync.Mutex
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
	return c.Conn.WriteMessage(messageType, data)
}

func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.Conn.Close()
}