package ocppcore

import "sync"

type Hub struct {
	mu      sync.RWMutex
	clients map[string]*Client
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[string]*Client),
	}
}

func (h *Hub) Add(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[client.OCPPID] = client
}

func (h *Hub) Remove(ocppID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, ocppID)
}

func (h *Hub) Get(ocppID string) (*Client, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	client, ok := h.clients[ocppID]
	return client, ok
}

func (h *Hub) List() []ChargerConnectionInfo {
	h.mu.RLock()
	defer h.mu.RUnlock()

	result := make([]ChargerConnectionInfo, 0, len(h.clients))
	for _, client := range h.clients {
		result = append(result, ChargerConnectionInfo{
			OCPPID:      client.OCPPID,
			ConnectedAt: client.ConnectedAt,
			RemoteAddr:  client.RemoteAddr,
			IsOnline:    true,
		})
	}
	return result
}