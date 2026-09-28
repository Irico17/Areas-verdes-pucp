// Package ratelimit provides in-memory rate limiting.
package ratelimit

import (
	"sync"
	"time"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
)

type memoriaLimitador struct {
	max     int
	ventana time.Duration
	mu      sync.Mutex
	hits    map[string][]time.Time
}

// NewMemoriaLimitador creates a thread-safe in-memory rate limiter.
func NewMemoriaLimitador(max int, ventana time.Duration) contracts.ILimitador {
	if max < 1 {
		max = 1
	}
	if ventana <= 0 {
		ventana = time.Minute
	}
	return &memoriaLimitador{
		max:     max,
		ventana: ventana,
		hits:    make(map[string][]time.Time),
	}
}

// Permitir checks if the key has exceeded max hits within the time window.
func (l *memoriaLimitador) Permitir(clave string) bool {
	ahora := time.Now()
	corte := ahora.Add(-l.ventana)
	l.mu.Lock()
	defer l.mu.Unlock()

	prev := l.hits[clave]
	keep := prev[:0]
	for _, t := range prev {
		if t.After(corte) {
			keep = append(keep, t)
		}
	}
	if len(keep) >= l.max {
		l.hits[clave] = keep
		return false
	}
	l.hits[clave] = append(keep, ahora)
	return true
}
