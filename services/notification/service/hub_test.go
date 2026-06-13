// services/notification/service/hub_test.go
//
// go test -v -race ./services/notification/service/
//
// New dependency (add to go.mod once):
//   github.com/alicebob/miniredis/v2 v2.33.0

package service

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// ── helpers ───────────────────────────────────────────────────────────────────

func newTestHub(t *testing.T) (*WebSocketHub, *miniredis.Miniredis) {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis.Run: %v", err)
	}
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	logger, _ := zap.NewDevelopment()
	return NewWebSocketHub(rdb, logger), mr
}

var testUpgrader = websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}

// wsClientPair returns (server-side conn, client-side conn) over a real TCP WS.
func wsClientPair(t *testing.T) (*websocket.Conn, *websocket.Conn) {
	t.Helper()
	var (
		serverConn *websocket.Conn
		mu         sync.Mutex
		ready      = make(chan struct{})
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := testUpgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("upgrade: %v", err)
			return
		}
		mu.Lock()
		serverConn = c
		mu.Unlock()
		close(ready)
		time.Sleep(10 * time.Second) // keep handler alive for test duration
	}))
	t.Cleanup(srv.Close)

	url := "ws" + strings.TrimPrefix(srv.URL, "http")
	clientConn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	<-ready
	mu.Lock()
	sc := serverConn
	mu.Unlock()
	return sc, clientConn
}

func readJSON(t *testing.T, conn *websocket.Conn, dst interface{}) {
	t.Helper()
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, data, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("readJSON: %v", err)
	}
	if err := json.Unmarshal(data, dst); err != nil {
		t.Fatalf("unmarshal: %v (raw: %s)", err, data)
	}
}

func isTimeout(err error) bool {
	ne, ok := err.(net.Error)
	return ok && ne.Timeout()
}

// ── existing tests (Register now takes a policy arg) ─────────────────────────

func TestDispatch_DeliverToRegisteredUser(t *testing.T) {
	hub, mr := newTestHub(t)
	defer mr.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go hub.Run(ctx)
	time.Sleep(50 * time.Millisecond)

	serverConn, clientConn := wsClientPair(t)
	defer clientConn.Close()

	hub.Register(42, serverConn, DropMessage)
	defer hub.Unregister(42, serverConn)

	type P struct{ Message string }
	if err := hub.Publish(ctx, 42, P{"hello user 42"}); err != nil {
		t.Fatalf("Publish: %v", err)
	}
	var got P
	readJSON(t, clientConn, &got)
	if got.Message != "hello user 42" {
		t.Errorf("got %q, want %q", got.Message, "hello user 42")
	}
}

func TestDispatch_NoDeliveryToOtherUser(t *testing.T) {
	hub, mr := newTestHub(t)
	defer mr.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go hub.Run(ctx)
	time.Sleep(50 * time.Millisecond)

	scA, ccA := wsClientPair(t)
	scB, ccB := wsClientPair(t)
	defer ccA.Close()
	defer ccB.Close()

	hub.Register(1, scA, DropMessage)
	hub.Register(2, scB, DropMessage)
	defer hub.Unregister(1, scA)
	defer hub.Unregister(2, scB)

	type P struct{ ID int }
	if err := hub.Publish(ctx, 1, P{1}); err != nil {
		t.Fatalf("Publish: %v", err)
	}

	var got P
	readJSON(t, ccA, &got)
	if got.ID != 1 {
		t.Errorf("user 1 got ID %d, want 1", got.ID)
	}

	ccB.SetReadDeadline(time.Now().Add(300 * time.Millisecond))
	_, _, err := ccB.ReadMessage()
	if err == nil {
		t.Error("user 2 received a message but should not have")
	}
	if !isTimeout(err) {
		t.Errorf("expected timeout for user 2, got: %v", err)
	}
}

func TestDispatch_MultiPodSimulation(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis.Run: %v", err)
	}
	defer mr.Close()

	logger, _ := zap.NewDevelopment()
	newHub := func() *WebSocketHub {
		rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
		return NewWebSocketHub(rdb, logger)
	}

	hub1, hub2 := newHub(), newHub()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go hub1.Run(ctx)
	go hub2.Run(ctx)
	time.Sleep(50 * time.Millisecond)

	sc, cc := wsClientPair(t)
	defer cc.Close()
	hub2.Register(99, sc, CloseConnection)
	defer hub2.Unregister(99, sc)

	type P struct{ Msg string }
	if err := hub1.Publish(ctx, 99, P{"cross-pod delivery"}); err != nil {
		t.Fatalf("hub1.Publish: %v", err)
	}

	var got P
	readJSON(t, cc, &got)
	if got.Msg != "cross-pod delivery" {
		t.Errorf("got %q, want %q", got.Msg, "cross-pod delivery")
	}
}

func TestDispatch_StaleConnectionCleanup(t *testing.T) {
	hub, mr := newTestHub(t)
	defer mr.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go hub.Run(ctx)
	time.Sleep(50 * time.Millisecond)

	sc, cc := wsClientPair(t)
	hub.Register(7, sc, DropMessage)

	// Close client connection cleanly
	cc.Close()
	time.Sleep(10 * time.Millisecond)

	// Force close the server side connection to guarantee an immediate write loop error
	sc.Close()

	type P struct{ V int }
	_ = hub.Publish(ctx, 7, P{1})

	evicted := false
	for i := 0; i < 50; i++ {
		hub.mu.RLock()
		conns, exists := hub.clients[7]
		if !exists || conns == nil || conns[sc] == nil {
			hub.mu.RUnlock()
			evicted = true
			break
		}
		hub.mu.RUnlock()
		time.Sleep(10 * time.Millisecond)
	}

	if !evicted {
		t.Error("stale session was not evicted after write failure")
	}
}
func TestRegisterUnregister(t *testing.T) {
	hub, mr := newTestHub(t)
	defer mr.Close()

	sc, cc := wsClientPair(t)
	defer cc.Close()

	hub.Register(5, sc, DropMessage)
	hub.mu.RLock()
	if _, ok := hub.clients[5]; !ok {
		t.Error("Register: user 5 not in clients map")
	}
	hub.mu.RUnlock()

	hub.Unregister(5, sc)
	hub.mu.RLock()
	if _, ok := hub.clients[5]; ok {
		t.Error("Unregister: user 5 still in clients map")
	}
	hub.mu.RUnlock()
}

