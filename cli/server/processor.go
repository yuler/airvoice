package server

import (
	"errors"
	"sync"
	"time"

	"github.com/yuler/airvoice/cli/paste"
	"github.com/yuler/airvoice/cli/protocol"
)

// ProcessorConfig configures a transport-agnostic inbound message processor
// shared between the LAN WebSocket server and the outbound Relay host client.
type ProcessorConfig struct {
	Hostname         string
	Version          string
	Paster           paste.Paster
	PasteMu          *sync.Mutex
	MaxPendingPastes int
	OnHello          func(device, app string)
	OnTextReceived   func(content, device string)
	WriteOutbound    func(protocol.Outbound) error
	Now              func() time.Time
}

// MessageProcessor handles inbound hello/text/ping protocol messages with a
// bounded paste queue and serialized paster access.
type MessageProcessor struct {
	cfg        ProcessorConfig
	pasteMu    *sync.Mutex
	now        func() time.Time
	pasteSem   chan struct{}
	pasteQueue chan protocol.Inbound
	done       chan struct{}

	mu     sync.Mutex
	closed bool
}

// NewMessageProcessor creates and starts a MessageProcessor worker.
// Call Close when the underlying peer session ends.
func NewMessageProcessor(cfg ProcessorConfig) *MessageProcessor {
	maxPending := cfg.MaxPendingPastes
	if maxPending <= 0 {
		maxPending = defaultMaxPendingPastes
	}
	pasteMu := cfg.PasteMu
	if pasteMu == nil {
		pasteMu = &sync.Mutex{}
	}
	nowFn := cfg.Now
	if nowFn == nil {
		nowFn = time.Now
	}

	p := &MessageProcessor{
		cfg:        cfg,
		pasteMu:    pasteMu,
		now:        nowFn,
		pasteSem:   make(chan struct{}, maxPending),
		pasteQueue: make(chan protocol.Inbound, maxPending),
		done:       make(chan struct{}),
	}

	go func() {
		for inbound := range p.pasteQueue {
			p.processText(inbound)
			<-p.pasteSem
		}
	}()

	return p
}

// Close stops accepting new text messages and signals queued items to skip.
func (p *MessageProcessor) Close() {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return
	}
	p.closed = true
	close(p.done)
	close(p.pasteQueue)
	p.mu.Unlock()
}

// HandleRaw parses a JSON application frame and dispatches it.
// Returns a non-nil error only when writing an outbound reply fails.
func (p *MessageProcessor) HandleRaw(message []byte) error {
	inbound, err := protocol.ParseInbound(message)
	if err != nil {
		logStatus("invalid json: %v", err)
		return nil
	}
	return p.HandleInbound(inbound)
}

// HandleInbound dispatches a parsed inbound application message.
// Returns a non-nil error only when writing an outbound reply fails.
func (p *MessageProcessor) HandleInbound(inbound protocol.Inbound) error {
	switch inbound.Type {
	case "hello":
		logStatus("hello from device=%q app=%q", inbound.Device, inbound.App)
		if p.cfg.OnHello != nil {
			p.cfg.OnHello(inbound.Device, inbound.App)
		}
		outbound := protocol.Outbound{
			Type:    "hello",
			Host:    p.cfg.Hostname,
			Version: p.cfg.Version,
			TS:      p.now().Unix(),
		}
		if err := p.writeOutbound(outbound); err != nil {
			logStatus("write error: %v", err)
			return err
		}
		logStatus("hello reply host=%s version=%s", outbound.Host, outbound.Version)
		return nil

	case "text":
		logStatus("text id=%s len=%d preview=%q", inbound.ID, len(inbound.Content), previewText(inbound.Content, 40))
		enqueued, closed := p.enqueueText(inbound)
		if closed {
			return nil
		}
		if !enqueued {
			outbound := protocol.Outbound{
				Type:    "ack",
				ID:      inbound.ID,
				OK:      false,
				Message: "too many pending paste requests",
			}
			if err := p.writeOutbound(outbound); err != nil {
				logStatus("write error: %v", err)
				return err
			}
			logStatus("ack fail id=%s message=%s", outbound.ID, outbound.Message)
		}
		return nil

	case "ping":
		logStatus("ping")
		outbound := protocol.Outbound{Type: "pong"}
		if err := p.writeOutbound(outbound); err != nil {
			logStatus("write error: %v", err)
			return err
		}
		logStatus("pong")
		return nil

	default:
		logStatus("ignored message type=%q", inbound.Type)
		return nil
	}
}

func (p *MessageProcessor) enqueueText(inbound protocol.Inbound) (enqueued bool, closed bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return false, true
	}
	select {
	case p.pasteSem <- struct{}{}:
		p.pasteQueue <- inbound
		return true, false
	default:
		return false, false
	}
}

func (p *MessageProcessor) writeOutbound(outbound protocol.Outbound) error {
	if p.cfg.WriteOutbound == nil {
		return nil
	}
	return p.cfg.WriteOutbound(outbound)
}

func (p *MessageProcessor) processText(inbound protocol.Inbound) {
	select {
	case <-p.done:
		return
	default:
	}

	if p.cfg.OnTextReceived != nil {
		p.cfg.OnTextReceived(inbound.Content, inbound.Device)
	}

	var outbound protocol.Outbound
	var err error
	method := "none"
	if p.cfg.Paster != nil {
		method = p.cfg.Paster.Name()
		p.pasteMu.Lock()
		err = p.cfg.Paster.Paste(inbound.Content)
		p.pasteMu.Unlock()
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

	select {
	case <-p.done:
		return
	default:
	}

	if err := p.writeOutbound(outbound); err != nil {
		logStatus("write error: %v", err)
		return
	}
	if outbound.OK {
		logStatus("ack ok id=%s", outbound.ID)
	} else {
		logStatus("ack fail id=%s message=%s", outbound.ID, outbound.Message)
	}
}
