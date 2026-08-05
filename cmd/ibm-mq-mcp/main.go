// Command ibm-mq-mcp runs the IBM MQ MCP server over stdio.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/mcpserver"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	server := mcpserver.New()
	if err := mcpserver.RunStdio(ctx, server); err != nil {
		log.Fatalf("mcp server: %v", err)
	}
}
