// Package ratelimit provides in-memory rate limiting.
package ratelimit

import (
	"sync"
	"time"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/shared/config"
)

type memoriaLimitador struct {
	max     int
	ventana time.Duration
	mu      sync.Mutex
	hits    map[string][]time.Time
}

// NewLimitadorFromConfig creates an ILimitador configured from application settings.
func NewLimitadorFromConfig(cfg *config.Config) contracts.ILimitador {
	max := cfg.Seguridad.LoginMax
	if max < 1 {
		max = 8
	}
	ventana := cfg.Seguridad.LoginVentana
	if ventana <= 0 {
		ventana = time.Minute
	}
	return NewMemoriaLimitador(max, ventana)
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

	// Poda de costo acotado: inspeccionar hasta 8 claves y eliminar las que no tienen hits en la ventana
	pruned := 0
	for k, times := range l.hits {
		if pruned >= 8 {
			break
		}
		if len(times) == 0 || !times[len(times)-1].After(corte) {
			delete(l.hits, k)
			pruned++
		}
	}

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

// Len returns the number of active tracked keys in memory.
func (l *memoriaLimitador) Len() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.hits)
}
