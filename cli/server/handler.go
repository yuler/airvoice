package server

import (
	"github.com/airvoice/airvoice/cli/protocol"
	"github.com/gorilla/websocket"
)

// handleConnection handles the WebSocket read/write loops.
func (s *Server) handleConnection(conn *websocket.Conn) {
	s.hub.Set(conn)
	defer func() {
		s.hub.Clear(conn)
		logStatus("client disconnected")
	}()

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			logStatus("read error: %v", err)
			break
		}
		if messageType != websocket.TextMessage {
			logStatus("ignored frame type=%d", messageType)
			continue
		}

		inbound, err := protocol.ParseInbound(message)
		if err != nil {
			logStatus("invalid json: %v", err)
			continue
		}

		var outbound protocol.Outbound
		switch inbound.Type {
		case "hello":
			logStatus("hello from device=%q app=%q", inbound.Device, inbound.App)
			outbound = protocol.Outbound{
				Type:    "hello",
				Host:    s.cfg.Hostname,
				Version: s.cfg.Version,
			}
		case "text":
			logStatus("text id=%s len=%d preview=%q", inbound.ID, len(inbound.Content), previewText(inbound.Content, 40))
			err := s.cfg.Paster.Paste(inbound.Content)
			if err != nil {
				logStatus("paste failed id=%s: %v", inbound.ID, err)
				outbound = protocol.Outbound{
					Type:    "ack",
					ID:      inbound.ID,
					OK:      false,
					Message: err.Error(),
				}
			} else {
				logStatus("paste ok id=%s backend=%s", inbound.ID, s.cfg.Paster.Name())
				outbound = protocol.Outbound{
					Type: "ack",
					ID:   inbound.ID,
					OK:   true,
				}
			}
		case "ping":
			logStatus("ping")
			outbound = protocol.Outbound{
				Type: "pong",
			}
		default:
			logStatus("ignored message type=%q", inbound.Type)
			continue
		}

		data, err := outbound.Bytes()
		if err != nil {
			logStatus("encode response failed: %v", err)
			continue
		}

		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			logStatus("write error: %v", err)
			break
		}

		if outbound.Type == "ack" {
			if outbound.OK {
				logStatus("ack ok id=%s", outbound.ID)
			} else {
				logStatus("ack fail id=%s message=%s", outbound.ID, outbound.Message)
			}
		} else if outbound.Type == "pong" {
			logStatus("pong")
		} else if outbound.Type == "hello" {
			logStatus("hello reply host=%s version=%s", outbound.Host, outbound.Version)
		}
	}
}
