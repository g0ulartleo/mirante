package api

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"github.com/g0ulartleo/mirante/internal/alarm"
	"github.com/g0ulartleo/mirante/internal/signal"
	"github.com/g0ulartleo/mirante/internal/web/dashboard"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"golang.org/x/net/websocket"
)

// jsonWSClient is a single connected websocket consumer of JSON alarm updates.
type jsonWSClient struct {
	broker *JSONWebSocketBroker
	conn   *websocket.Conn
	send   chan []byte
}

// JSONWebSocketBroker fans out raw JSON []alarm.AlarmSignals payloads (published
// on the Redis "dashboard:updates" channel) to every connected client. Unlike
// the dashboard broker it forwards the raw JSON untouched so non-browser clients
// (the TUI / SSH app) can decode it directly.
type JSONWebSocketBroker struct {
	clients    map[*jsonWSClient]bool
	broadcast  chan []byte
	register   chan *jsonWSClient
	unregister chan *jsonWSClient
	mu         sync.RWMutex
	redis      *redis.Client
	pubsub     *redis.PubSub
}

func NewJSONWebSocketBroker(redisAddr string) *JSONWebSocketBroker {
	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
	pubsub := rdb.Subscribe(context.Background(), "dashboard:updates")

	broker := &JSONWebSocketBroker{
		clients:    make(map[*jsonWSClient]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *jsonWSClient),
		unregister: make(chan *jsonWSClient),
		redis:      rdb,
		pubsub:     pubsub,
	}

	go func() {
		for msg := range pubsub.Channel() {
			broker.broadcast <- []byte(msg.Payload)
		}
	}()
	go broker.run()

	return broker
}

func (b *JSONWebSocketBroker) run() {
	for {
		select {
		case client := <-b.register:
			b.mu.Lock()
			b.clients[client] = true
			b.mu.Unlock()

		case client := <-b.unregister:
			b.mu.Lock()
			if _, ok := b.clients[client]; ok {
				delete(b.clients, client)
				close(client.send)
			}
			b.mu.Unlock()

		case message := <-b.broadcast:
			b.mu.RLock()
			for client := range b.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(b.clients, client)
				}
			}
			b.mu.RUnlock()
		}
	}
}

func (b *JSONWebSocketBroker) Close() error {
	if b.pubsub != nil {
		return b.pubsub.Close()
	}
	return nil
}

func (c *jsonWSClient) readPump() {
	defer func() {
		c.broker.unregister <- c
		c.conn.Close()
	}()

	var message []byte
	for {
		if err := websocket.Message.Receive(c.conn, &message); err != nil {
			return
		}
	}
}

func (c *jsonWSClient) writePump() {
	defer c.conn.Close()
	for message := range c.send {
		if err := websocket.Message.Send(c.conn, message); err != nil {
			return
		}
	}
}

// HandleJSONWebSocket upgrades the connection, sends an initial snapshot of the
// current alarm signals, then streams live updates as raw JSON.
func HandleJSONWebSocket(broker *JSONWebSocketBroker, signalService *signal.Service, alarmService *alarm.AlarmService) echo.HandlerFunc {
	return func(c echo.Context) error {
		websocket.Handler(func(ws *websocket.Conn) {
			client := &jsonWSClient{
				broker: broker,
				conn:   ws,
				send:   make(chan []byte, 256),
			}
			client.broker.register <- client

			go client.writePump()

			if snapshot, err := snapshotJSON(signalService, alarmService); err != nil {
				log.Printf("Error building alarm signals snapshot: %v", err)
			} else {
				client.send <- snapshot
			}

			client.readPump()
		}).ServeHTTP(c.Response(), c.Request())
		return nil
	}
}

func snapshotJSON(signalService *signal.Service, alarmService *alarm.AlarmService) ([]byte, error) {
	alarmSignals, err := dashboard.GetAlarmSignals(signalService, alarmService)
	if err != nil {
		return nil, err
	}
	return json.Marshal(alarmSignals)
}
