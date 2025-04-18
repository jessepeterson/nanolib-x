package http

import (
	"net/http"
	"strings"
	"sync"
)

// Mux represents an HTTP muxer that cannot handle HTTP methods.
type Mux interface {
	Handle(pattern string, handler http.Handler)
	HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request))
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

// MWMethodMux applies middleware when registering HTTP handlers.
// It uses a [MethodMux] to handle per-method dispatching.
type MWMethodMux struct {
	mux Mux

	mu             sync.RWMutex
	middlewares    []func(http.Handler) http.Handler
	methodHandlers map[string]*MethodMux
}

// NewMWMethodMux creates a new MWMethodMux using mux.
func NewMWMethodMux(mux Mux) *MWMethodMux {
	return &MWMethodMux{mux: mux, methodHandlers: make(map[string]*MethodMux)}
}

// Use adds middlewares to be applied when registering HTTP handlers.
func (m *MWMethodMux) Use(middlewares ...func(http.Handler) http.Handler) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.middlewares = append(m.middlewares, middlewares...)
}

// Handle layers middlewares around handler and registers them with pattern.
func (m *MWMethodMux) Handle(pattern string, handler http.Handler) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// assemble middleware in descending order
	for i := len(m.middlewares) - 1; i >= 0; i-- {
		handler = m.middlewares[i](handler)
	}

	parts := strings.SplitN(pattern, " ", 2)
	if len(parts) < 2 {
		// no HTTP method, assign the entire handler with pattern
		m.mux.Handle(pattern, handler)
		return
	}

	pattern = parts[1]

	if m.methodHandlers == nil {
		m.methodHandlers = make(map[string]*MethodMux)
	}

	if m.methodHandlers[pattern] == nil {
		m.methodHandlers[pattern] = NewMethodMux()
		m.mux.Handle(pattern, m.methodHandlers[pattern])
	}

	m.methodHandlers[pattern].Handle(parts[0], handler)
}

// HandleFunc layers middlewares around handler and registers them with pattern.
func (m *MWMethodMux) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	m.Handle(pattern, http.HandlerFunc(handler))
}

// ServeHTTP is a convenience wrapper to use m itself as an HTTP handler.
func (m *MWMethodMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.mux.ServeHTTP(w, r)
}
