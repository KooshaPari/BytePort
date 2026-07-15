// Binary mcp-server starts a BytePort MCP server over stdio transport.
//
// Usage:
//
//	go run ./backend/cmd/mcp-server
//
// The server reads JSON-RPC requests from stdin and writes responses to
// stdout.  Logs go to stderr so they don't interfere with the MCP protocol.
//
// Integration test:
//
//	echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}' \
//	  | go run ./backend/cmd/mcp-server
package main

import (
	"log"
	"os"

	"github.com/byteport/api/mcp"
)

func main() {
	// Initialise built-in tools.
	mcp.InitTools()

	// Create the BytePort API client.
	client := mcp.NewClient(nil)
	// client.BaseURL is configured via DefaultConfig(); override from env.
	if url := os.Getenv("BYTEPORT_API_URL"); url != "" {
		client = mcp.NewClient(&mcp.ClientConfig{
			BaseURL:   url,
			AuthToken: os.Getenv("BYTEPORT_API_TOKEN"),
		})
	}

	server := mcp.NewServer(client)
	log.SetOutput(os.Stderr)
	log.SetPrefix("[mcp-server] ")
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	if err := server.Run(); err != nil {
		log.Fatalf("MCP server exited with error: %v", err)
	}
}