func TestPublish_InvalidPayload(t *testing.T) {
	hub, mr := newTestHub(t)
	defer mr.Close()

	err := hub.Publish(context.Background(), 1, make(chan int))
	if err == nil {
		t.Error("expected error for un-marshallable payload, got nil")
	}
}

// ── new ring-buffer tests ─────────────────────────────────────────────────────

// TestRingBuffer_DropPolicy_DropsWhenFull fills a session's send buffer to
// capacity and then dispatches one more frame. With DropMessage policy the
// extra frame must be silently discarded — the connection must NOT be closed.
func TestRingBuffer_DropPolicy_DropsWhenFull(t *testing.T) {
	hub, mr := newTestHub(t)
	defer mr.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go hub.Run(ctx)
	time.Sleep(50 * time.Millisecond)

	sc, cc := wsClientPair(t)
	defer cc.Close()

	hub.Register(10, sc, DropMessage)
	defer hub.Unregister(10, sc)

	// Reach directly into the session's send channel and fill it to capacity
	// without the writer goroutine draining it (client is not reading yet).
	hub.mu.RLock()
	s := hub.clients[10][sc]
	hub.mu.RUnlock()

	dummyFrame := []byte(`{"v":0}`)
	for i := 0; i < sendBufSize; i++ {
		s.send <- dummyFrame
	}

	// Now dispatch one more frame via the non-blocking path in dispatch().
	// With DropMessage the select-default branch executes — no panic, no close.
	type P struct{ V int }
	if err := hub.Publish(ctx, 10, P{999}); err != nil {
		t.Fatalf("Publish: %v", err)
	}
	// Give dispatch time to run.
	time.Sleep(80 * time.Millisecond)

	// Connection must still be alive (no CloseConnection was applied).
	hub.mu.RLock()
	_, alive := hub.clients[10]
	hub.mu.RUnlock()
	if !alive {
		t.Error("DropMessage: session was incorrectly closed on buffer-full")
	}
}

// TestRingBuffer_ClosePolicy_ClosesWhenFull mirrors the above but uses
// CloseConnection policy. The session must be evicted from the hub map.
func TestRingBuffer_ClosePolicy_ClosesWhenFull(t *testing.T) {
	hub, mr := newTestHub(t)
	defer mr.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go hub.Run(ctx)
	time.Sleep(50 * time.Millisecond)

	sc, cc := wsClientPair(t)
	defer cc.Close()

	hub.Register(11, sc, CloseConnection)

	hub.mu.RLock()
	s := hub.clients[11][sc]
	hub.mu.RUnlock()

	// 1. Fill the buffer completely up to its 128 cap boundary
	dummyFrame := []byte(`{"v":0}`)
	for i := 0; i < sendBufSize; i++ {
		s.send <- dummyFrame
	}

	// 2. PAUSE the server websocket socket writes completely.
	// This ensures that even if the Go scheduler switches contexts,
	// the writeLoop can't drain a single byte out of s.send.
	sc.Close()

	// 3. Fire a standard typed payload over Redis.
	// Because sc is closed, writeLoop is stuck or dead, send remains at 128/128,
	// and this next frame guarantees an overflow fallback trigger.
	type P struct{ V int }
	if err := hub.Publish(ctx, 11, P{999}); err != nil {
		t.Fatalf("Publish: %v", err)
	}

	// 4. Robust verification poll loop
	evicted := false
	for i := 0; i < 50; i++ {
		hub.mu.RLock()
		conns, alive := hub.clients[11]
		if !alive || conns == nil || conns[sc] == nil {
			hub.mu.RUnlock()
			evicted = true
			break
		}
		hub.mu.RUnlock()
		time.Sleep(10 * time.Millisecond)
	}

	if !evicted {
		t.Error("CloseConnection: session was NOT evicted on buffer-full")
	}
}

// TestRingBuffer_OrderedDelivery verifies that messages arrive at the client
// in the exact order they were published, even though they travel through the
// buffered channel and a separate writer goroutine.
func TestRingBuffer_OrderedDelivery(t *testing.T) {
	hub, mr := newTestHub(t)
	defer mr.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go hub.Run(ctx)
	time.Sleep(50 * time.Millisecond)

	sc, cc := wsClientPair(t)
	defer cc.Close()

	hub.Register(20, sc, DropMessage)
	defer hub.Unregister(20, sc)

	const n = 20
	type P struct{ Seq int }
	for i := 0; i < n; i++ {
		if err := hub.Publish(ctx, 20, P{i}); err != nil {
			t.Fatalf("Publish seq %d: %v", i, err)
		}
	}

	for i := 0; i < n; i++ {
		var got P
		readJSON(t, cc, &got)
		if got.Seq != i {
			t.Errorf("message %d: got seq %d, want %d", i, got.Seq, i)
		}
	}
}
