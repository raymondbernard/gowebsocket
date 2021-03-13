// Implementation inspired by http://gary.beagledreams.com/page/go-websocket-chat.html

package gowebsocket

import (
	"log"
	"net/http"

	"golang.org/x/net/websocket"
)

type hub struct {

	// registered cnnections
	connections map[*connection]bool

	// inbound messages from connections
	broadcast chan string

	// register requests from connections
	register chan *connection

	// unregister requests from connections
	unregister chan *connection
}

var h = hub{
	connections: make(map[*connection]bool),
	broadcast:   make(chan string),
	register:    make(chan *connection),
	unregister:  make(chan *connection),
}

func (h *hub) Run() {
	for {
		select {

		// Register new connection
		case c := <-h.register:
			h.connections[c] = true

		// Unregister connection
		case c := <-h.unregister:
			delete(h.connections, c)
			//close(c.send)

		// Broadcast message to all connections. If send buffer
		// is full unregister and close websocket connection
		case m := <-h.broadcast:
			for c := range h.connections {
				select {
				case c.send <- m:
				default:
					delete(h.connections, c)
					close(c.send)
					go c.conn.Close()
				}
			}

		}
	}
}

type connection struct {
	// connection
	conn *websocket.Conn

	// buffered channel of outbund messages
	send chan string
}

func (c *connection) reader(h *hub) {
	for {
		var message [1000]byte
		n, err := c.conn.Read(message[:])
		if err != nil {
			break
		}
		h.broadcast <- string(message[:n])
	}
	c.conn.Close()
}

func (c *connection) writer(h *hub) {
	for message := range c.send {
		err := websocket.Message.Send(c.conn, message)
		if err != nil {
			break
		}
	}
}

func (s *WSServer) connHandler(conn *websocket.Conn) {
	/* If a handler for new connections is registered, invoke it first
	 * so it can read to or write from the connection before we setup
	 * broadcasting
	 */
	if s.ConnHandler != nil {
		s.ConnHandler.Handler(conn)
	}

	s.Conn = &connection{send: make(chan string, 256), conn: conn}
	s.Hub.register <- s.Conn
	defer func() {
		s.Hub.unregister <- s.Conn
	}()

	go s.Conn.writer(s.Hub)
	s.Conn.reader(s.Hub)
}

type WSConnHandler interface {
	Handler(conn *websocket.Conn)
}

type WSServer struct {
	Hub         *hub
	Server      *http.Server
	Conn        *connection
	ConnHandler WSConnHandler
}

func New(ip, port string) (s *WSServer) {

	s = new(WSServer)

	s.Hub = &hub{
		connections: make(map[*connection]bool),
		broadcast:   make(chan string),
		register:    make(chan *connection),
		unregister:  make(chan *connection),
	}

	s.Server = &http.Server{
		Addr:    ip + port,
		Handler: websocket.Handler(s.connHandler),
	}

	return s
}

/* connHandler is an optional connection handler that may be registered
 * to receive new connections before they are attached to the hub.
 * Can be nil if no connection handler is desired.
 */
func (s *WSServer) SetConnectionHandler(connHandler WSConnHandler) {
	s.ConnHandler = connHandler
}

func (s *WSServer) Start() {
	go s.Hub.Run()

	go func() {

		log.Print("handlr:", s.Server.Handler)
		http.Handle("/"+s.Server.Addr, s.Server.Handler)
		err := s.Server.ListenAndServe()

		if err != nil {
			log.Panic("Websocket server could not start:", err)
		}
	}()
	log.Print("Websocket server started successfully.")
	log.Print("Server info:", s.GetServerInfo())

}

func customHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	})
}

func (s *WSServer) GetServerInfo() string {
	return s.Server.Addr
}
