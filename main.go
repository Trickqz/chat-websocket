package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/trickqz/chat-websocket/client"
	"github.com/trickqz/chat-websocket/server"
)

func main() {
	if err := run(); err != nil {
		slog.Error("erro fatal", "erro", err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) < 2 {
		printUsage()
		return nil
	}

	cmd := os.Args[1]

	fs := flag.NewFlagSet(cmd, flag.ExitOnError)
	addr := fs.String("addr", "", "endereço para conexão (client) ou escuta (server)")
	if err := fs.Parse(os.Args[2:]); err != nil {
		return fmt.Errorf("erro ao processar argumentos: %w", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	switch cmd {
	case "server":
		return server.New(*addr).Start(ctx)
	case "client":
		return client.New(*addr).Start(ctx)
	default:
		printUsage()
		return fmt.Errorf("opção inválida: %s", cmd)
	}
}

func printUsage() {
	fmt.Println("Uso:")
	fmt.Println("  go run . server [-addr :8080]")
	fmt.Println("  go run . client [-addr ws://localhost:8080/ws]")
}
