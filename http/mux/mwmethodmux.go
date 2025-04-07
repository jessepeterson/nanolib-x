package http

import (
	"net/http"
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
func (m *MWMethodMux) Handle(pattern string, handler http.Handler, methods ...string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// assemble middleware in descending order
	for i := len(m.middlewares) - 1; i >= 0; i-- {
		handler = m.middlewares[i](handler)
	}

	if len(methods) < 1 {
		// no methods, assign the entire handler on pattern
		m.mux.Handle(pattern, handler)
		return
	}

	if m.methodHandlers == nil {
		m.methodHandlers = make(map[string]*MethodMux)
	}

	if m.methodHandlers[pattern] == nil {
		m.methodHandlers[pattern] = NewMethodMux()
		m.mux.Handle(pattern, m.methodHandlers[pattern])
	}

	for _, method := range methods {
		m.methodHandlers[pattern].Handle(method, handler)
	}
}

// HandleFunc layers middlewares around handler and registers them with pattern.
func (m *MWMethodMux) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request), methods ...string) {
	m.Handle(pattern, http.HandlerFunc(handler), methods...)
}

// ServeHTTP is a convenience wrapper to use m itself as an HTTP handler.
func (m *MWMethodMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.mux.ServeHTTP(w, r)
}
