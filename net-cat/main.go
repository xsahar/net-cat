package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

// Message represents a chat message with a sender and content.
type Message struct {
	Sender  string
	Content string
}

// Client represents a connected client.
type Client struct {
	conn         net.Conn     // Client network connection
	name         string       // Client name
	msgch        chan Message // Channel to send messages to the client
	joinTime     time.Time    // The time the client joined the chat
	leftChatFlag bool         // Flag to track if the client has left the chat
}

// Server represents the chat server.
type Server struct {
	listenAddr  string           // The address to listen for incoming connections
	ln          net.Listener     // The network listener
	quitch      chan struct{}    // A channel to signal the server to quit
	clients     map[*Client]bool // A map to store the connected clients
	messages    []string         // A slice to store the chat messages
	welcomeMsg  string           // The welcome message read from a file
	mutex       sync.Mutex       // A mutex to synchronize access to shared data
	activeConns int              // tracks the number of active connections

}

// NewServer creates a new chat server.
func NewServer(listenAddr string) (*Server, error) {
	file, err := os.Open("h.txt") // Open a file containing the welcome message
	if err != nil {
		return nil, err
	}
	defer file.Close()

	welcomeMsg, err := io.ReadAll(file) // Read the welcome message from the file
	if err != nil {
		return nil, err
	}

	return &Server{
		listenAddr: listenAddr,
		quitch:     make(chan struct{}),
		clients:    make(map[*Client]bool),
		messages:   []string{},
		welcomeMsg: string(welcomeMsg),
	}, nil
}

// start starts the chat server.
func (s *Server) start() error {
	ln, err := net.Listen("tcp", s.listenAddr) // Create a network listener
	if err != nil {
		return err
	}
	defer ln.Close()
	s.ln = ln

	go s.acceptLoop() // Start accepting client connections in a loop

	<-s.quitch // Wait for the quit signal

	s.closeClients() // Close all client connections

	return nil
}

// acceptLoop accepts incoming client connections.
func (s *Server) acceptLoop() {
	for {
		conn, err := s.ln.Accept() // Attempt to accept a new client connection
		if err != nil {
			fmt.Println("Accept error:", err)
			if netErr, ok := err.(net.Error); ok && !netErr.Temporary() {
				fmt.Println("Stopping accept loop due to non-temporary error")
				break
			}
			continue
		}

		// Check connection limit here and handle the new connection accordingly
		s.mutex.Lock()
		if s.activeConns >= 10 {
			conn.Write([]byte("Server at capacity: Maximum connections reached.\n"))
			conn.Close()
			s.mutex.Unlock()
			fmt.Println("Refused a connection: server at capacity.")
			continue
		}
		s.activeConns++
		s.mutex.Unlock()

		fmt.Println("New connection:", conn.RemoteAddr())
		client := s.createNewClient(conn)
		go s.handleClient(client)
	}
}

func (s *Server) createNewClient(conn net.Conn) *Client {
	client := &Client{
		conn:         conn,
		name:         "",
		msgch:        make(chan Message, 10),
		joinTime:     time.Now(),
		leftChatFlag: false,
	}

	s.mutex.Lock()
	s.clients[client] = true
	s.mutex.Unlock()

	fmt.Println("Sending welcome message")
	portMessage := fmt.Sprintf("nc localhost %s\n", s.listenAddr)
	s.sendClientMessage(client, portMessage+s.welcomeMsg) // Append the welcome message with ASCII art here

	return client
}

// handleClient handles communication with a specific client.
func (s *Server) handleClient(client *Client) {
	reader := bufio.NewReader(client.conn)
	defer func() {
		s.removeClient(client)
		client.conn.Close()
	}()

	// Set client name before entering the loop
	clientName, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Failed to read client name:", err)
		return
	}

	client.name = strings.TrimSpace(clientName)
	s.sendClientMessage(client, fmt.Sprintf("[Server] Welcome, %s!", client.name))
	s.broadcastMessage(fmt.Sprintf("[Server] %s joined the chat.", client.name)) // Correctly announce client joining after setting name

	if len(s.messages) > 0 {
		for _, text := range s.messages {
			s.sendClientMessage(client, fmt.Sprintf(text))
		}
	}

	// Continue with reading further messages
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Println("Client", client.name, "left the chat.")
			} else {
				fmt.Println("Error reading from client:", err)
			}
			return
		}

		message = strings.TrimSpace(message)
		if message == "--quit" { // Check if the message is CTRL+C
			fmt.Println("Client", client.name, "sent (--quit) and is leaving the chat.")
			return // Exit the loop, triggering the defer function that closes the connection
		}

		if message != "" {
			// Broadcast the message to other clients
			msg := fmt.Sprintf("[%s][%s]: %s", time.Now().Format("2006-01-02 15:04:05"), client.name, message)
			s.broadcastMessage(msg)
		}
	}
}

// sendClientMessage sends a message to a specific client
func (s *Server) sendClientMessage(client *Client, message string) {
	_, err := client.conn.Write([]byte(message + "\n"))
	if err != nil {
		fmt.Println("Error sending message to client:", err)
	}
}

// broadcastMessage sends a message to all connected clients
func (s *Server) broadcastMessage(message string) {
	s.mutex.Lock()
	s.messages = append(s.messages, message)
	clientsCopy := make([]*Client, 0, len(s.clients))
	for client := range s.clients {
		if client.name != "" {
			clientsCopy = append(clientsCopy, client)
		}
	}
	s.mutex.Unlock()

	for _, client := range clientsCopy {
		s.sendClientMessage(client, message)
	}
}

// removeClient removes a client from the server
func (s *Server) removeClient(client *Client) {
	s.mutex.Lock()
	if client.leftChatFlag {
		s.mutex.Unlock()
		return
	}
	client.leftChatFlag = true
	delete(s.clients, client)
	s.activeConns-- // Decrement active connections count
	s.mutex.Unlock()

	s.broadcastMessage(fmt.Sprintf("[Server] %s left the chat.", client.name))
}

// closeClients closes all client connections
func (s *Server) closeClients() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for client := range s.clients {
		client.leftChatFlag = true
		client.conn.Close()
	}
}

func main() {
	args := os.Args
	port := "2525" // Default port

	// Check if a command-line argument for the port was provided
	if len(args) > 2 {
		fmt.Println("[USAGE]: ./TCPChat [port]")
		return
	} else if len(args) == 2 {
		port = args[1]
	}

	// Create the address string using the specified port
	addr := fmt.Sprintf(":%s", port)

	// Create a new server instance
	server, err := NewServer(addr)
	if err != nil {
		fmt.Println("error creating server:", err)
		return
	}

	// Print server startup message
	fmt.Printf("Chat server started on port %s\n", port)
	fmt.Println("Press Ctrl+C to exit")

	// Start the server
	err = server.start()
	if err != nil {
		fmt.Println("error starting server:", err)
	}
}
