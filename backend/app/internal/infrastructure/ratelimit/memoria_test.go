package ratelimit

import (
	"testing"
	"time"
)

func TestLimiteDeLogin(t *testing.T) {
	l := NewMemoriaLimitador(2, time.Minute)
	if !l.Permitir("a") || !l.Permitir("a") {
		t.Fatal("los dos primeros pasan")
	}
	if l.Permitir("a") {
		t.Fatal("el tercero debe cortar")
	}
	if !l.Permitir("b") {
		t.Fatal("otra clave sigue libre")
	}
}

func TestMemoriaLimitadorPodaClavesExpiradas(t *testing.T) {
	lim := NewMemoriaLimitador(5, 20*time.Millisecond)
	memLim, ok := lim.(*memoriaLimitador)
	if !ok {
		t.Fatal("se esperaba tipo *memoriaLimitador")
	}

	// Insertar dos claves
	lim.Permitir("ip1")
	lim.Permitir("ip2")
	if memLim.Len() != 2 {
		t.Fatalf("se esperaban 2 claves activas, hay %d", memLim.Len())
	}

	// Esperar que expire la ventana
	time.Sleep(35 * time.Millisecond)

	// Al acceder a una nueva clave, la poda acotada elimina las expiradas
	lim.Permitir("ip3")
	// Se esperaba que ip1 y ip2 fueran podadas, quedando solo ip3
	if memLim.Len() > 1 {
		// En caso de que la poda inspeccione un subconjunto, otra llamada completa la poda
		lim.Permitir("ip3")
	}
	if memLim.Len() != 1 {
		t.Fatalf("se esperaba que las claves expiradas se podaran (quedando 1), pero hay %d", memLim.Len())
	}
}
