package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

const (
	message = "Hello, websocket"
)

func connect(domain, address, timeout string) (*websocket.Conn, error) {
	var conn *websocket.Conn
	rawQuery := fmt.Sprintf("delay=%s", timeout)
	u := url.URL{Scheme: "ws", Host: address, Path: "/", RawQuery: rawQuery}
	log.Printf("Connecting using websocket: url=%s, host=%s", u.String(), domain)

	dialer := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 45 * time.Second,
	}

	c, resp, err := dialer.Dial(u.String(), http.Header{"Host": {domain}})
	if err == nil {
		log.Printf("WebSocket connection established.")
		conn = c
	}
	if resp == nil {
		// We don't have an HTTP response, probably TCP errors.
		return nil, fmt.Errorf("failed to connect to %s: %v", domain, err)
	}

	_, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, fmt.Errorf("connection failed: %v. Failed to read HTTP response: %v", err, readErr)
	}
	return conn, nil
}

func main() {
	conn, err := connect(os.Args[1], os.Args[2], os.Args[3])
	if err != nil {
		log.Fatalf("Error: %q", err)
	}
	defer conn.Close()

	// Send a message.
	log.Printf("Sending message %q to server.", message)
	for i := 0; i < 1000; i++ {
		if err = conn.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
			log.Fatalf("Error: %q", err)
		}
		log.Printf("Message sent.")

		// Read back the echoed message and compared with sent.
		_, recv, err := conn.ReadMessage()
		if err != nil {
			log.Fatalf("Error: %q", err)
		} else if strings.HasPrefix(string(recv), message) {
			log.Printf("Received message %q from echo server.", recv)
		}
		time.Sleep(100 * time.Millisecond)
	}
}
