package server

import (
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

	proc := NewMessageProcessor(ProcessorConfig{
		Hostname:         s.cfg.Hostname,
		Version:          s.cfg.Version,
		Paster:           s.cfg.Paster,
		PasteMu:          &s.pasteMu,
		MaxPendingPastes: s.maxPendingPastes(),
		OnTextReceived:   s.cfg.OnTextReceived,
		WriteOutbound:    writeOutbound,
	})
	defer proc.Close()

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

		if err := proc.HandleRaw(message); err != nil {
			return
		}
	}
}
