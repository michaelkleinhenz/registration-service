package websockets

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/codeready-toolchain/registration-service/pkg/middleware"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

const (
	// time allowed to write a message to the peer.
	writeWait = 10 * time.Second
	// time allowed to read the next pong message from the peer.
	pongWait = 10 * time.Second
	// send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
	// maximum message size allowed from peer.
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// ChannelMessage represents a message over the channel
type channelMessage struct {
	// subscription of the peer identity
	sub string
	// message data
	body []byte
}

// Hub maintains the set of active clients
type Hub struct {
	// registered clients, mapped to the sub id.
	clients map[*client]string
	// outbound messages to clients.
	outbound chan *channelMessage
	// inbound messages from clients.
	inbound chan *channelMessage
	// register requests from clients.
	register chan *client
	// unregister requests from clients.
	unregister chan *client
	// MessageHandler will be called when a message arrives
	messageHandler func(string, []byte)
}

// NewHub creates a new Hub instance.
func NewHub() *Hub {
	hub := &Hub{
		outbound:       make(chan *channelMessage),
		inbound:        make(chan *channelMessage),
		register:       make(chan *client),
		unregister:     make(chan *client),
		clients:        make(map[*client]string),
		messageHandler: nil,
	}
	go hub.run()
	return hub
}

// SetMessageHandler sets the message handler for this hub.
func (h *Hub) SetMessageHandler(handlerFunc func(string, []byte)) error {
	if handlerFunc == nil {
		return errors.New("error setting websockets handler func, is nil")
	}
	h.messageHandler = handlerFunc
	return nil
}

// SendMessage sends a message to a client identified with the subject.
// Returns false if there is no active connection to a client
// identified with the given subject.
func (h *Hub) SendMessage(subject string, message []byte) bool {
	fmt.Println("SEND MESSAGE")
	if !h.IsSubjectAvailable(subject) {
		return false
	}
	h.outbound <- &channelMessage{
		sub:  subject,
		body: message,
	}
	return true
}

// IsSubjectAvailable checks if there is an active connection to a client
// identified by the given subject. Returns false if not.
func (h *Hub) IsSubjectAvailable(subject string) bool {
	for _, sub := range h.clients {
		if sub == subject {
			return true
		}
	}
	return false
}

// Run runs the hub's main loop.
func (h *Hub) run() {
	for {
		fmt.Print("LOOP ")

		select {
		case client := <-h.register:
			h.clients[client] = client.sub
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				log.Printf("unregistering client for sub %s", client.sub)
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.inbound:
			// message incoming from the pump
			log.Printf("incoming message from client sub %s", message.sub)
			// send the message out with using the handler
			if h.messageHandler != nil {
				h.messageHandler(message.sub, message.body)
			}
			log.Printf("warning: incoming message from client sub %s is ignored, no handler set", message.sub)
		case message := <-h.outbound:
			fmt.Println("SEND MESSAGE DETECTED ON RUN")
			// message appeared on the outbound channel, find the matching client
			for client, sub := range h.clients {
				if sub == message.sub {
					// found the client, send message out
					select {
					case client.send <- message.body:
						log.Printf("sending message to client for sub %s", client.sub)
						return
					default:
						// send message failed, terminate conn with this client
						close(client.send)
						delete(h.clients, client)
						return
					}
				}
			}
			// the client was not found for this sub
			log.Printf("error client not found for sub %s when trying to send outbound message", message.sub)
		}
	}
}

// Client is a middleman between the websocket connection and the hub.
type client struct {
	hub *Hub
	// the websocket connection.
	conn *websocket.Conn
	// the sub of the client identity
	sub string
	// buffered channel of outbound messages.
	send chan []byte
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(s string) error {
		// received pong, reset deadline for this conn
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			log.Printf("error reading from websocket connection with sub %s: %v", c.sub, err)
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error closing websocket connection with sub %s: %v", c.sub, err)
			}
			break
		}
		receivedMessage := &channelMessage{
			sub:  c.sub,
			body: message,
		}
		log.Printf("connection received message from sub %s: %s", receivedMessage.sub, receivedMessage.body)
		// put the message on the hub channel
		c.hub.inbound <- receivedMessage
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			// a message is on the outbound client queue
			fmt.Println("MESSAGE DETECTED ON OUTGOING PUMP")
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)
			// Add queued chat messages to the current websocket message
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(<-c.send)
			}
			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			// the ticker kicked in, refresh the deadline for this connection
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// HTTPHandler handles websocket requests from the peers.
func HTTPHandler(hub *Hub, c *gin.Context) {
	w := c.Writer
	r := c.Request
	// the subject is injected into the context by the
	// auth middleware and is trusted info
	subject, exists := c.Get(middleware.SubKey)
	if !exists {
		log.Println(errors.New("websocket connect without subject claim"))
		return
	}
	subjStr, ok := subject.(string)
	if !ok {
		log.Println(errors.New("websocket connect with non-string subject claim"))
		return
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &client{hub: hub, conn: conn, sub: subjStr, send: make(chan []byte, 256)}
	hub.register <- client
	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}