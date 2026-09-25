package accesos

import (
	"sync"
	"time"
)

// Limite cuenta intentos por clave en una ventana. Sirve para el login.
type Limite struct {
	max     int
	ventana time.Duration
	mu      sync.Mutex
	hits    map[string][]time.Time
}

func NuevoLimite(max int, ventana time.Duration) *Limite {
	if max < 1 {
		max = 1
	}
	if ventana <= 0 {
		ventana = time.Minute
	}
	return &Limite{max: max, ventana: ventana, hits: map[string][]time.Time{}}
}

// Permitir deja pasar hasta max eventos por clave dentro de la ventana.
func (l *Limite) Permitir(clave string) bool {
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
