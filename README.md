# chat-websocket

Um chat simples via WebSocket escrito em Go, com um servidor que recebe mensagens e responde (eco) e um cliente de linha de comando para enviar e receber mensagens em tempo real.

Projeto de estudo focado em `net/http`, [`gorilla/websocket`](https://github.com/gorilla/websocket) e uso de `context` para encerramento gracioso.

## Funcionalidades

- Servidor WebSocket com endpoint `/ws`
- Cliente CLI interativo via terminal
- Encerramento gracioso com `Ctrl+C` (`SIGINT`/`SIGTERM`)
- Timeouts de leitura/escrita para evitar conexões penduradas
- Logs estruturados no servidor (`log/slog`)
- Endereço configurável via flag `-addr`

## Requisitos

- Go 1.21 ou superior
- Módulo [`github.com/gorilla/websocket`](https://github.com/gorilla/websocket)

## Instalação

```bash
git clone https://github.com/trickqz/chat-websocket.git
cd chat-websocket
go mod tidy
```

## Como usar

Em um terminal, inicie o servidor:

```bash
go run . server
```

Em outro terminal, inicie um cliente:

```bash
go run . client
```

Digite seu nome e comece a enviar mensagens. Digite `sair` para encerrar o cliente.

### Endereço customizado

Por padrão o servidor escuta em `:8080` e o cliente conecta em `ws://localhost:8080/ws`. Para mudar:

```bash
go run . server -addr :9090
go run . client -addr ws://localhost:9090/ws
```

## Estrutura do projeto

```
.
├── main.go            # ponto de entrada, parsing de flags e sinais
├── client/
│   └── client.go       # lógica do cliente CLI
└── server/
    └── server.go        # lógica do servidor WebSocket
```

## Limitações conhecidas

- O cliente é de eco simples: o servidor apenas devolve a mesma mensagem recebida, não há broadcast entre múltiplos clientes.
- A leitura do terminal (`bufio.Scanner`) é bloqueante — o encerramento via `Ctrl+C` só é percebido entre uma mensagem e outra.

## Possíveis evoluções

- Broadcast entre múltiplos clientes conectados (hub central)
- Mensagens estruturadas em JSON (autor, horário, tipo de evento)
- Ping/Pong para detectar conexões mortas
- Testes automatizados

## Licença

Este projeto está sob a licença MIT.
