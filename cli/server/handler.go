package server

import (
	"errors"
	"sync"

	"github.com/airvoice/airvoice/cli/protocol"
	"github.com/gorilla/websocket"
)

// handleConnection handles the WebSocket read/write loops.
func (s *Server) handleConnection(conn *websocket.Conn) {
	s.hub.Add(conn)
	defer func() {
		if s.hub.Remove(conn) {
			logStatus("client disconnected (active: %d)", s.hub.Count())
			if s.hub.Count() == 0 {
				logStatus("all clients disconnected — waiting for reconnection...")
			}
		}
	}()

	var writeMu sync.Mutex
	writeOutbound := func(outbound protocol.Outbound) error {
		data, err := outbound.Bytes()
		if err != nil {
			return err
		}
		writeMu.Lock()
		defer writeMu.Unlock()
		return conn.WriteMessage(websocket.TextMessage, data)
	}

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

		switch inbound.Type {
		case "hello":
			logStatus("hello from device=%q app=%q", inbound.Device, inbound.App)
			outbound := protocol.Outbound{
				Type:    "hello",
				Host:    s.cfg.Hostname,
				Version: s.cfg.Version,
			}
			if err := writeOutbound(outbound); err != nil {
				logStatus("write error: %v", err)
				return
			}
			logStatus("hello reply host=%s version=%s", outbound.Host, outbound.Version)

		case "text":
			logStatus("text id=%s len=%d preview=%q", inbound.ID, len(inbound.Content), previewText(inbound.Content, 40))
		go func(inbound protocol.Inbound) {
			var outbound protocol.Outbound
			var err error
			if s.cfg.Paster != nil {
				err = s.cfg.Paster.Paste(inbound.Content)
			} else {
				err = errors.New("paster not initialized")
			}
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
				if err := writeOutbound(outbound); err != nil {
					logStatus("write error: %v", err)
				} else {
					if outbound.OK {
						logStatus("ack ok id=%s", outbound.ID)
					} else {
						logStatus("ack fail id=%s message=%s", outbound.ID, outbound.Message)
					}
				}
			}(inbound)

		case "ping":
			logStatus("ping")
			outbound := protocol.Outbound{
				Type: "pong",
			}
			if err := writeOutbound(outbound); err != nil {
				logStatus("write error: %v", err)
				return
			}
			logStatus("pong")

		default:
			logStatus("ignored message type=%q", inbound.Type)
		}
	}
}
