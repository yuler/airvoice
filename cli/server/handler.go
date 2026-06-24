package server

import (
	"github.com/airvoice/airvoice/cli/protocol"
	"github.com/gorilla/websocket"
)

// handleConnection handles the WebSocket read/write loops.
func (s *Server) handleConnection(conn *websocket.Conn) {
	s.hub.Set(conn)
	defer s.hub.Clear(conn)

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			break
		}
		if messageType != websocket.TextMessage {
			continue
		}

		inbound, err := protocol.ParseInbound(message)
		if err != nil {
			continue
		}

		var outbound protocol.Outbound
		switch inbound.Type {
		case "hello":
			outbound = protocol.Outbound{
				Type:    "hello",
				Host:    s.cfg.Hostname,
				Version: s.cfg.Version,
			}
		case "text":
			err := s.cfg.Paster.Paste(inbound.Content)
			if err != nil {
				outbound = protocol.Outbound{
					Type:    "ack",
					ID:      inbound.ID,
					OK:      false,
					Message: err.Error(),
				}
			} else {
				outbound = protocol.Outbound{
					Type: "ack",
					ID:   inbound.ID,
					OK:   true,
				}
			}
		case "ping":
			outbound = protocol.Outbound{
				Type: "pong",
			}
		default:
			continue
		}

		data, err := outbound.Bytes()
		if err != nil {
			continue
		}

		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			break
		}
	}
}
