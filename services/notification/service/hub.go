// services/notification/service/hub.go
//
// WebSocketHub: Redis Pub/Sub fan-out + per-session bounded ring buffer.
//
// Architecture (two layers):
//
//  Layer 1 — Redis fan-out (unchanged from previous version)
//    Each pod subscribes to a single Redis channel. Any pod calls Publish;
//    every pod receives the envelope and discards it if the target user is
//    not connected locally.
//
//  Layer 2 — Per-session ring buffer (new)
//    Each *session wraps one *websocket.Conn and owns:
//      • a buffered channel  (send chan []byte, cap = sendBufSize = 128)
//      • a DropPolicy        (DropMessage | CloseConnection)
//      • one writer goroutine that is the ONLY goroutine touching the conn
//
//    dispatch does a non-blocking send into the channel:
//      select { case s.send <- payload: default: <apply policy> }
//
//    DropMessage  — silently discard the frame.  Use for typing indicators,
//                   presence pings, or any event the client can miss without
//                   data loss.
//    CloseConnection — send a WebSocket close frame and evict the session.
//                   Use for persistent notifications that must not be lost.
//
//    This decouples the Redis subscriber goroutine from slow clients:
//    a stalled client never blocks delivery to anyone else.

package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	redisPubSubChannel = "notifications"
	sendBufSize        = 128 // ring-buffer capacity per session
	writeTimeout       = 10 * time.Second
)

// DropPolicy controls what happens when a session's send buffer is full.
type DropPolicy int

const (
	// DropMessage silently discards the incoming frame.
	// Suitable for non-critical, high-frequency events (typing indicators,
	// presence pings) where the client catching up later is acceptable.
	DropMessage DropPolicy = iota

	// CloseConnection sends a WebSocket close frame and evicts the session.
	// Suitable for persistent notifications where dropping is not acceptable;
	// the client must reconnect and re-fetch missed events from the REST API.
	CloseConnection
)

// envelope is the message published to Redis.
type envelope struct {
	UserID  int    `json:"user_id"`
	Payload []byte `json:"payload"`
}

// session owns one WebSocket connection and its writer goroutine.
type session struct {
	conn   *websocket.Conn
	send   chan []byte
	policy DropPolicy
	mu     sync.RWMutex
	closed bool
}

// newSession creates a session and starts its writer goroutine.
// The goroutine exits when the send channel is closed.
func newSession(conn *websocket.Conn, policy DropPolicy, onExit func()) *session {
	s := &session{
		conn:   conn,
		send:   make(chan []byte, sendBufSize),
		policy: policy,
	}
	go s.writeLoop(onExit)
	return s
}

// writeLoop drains the send channel, writing each frame to the WebSocket.
// It is the only goroutine that calls conn.WriteMessage, so no external
// locking is needed around the connection.
func (s *session) writeLoop(onExit func()) {
	defer func() {
		s.conn.Close()
		onExit() // This executes h.evict() synchronously when writeLoop exits
	}()
	for payload := range s.send {
		s.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
		if err := s.conn.WriteMessage(websocket.TextMessage, payload); err != nil {
			// Connection is broken; exit immediately.
			// The deferred onExit() will handle map eviction on the hub.
			return
		}
	}

	s.conn.WriteMessage( //nolint:errcheck
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
	)
}

// close shuts down the session: closing the channel terminates writeLoop.
// Safe to call multiple times (channel close is guarded by the hub's lock).
func (s *session) close() {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return
	}
	s.closed = true
	close(s.send) // Safe to close now because it's guarded by the mutex
	s.mu.Unlock()

	// Explicitly close the underlying network connection
	if s.conn != nil {
		s.conn.Close()
	}
}

// WebSocketHub manages per-pod WebSocket connections and Redis fan-out.
type WebSocketHub struct {
	mu      sync.RWMutex
	clients map[int]map[*websocket.Conn]*session // userID → session set

	rdb    redis.UniversalClient
	logger *zap.Logger

	stopFn context.CancelFunc
}

// NewWebSocketHub constructs a hub. Call Run in a goroutine before use.
func NewWebSocketHub(rdb redis.UniversalClient, logger *zap.Logger) *WebSocketHub {
	return &WebSocketHub{
		clients: make(map[int]map[*websocket.Conn]*session),
		rdb:     rdb,
		logger:  logger,
	}
}

