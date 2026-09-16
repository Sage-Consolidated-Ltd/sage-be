package http

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gofiber/websocket/v2"
)

// DashboardHub manages active WebSocket connections subscribed to real-time organization dashboard updates.
type DashboardHub struct {
	mu      sync.RWMutex
	clients map[string]map[*websocket.Conn]bool
}

func NewDashboardHub() *DashboardHub {
	return &DashboardHub{
		clients: make(map[string]map[*websocket.Conn]bool),
	}
}

// Register adds a new client connection for a specific organization.
func (h *DashboardHub) Register(orgID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[orgID]; !ok {
		h.clients[orgID] = make(map[*websocket.Conn]bool)
	}
	h.clients[orgID][conn] = true
	log.Printf("[WebSocket Hub] Client registered for org %s. Total connections: %d", orgID, len(h.clients[orgID]))
}

// Unregister removes a client connection.
func (h *DashboardHub) Unregister(orgID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if connections, ok := h.clients[orgID]; ok {
		delete(connections, conn)
		if len(connections) == 0 {
			delete(h.clients, orgID)
		}
	}
	_ = conn.Close()
	log.Printf("[WebSocket Hub] Client unregistered for org %s", orgID)
}

// Broadcast serializes the payload and pushes it to all active connections for the given organization.
func (h *DashboardHub) Broadcast(orgID string, payload interface{}) {
	h.mu.RLock()
	connections, ok := h.clients[orgID]
	if !ok || len(connections) == 0 {
		h.mu.RUnlock()
		return
	}

	data, err := json.Marshal(payload)
	if err != nil {
		h.mu.RUnlock()
		log.Printf("[WebSocket Hub] Broadcast marshal error: %v", err)
		return
	}

	// Copy active connections to minimize lock contention during network writes
	activeConns := make([]*websocket.Conn, 0, len(connections))
	for conn := range connections {
		activeConns = append(activeConns, conn)
	}
	h.mu.RUnlock()

	for _, conn := range activeConns {
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Printf("[WebSocket Hub] Write error for org %s, unregistering conn: %v", orgID, err)
			h.Unregister(orgID, conn)
		}
	}
}
