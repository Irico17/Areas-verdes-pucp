package postgres_test

import (
	"testing"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/repository/postgres"
)

func TestIntervaloReservaFusionado(t *testing.T) {
	casos := []struct {
		nombre                      string
		traeInicio, traeFin         bool
		cuerpoInicio, cuerpoFin     string
		guardadoInicio, guardadoFin string
		quiereInicio, quiereFin     string
		ok                          bool
	}{
		{
			nombre:     "solo hora_inicio conserva el fin guardado",
			traeInicio: true, cuerpoInicio: "10:00",
			guardadoInicio: "09:00", guardadoFin: "11:00",
			quiereInicio: "10:00", quiereFin: "11:00", ok: true,
		},
		{
			nombre:  "solo hora_fin conserva el inicio guardado",
			traeFin: true, cuerpoFin: "12:00",
			guardadoInicio: "09:00", guardadoFin: "11:00",
			quiereInicio: "09:00", quiereFin: "12:00", ok: true,
		},
		{
			nombre:  "fin anterior al inicio guardado queda invertido",
			traeFin: true, cuerpoFin: "08:00",
			guardadoInicio: "09:00", guardadoFin: "11:00",
			quiereInicio: "09:00", quiereFin: "08:00", ok: false,
		},
		{
			nombre:     "inicio posterior al fin guardado queda invertido",
			traeInicio: true, cuerpoInicio: "12:00",
			guardadoInicio: "09:00", guardadoFin: "11:00",
			quiereInicio: "12:00", quiereFin: "11:00", ok: false,
		},
		{
			nombre:     "las dos horas del cuerpo se validan entre sí",
			traeInicio: true, traeFin: true,
			cuerpoInicio: "10:00", cuerpoFin: "12:00",
			guardadoInicio: "09:00", guardadoFin: "11:00",
			quiereInicio: "10:00", quiereFin: "12:00", ok: true,
		},
		{
			nombre:     "las dos horas invertidas no usan la fila",
			traeInicio: true, traeFin: true,
			cuerpoInicio: "12:00", cuerpoFin: "10:00",
			guardadoInicio: "09:00", guardadoFin: "11:00",
			quiereInicio: "12:00", quiereFin: "10:00", ok: false,
		},
		{
			nombre:     "horas iguales no son un intervalo",
			traeInicio: true, cuerpoInicio: "11:00",
			guardadoInicio: "09:00", guardadoFin: "11:00",
			quiereInicio: "11:00", quiereFin: "11:00", ok: false,
		},
		{
			nombre:     "una hora vacía no es un intervalo",
			traeInicio: true, cuerpoInicio: "  ",
			guardadoInicio: "09:00", guardadoFin: "11:00",
			quiereInicio: "", quiereFin: "11:00", ok: false,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			inicio, fin, ok := postgres.IntervaloReservaFusionado(
				caso.traeInicio, caso.traeFin,
				caso.cuerpoInicio, caso.cuerpoFin,
				caso.guardadoInicio, caso.guardadoFin,
			)
			if inicio != caso.quiereInicio || fin != caso.quiereFin || ok != caso.ok {
				t.Fatalf("intervalo = %q–%q ok=%v; se esperaba %q–%q ok=%v", inicio, fin, ok, caso.quiereInicio, caso.quiereFin, caso.ok)
			}
		})
	}
}