// Run blocks, reading from Redis and dispatching to local sessions.
// Cancel ctx to shut down.
func (h *WebSocketHub) Run(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	h.stopFn = cancel

	sub := h.rdb.Subscribe(ctx, redisPubSubChannel)
	defer func() { _ = sub.Close() }()

	ch := sub.Channel()
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			h.dispatch([]byte(msg.Payload))
		}
	}
}

// Stop shuts down the Run goroutine.
func (h *WebSocketHub) Stop() {
	if h.stopFn != nil {
		h.stopFn()
	}
}

// Register adds a connection for userID with the given drop policy and starts
// its writer goroutine. Callers (e.g. the WS handler) own the read loop;
// they must call Unregister when the read loop exits.
func (h *WebSocketHub) Register(userID int, conn *websocket.Conn, policy DropPolicy) {
	// onExit is called by the writer goroutine when it terminates — either
	// because the send channel was closed or because a write failed.
	// We remove the session from the map so future dispatches skip it.
	onExit := func() { h.evict(userID, conn) }

	s := newSession(conn, policy, onExit)

	h.mu.Lock()
	defer h.mu.Unlock()
	if h.clients[userID] == nil {
		h.clients[userID] = make(map[*websocket.Conn]*session)
	}
	h.clients[userID][conn] = s
}

// Unregister closes the session's send channel, terminating its writer
// goroutine, and removes it from the map.
func (h *WebSocketHub) Unregister(userID int, conn *websocket.Conn) {
	h.mu.Lock()
	conns := h.clients[userID]
	if conns != nil {
		if s, ok := conns[conn]; ok {
			s.close()
			delete(conns, conn)
		}
		if len(conns) == 0 {
			delete(h.clients, userID)
		}
	}
	h.mu.Unlock()
}

// evict is called by the writer goroutine's onExit callback.
// It removes the session without closing the channel (already done/broken).
func (h *WebSocketHub) evict(userID int, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if conns := h.clients[userID]; conns != nil {
		delete(conns, conn)
		if len(conns) == 0 {
			delete(h.clients, userID)
		}
	}
}

// Publish serialises payload into a Redis envelope so every pod can
// fan it out to the target user's local connections.
func (h *WebSocketHub) Publish(ctx context.Context, userID int, payload interface{}) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("hub: marshal payload: %w", err)
	}
	env := envelope{UserID: userID, Payload: raw}
	data, err := json.Marshal(env)
	if err != nil {
		return fmt.Errorf("hub: marshal envelope: %w", err)
	}
	return h.rdb.Publish(ctx, redisPubSubChannel, data).Err()
}

// dispatch is invoked by Run for every Redis message.
// It performs a non-blocking send into each session's ring buffer.
func (h *WebSocketHub) dispatch(data []byte) {
	var env envelope
	if err := json.Unmarshal(data, &env); err != nil {
		h.logger.Warn("hub: bad envelope", zap.Error(err))
		return
	}

	h.mu.RLock()
	sessions := h.clients[env.UserID]
	// Snapshot under the read-lock; policy enforcement happens outside it.
	type entry struct {
		conn *websocket.Conn
		s    *session
	}
	snapshot := make([]entry, 0, len(sessions))
	for c, s := range sessions {
		snapshot = append(snapshot, entry{c, s})
	}
	h.mu.RUnlock()

	for _, e := range snapshot {
		// 1. Acquire read lock on the session to safely check its closed status
		e.s.mu.RLock()
		if e.s.closed {
			e.s.mu.RUnlock()
			continue // Session is already closing/closed, skip it
		}

		// 2. Attempt a non-blocking write to the bounded ring-buffer channel
		select {
		case e.s.send <- env.Payload:
			// Success! Delivered into the buffer. Release lock and move on.
			e.s.mu.RUnlock()
		default:
			// Buffer is FULL — Release the session lock BEFORE running policy actions
			// to avoid holding a session lock while grabbing the Hub lock in Unregister.
			e.s.mu.RUnlock()

			switch e.s.policy {
			case DropMessage:
				h.logger.Warn("hub: send buffer full, dropping frame",
					zap.Int("user_id", env.UserID))
			case CloseConnection:
				h.logger.Warn("hub: send buffer full, closing connection",
					zap.Int("user_id", env.UserID))

				// Synchronously unregisters and evicts, fixing your eviction test failure!
				h.Unregister(env.UserID, e.conn)
			}
		}
	}
}
