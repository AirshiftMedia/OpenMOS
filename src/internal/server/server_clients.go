package server

import "airshift/openmos/pkg/logger"

// registerClient registers a client connection
func (s *TCPServer) registerClient(client *ClientConnection) {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()

	s.clients[client.ID()] = client
	logger.Infof("Client registered: %s", client.ID())
}

// unregisterClient removes a client connection
func (s *TCPServer) unregisterClient(clientID string) {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()

	delete(s.clients, clientID)
	logger.Infof("Client unregistered: %s", clientID)
}

// GetClient returns a client connection by ID
func (s *TCPServer) GetClient(clientID string) (*ClientConnection, bool) {
	s.clientsMu.RLock()
	defer s.clientsMu.RUnlock()

	client, ok := s.clients[clientID]
	return client, ok
}

// BroadcastMessage sends a message to all connected clients
func (s *TCPServer) BroadcastMessage(data []byte) {
	s.clientsMu.RLock()
	defer s.clientsMu.RUnlock()

	for _, client := range s.clients {
		// Send in a non-blocking way to avoid one slow client affecting others
		go func(c *ClientConnection) {
			if err := c.Write(data); err != nil {
				logger.Errorf("Error broadcasting to client %s: %v", c.ID(), err)
			}
		}(client)
	}
}
