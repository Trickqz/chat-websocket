package client

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

const (
	defaultAddr  = "ws://localhost:8080/ws"
	writeTimeout = 10 * time.Second
)

type Client struct {
	addr string
}

func New(addr string) *Client {
	if addr == "" {
		addr = defaultAddr
	}
	return &Client{addr: addr}
}

func (c *Client) Start(ctx context.Context) error {
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, c.addr, nil)
	if err != nil {
		return fmt.Errorf("erro ao conectar ao servidor: %w", err)
	}
	defer func() {
		_ = conn.Close()
	}()

	scanner := bufio.NewScanner(os.Stdin)

	nome, err := readNome(scanner)
	if err != nil {
		return err
	}

	fmt.Printf("Olá, %s! Você pode começar a enviar mensagens.\n", nome)

	for {
		select {
		case <-ctx.Done():
			fmt.Println("\nencerrando...")
			return sendClose(conn)
		default:
		}

		fmt.Print("Escreva uma mensagem (ou 'sair'): ")

		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return fmt.Errorf("erro ao ler entrada: %w", err)
			}
			return nil
		}

		msg := strings.TrimSpace(scanner.Text())
		if msg == "" {
			continue
		}
		if msg == "sair" {
			fmt.Println("encerrando...")
			return sendClose(conn)
		}

		fmt.Printf("%s: %s\n", nome, msg)

		if err := sendMessage(conn, msg); err != nil {
			return err
		}

		resp, err := readResponse(conn)
		if err != nil {
			if errors.Is(err, websocket.ErrCloseSent) {
				return nil
			}
			return err
		}

		fmt.Printf("resposta: %s\n", resp)
	}
}

func readNome(scanner *bufio.Scanner) (string, error) {
	fmt.Print("Digite seu nome: ")

	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return "", fmt.Errorf("erro ao ler nome: %w", err)
		}
		return "", errors.New("nome não informado")
	}

	nome := strings.TrimSpace(scanner.Text())
	if nome == "" {
		return "", errors.New("nome não pode ser vazio")
	}

	return nome, nil
}

func sendMessage(conn *websocket.Conn, msg string) error {
	if err := conn.SetWriteDeadline(time.Now().Add(writeTimeout)); err != nil {
		return fmt.Errorf("erro ao definir prazo de escrita: %w", err)
	}
	if err := conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
		return fmt.Errorf("erro ao enviar mensagem: %w", err)
	}
	return nil
}

func readResponse(conn *websocket.Conn) (string, error) {
	_, resp, err := conn.ReadMessage()
	if err != nil {
		return "", fmt.Errorf("erro ao ler resposta do servidor: %w", err)
	}
	return string(resp), nil
}

func sendClose(conn *websocket.Conn) error {
	msg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")
	deadline := time.Now().Add(writeTimeout)

	if err := conn.WriteControl(websocket.CloseMessage, msg, deadline); err != nil &&
		!errors.Is(err, websocket.ErrCloseSent) {
		return fmt.Errorf("erro ao enviar mensagem de fechamento: %w", err)
	}
	return nil
}
