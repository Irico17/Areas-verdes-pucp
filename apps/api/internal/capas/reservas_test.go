package capas

import (
	"strings"
	"testing"
)

func TestConsultaReservasOmiteFechasVacias(t *testing.T) {
	q, args := consultaReservas("", "")
	if strings.Contains(q, "::date") {
		t.Fatalf("un filtro vacío no se castea a date: %s", q)
	}
	if len(args) != 0 {
		t.Fatalf("args = %v", args)
	}
	if !strings.Contains(q, "origen = 'ficticio'") {
		t.Fatalf("falta el origen ficticio: %s", q)
	}

	q, args = consultaReservas("2026-09-01", "2026-09-30")
	if !strings.Contains(q, "fecha >= $1::date") || !strings.Contains(q, "fecha <= $2::date") {
		t.Fatalf("filtros con valor: %s", q)
	}
	if len(args) != 2 || args[0] != "2026-09-01" || args[1] != "2026-09-30" {
		t.Fatalf("args = %v", args)
	}

	q, args = consultaReservas("2026-09-01", "")
	if strings.Count(q, "::date") != 1 || !strings.Contains(q, "$1::date") || strings.Contains(q, "$2") {
		t.Fatalf("solo desde: %s", q)
	}
	if len(args) != 1 {
		t.Fatalf("args = %v", args)
	}
}
