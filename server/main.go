package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	defaultAddr     = ":8080"
	readTimeout     = 60 * time.Second
	writeTimeout    = 10 * time.Second
	shutdownTimeout = 5 * time.Second
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Server struct {
	addr string
}

func New(addr string) *Server {
	if addr == "" {
		addr = defaultAddr
	}
	return &Server{addr: addr}
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("falha no upgrade da conexão", "erro", err)
		return
	}
	defer func() {
		if err := conn.Close(); err != nil {
			slog.Error("erro ao fechar conexão", "erro", err)
		}
	}()

	slog.Info("cliente conectado", "remote_addr", r.RemoteAddr)

	for {
		if err := conn.SetReadDeadline(time.Now().Add(readTimeout)); err != nil {
			slog.Error("erro ao definir prazo de leitura", "erro", err)
			return
		}

		messageType, msg, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				slog.Warn("conexão encerrada inesperadamente", "remote_addr", r.RemoteAddr, "erro", err)
			} else {
				slog.Info("conexão encerrada", "remote_addr", r.RemoteAddr, "erro", err)
			}
			return
		}

		slog.Info("mensagem recebida", "remote_addr", r.RemoteAddr, "conteudo", string(msg))

		if err := conn.SetWriteDeadline(time.Now().Add(writeTimeout)); err != nil {
			slog.Error("erro ao definir prazo de escrita", "erro", err)
			return
		}

		if err := conn.WriteMessage(messageType, msg); err != nil {
			slog.Error("erro ao enviar resposta", "erro", err)
			return
		}
	}
}

func (s *Server) Start(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", s.handleWS)

	httpServer := &http.Server{
		Addr:    s.addr,
		Handler: mux,
	}

	errCh := make(chan error, 1)
	go func() {
		slog.Info("servidor iniciado", "endereco", s.addr)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("erro ao iniciar servidor: %w", err)
			return
		}
		errCh <- nil
	}()

	select {
	case <-ctx.Done():
		slog.Info("encerrando servidor...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("erro ao encerrar servidor: %w", err)
		}
		return nil

	case err := <-errCh:
		return err
	}
}
