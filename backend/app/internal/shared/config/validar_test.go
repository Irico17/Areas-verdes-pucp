package config

import (
	"strings"
	"testing"
)

func TestValidarClaves(t *testing.T) {
	casos := []struct {
		nombre        string
		appEnv        string
		devPassword   string
		dbPassword    string
		dbURL         string
		esperaError   bool
		textoEsperado string
	}{
		// Casos requeridos por la tarea B3:
		{
			nombre:        "produccion + pando-local -> error",
			appEnv:        "produccion",
			devPassword:   "pando-local",
			esperaError:   true,
			textoEsperado: "no puede ser una clave de laboratorio",
		},
		{
			nombre:        "produccion + campus-lab -> error",
			appEnv:        "produccion",
			devPassword:   "campus-lab",
			esperaError:   true,
			textoEsperado: "no puede ser una clave de laboratorio",
		},
		{
			nombre:        "produccion + clave corta -> error",
			appEnv:        "produccion",
			devPassword:   "corta12345",
			esperaError:   true,
			textoEsperado: "al menos 16 caracteres",
		},
		{
			nombre:      "produccion + clave válida -> ok",
			appEnv:      "produccion",
			devPassword: "produccion-local-demo-2026",
			esperaError: false,
		},
		{
			nombre:      "develop + pando-local -> ok",
			appEnv:      "develop",
			devPassword: "pando-local",
			esperaError: false,
		},
		{
			nombre:      "sin APP_ENV + pando-local -> ok",
			appEnv:      "",
			devPassword: "pando-local",
			esperaError: false,
		},

		// Casos adicionales de robustez (espacios alrededor):
		{
			nombre:        "produccion + pando-local con espacios -> error",
			appEnv:        "produccion",
			devPassword:   "   pando-local   ",
			esperaError:   true,
			textoEsperado: "no puede ser una clave de laboratorio",
		},
		{
			nombre:        "produccion + campus-lab con espacios -> error",
			appEnv:        "produccion",
			devPassword:   "\tcampus-lab\n",
			esperaError:   true,
			textoEsperado: "no puede ser una clave de laboratorio",
		},
		{
			nombre:        "produccion + clave corta con espacios que suman 16 -> error",
			appEnv:        "produccion",
			devPassword:   "   corta123   ",
			esperaError:   true,
			textoEsperado: "al menos 16 caracteres",
		},

		// Contraseña de Postgres en produccion:
		{
			nombre:        "produccion + postgres pando-local en DATABASE_URL -> error",
			appEnv:        "produccion",
			devPassword:   "produccion-local-demo-2026",
			dbURL:         "postgres://campus:pando-local@db:5432/campus_verde_produccion",
			esperaError:   true,
			textoEsperado: "la clave de Postgres no puede ser una clave de laboratorio",
		},
		{
			nombre:        "produccion + postgres campus-lab en DATABASE_URL -> error",
			appEnv:        "produccion",
			devPassword:   "produccion-local-demo-2026",
			dbURL:         "postgres://campus:campus-lab@db:5432/campus_verde_produccion",
			esperaError:   true,
			textoEsperado: "la clave de Postgres no puede ser una clave de laboratorio",
		},
		{
			nombre:        "produccion + postgres corta en DATABASE_URL -> error",
			appEnv:        "produccion",
			devPassword:   "produccion-local-demo-2026",
			dbURL:         "postgres://campus:corta123@db:5432/campus_verde_produccion",
			esperaError:   true,
			textoEsperado: "la clave de Postgres debe tener al menos 16 caracteres",
		},
		{
			nombre:      "produccion + postgres válida en DATABASE_URL -> ok",
			appEnv:      "produccion",
			devPassword: "produccion-local-demo-2026",
			dbURL:       "postgres://campus:campus-produccion-local@db:5432/campus_verde_produccion",
			esperaError: false,
		},
	}

	for _, tc := range casos {
		t.Run(tc.nombre, func(t *testing.T) {
			t.Setenv("APP_ENV", tc.appEnv)
			t.Setenv("CAMPUS_DEV_PASSWORD", tc.devPassword)
			t.Setenv("DATABASE_URL", tc.dbURL)
			t.Setenv("DATABASE_PASSWORD", tc.dbPassword)

			cfg := New()
			err := cfg.Validar()

			if tc.esperaError {
				if err == nil {
					t.Fatalf("se esperaba error y se obtuvo nil (env=%q, pass=%q)", tc.appEnv, tc.devPassword)
				}
				if tc.textoEsperado != "" && !strings.Contains(err.Error(), tc.textoEsperado) {
					t.Fatalf("se esperaba mensaje conteniendo %q, pero se obtuvo %q", tc.textoEsperado, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("no se esperaba error, se obtuvo: %v (env=%q, pass=%q)", err, tc.appEnv, tc.devPassword)
				}
			}
		})
	}
}
