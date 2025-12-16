<div align="center">
  <img src="assets/logo.png" alt="Mailman Logo" width="250"/>
  <h1>Mailman</h1>
  <p>
    <b>Simplifique e unifique o processamento de mensagens em Go.</b>
  </p>
</div>

---

## 📝 Sobre

**Mailman** é uma biblioteca projetada para abstrair a complexidade de consumir mensagens de múltiplos sistemas (como SQS, Kafka, RabbitMQ, Pub/Sub, etc.). Ela oferece uma interface extensível, leve e idiomática, permitindo que desenvolvedores integrem múltiplos backends através de um único fluxo de consumo unificado.

Com o Mailman, você foca na **lógica de negócio** enquanto a biblioteca gerencia a orquestração, concorrência e ciclo de vida das mensagens.

## ✨ Funcionalidades

- **Interface Unificada**: API consistente para qualquer sistema de mensageria.
- **Suporte a Middlewares**: Adicione logs, métricas, tracing e tratamento de erros de forma global ou por handler.
- **Gerenciamento de Concorrência**: Controle granular de workers e tamanho de buffer por consumidor.
- **Contexto Rico**: Acesso a metadados da execução (Handler Name, PID, Timestamp) diretamente no `context.Context`.
- **Extensibilidade**: Interface `Router` simples para adicionar suporte a novos backends.

## 🚀 Instalação

Adicione o Mailman ao seu projeto Go:

```bash
go get github.com/guilhermealvess/mailman
```

## 💡 Exemplo de Uso

Abaixo está um exemplo básico utilizando o `generic.Router` para processamento em memória, ilustrando a simplicidade da API:

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/guilhermealvess/mailman"
	"github.com/guilhermealvess/mailman/generic"
)

type Notification struct {
	UserID  int    `json:"user_id"`
	Message string `json:"message"`
}

func main() {
	// 1. Inicialize o Manager
	manager := mailman.New()

	// 2. Adicione Middlewares globais (opcional)
	manager.Use(func(next mailman.HandlerFunction) mailman.HandlerFunction {
		return func(ctx context.Context, event mailman.Event) error {
			fmt.Println("[Log] Iniciando processamento...")
			return next(ctx, event)
		}
	})

	// 3. Crie um Router (Consumer)
	// O generic.NewGenericRouter é útil para testes ou canais em memória
	handler := func(ctx context.Context, event mailman.Event) error {
		var notif Notification
		if err := event.Bind(&notif); err != nil {
			return err
		}
		fmt.Printf("Enviando notificação para User %d: %s\n", notif.UserID, notif.Message)
		return nil
	}
	
	router, channel := generic.NewGenericRouter[Notification](handler)

	// 4. Registre o Router no Manager
	manager.Register("notification-service", router)

	// Simulação de produção de mensagens
	go func() {
		for {
			channel <- Notification{UserID: 1, Message: "Bem-vindo ao Mailman!"}
			time.Sleep(2 * time.Second)
		}
	}()

	// 5. Inicie o Manager (bloqueante)
	fmt.Println("Mailman rodando...")
	manager.Run()
}
```

## 🤝 Contribuindo

Contribuições são super bem-vindas! Se você tiver uma ideia de melhoria, correção de bug ou implementação de um novo adaptador (Router), sinta-se à vontade para abrir uma **Issue** ou enviar um **Pull Request**.

## 📄 Licença

Este projeto está licenciado sob a licença MIT - veja o arquivo [LICENSE](LICENSE) para mais detalhes.
