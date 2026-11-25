package server

import (
	"errors"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/yuler/airvoice/cli/protocol"
)

// handleConnection handles the WebSocket read/write loops.
func (s *Server) handleConnection(conn *websocket.Conn) {
	s.hub.Add(conn)
	done := make(chan struct{})
	defer func() {
		close(done)
		if s.hub.Remove(conn) {
			logStatus("client disconnected (active: %d)", s.hub.Count())
			if s.hub.Count() == 0 {
				logStatus("all clients disconnected — waiting for reconnection...")
				if s.cfg.OnDisconnect != nil {
					s.cfg.OnDisconnect()
				}
			}
		}
	}()

	pongWait := s.pongWait()
	pingInterval := s.pingInterval()
	writeTimeout := s.writeTimeout()

	conn.SetReadLimit(s.maxMessageBytes())
	_ = conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	var writeMu sync.Mutex
	writeOutbound := func(outbound protocol.Outbound) error {
		data, err := outbound.Bytes()
		if err != nil {
			return err
		}
		writeMu.Lock()
		defer writeMu.Unlock()
		_ = conn.SetWriteDeadline(time.Now().Add(writeTimeout))
		return conn.WriteMessage(websocket.TextMessage, data)
	}

	if pingInterval > 0 {
		go func() {
			ticker := time.NewTicker(pingInterval)
			defer ticker.Stop()
			for {
				select {
				case <-done:
					return
				case <-ticker.C:
					writeMu.Lock()
					_ = conn.SetWriteDeadline(time.Now().Add(writeTimeout))
					err := conn.WriteMessage(websocket.PingMessage, nil)
					writeMu.Unlock()
					if err != nil {
						return
					}
				}
			}
		}()
	}

	maxPending := s.maxPendingPastes()
	pasteSem := make(chan struct{}, maxPending)
	pasteQueue := make(chan protocol.Inbound, maxPending)
	defer close(pasteQueue)

	go func() {
		for inbound := range pasteQueue {
			s.processText(inbound, done, writeOutbound)
			<-pasteSem
		}
	}()

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			logStatus("read error: %v", err)
			break
		}
		_ = conn.SetReadDeadline(time.Now().Add(pongWait))

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
			select {
			case pasteSem <- struct{}{}:
				pasteQueue <- inbound
			default:
				outbound := protocol.Outbound{
					Type:    "ack",
					ID:      inbound.ID,
					OK:      false,
					Message: "too many pending paste requests",
				}
				if err := writeOutbound(outbound); err != nil {
					logStatus("write error: %v", err)
					return
				}
				logStatus("ack fail id=%s message=%s", outbound.ID, outbound.Message)
			}

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

func (s *Server) processText(inbound protocol.Inbound, done <-chan struct{}, writeOutbound func(protocol.Outbound) error) {
	select {
	case <-done:
		return
	default:
	}

	if s.cfg.OnTextReceived != nil {
		s.cfg.OnTextReceived(inbound.Content, inbound.Device)
	}

	var outbound protocol.Outbound
	var err error
	method := "none"
	if s.cfg.Paster != nil {
		method = s.cfg.Paster.Name()
		s.pasteMu.Lock()
		err = s.cfg.Paster.Paste(inbound.Content)
		s.pasteMu.Unlock()
	} else {
		err = errors.New("paster not initialized")
	}

	if err != nil {
		logStatus("paste failed id=%s method=%s: %v", inbound.ID, method, err)
		outbound = protocol.Outbound{
			Type:    "ack",
			ID:      inbound.ID,
			OK:      false,
			Message: err.Error(),
		}
	} else {
		logStatus("paste ok id=%s method=%s", inbound.ID, method)
		outbound = protocol.Outbound{
			Type: "ack",
			ID:   inbound.ID,
			OK:   true,
		}
	}

	if err := writeOutbound(outbound); err != nil {
		logStatus("write error: %v", err)
		return
	}
	if outbound.OK {
		logStatus("ack ok id=%s", outbound.ID)
	} else {
		logStatus("ack fail id=%s message=%s", outbound.ID, outbound.Message)
	}
}
