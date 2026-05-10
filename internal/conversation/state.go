// Padrão: State Machine.
// O LLM é stateless. Este pacote mantém o histórico completo em memória
// e o reenvia integralmente a cada inferência.
package conversation

import (
	"sync"

	ollamaapi "github.com/ollama/ollama/api"
)

type State struct {
	mu       sync.RWMutex
	messages []ollamaapi.Message
}

func New() *State {
	return &State{messages: make([]ollamaapi.Message, 0, 32)}
}

// Append adiciona uma mensagem ao histórico de forma thread-safe.
func (s *State) Append(msg ollamaapi.Message) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messages = append(s.messages, msg)
}

// Snapshot retorna uma cópia do histórico atual para envio ao Ollama.
func (s *State) Snapshot() []ollamaapi.Message {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]ollamaapi.Message, len(s.messages))
	copy(out, s.messages)
	return out
}

// Len retorna o número atual de mensagens.
func (s *State) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.messages)
}