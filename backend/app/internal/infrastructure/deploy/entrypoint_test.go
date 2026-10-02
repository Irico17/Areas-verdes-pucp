package deploy_test

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func createFakeBinary(t *testing.T, dir, name, script string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	content := fmt.Sprintf("#!/bin/sh\n%s\n", script)
	if err := os.WriteFile(p, []byte(content), 0o755); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestDockerEntrypoint(t *testing.T) {
	t.Setenv("CAMPUS_DEV_PASSWORD", "pando-local")
	t.Setenv("MIGRATIONS_DIR", t.TempDir())
	entrypointPath, err := filepath.Abs(filepath.Join("..", "..", "..", "..", "docker-entrypoint.sh"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(entrypointPath); err != nil {
		t.Skipf("no se encontró docker-entrypoint.sh: %v", err)
	}

	t.Run("BaseConDatos_Codigo10_NoEjecutaETL_CreaEtlDone", func(t *testing.T) {
		tmpDir := t.TempDir()
		binDir := filepath.Join(tmpDir, "bin")
		if err := os.MkdirAll(binDir, 0o755); err != nil {
			t.Fatal(err)
		}
		dataDir := filepath.Join(tmpDir, "data")
		etlDone := filepath.Join(dataDir, ".etl-done")

		migrateBin := createFakeBinary(t, binDir, "migrate", `
if [ "$1" = "-catastro-incompleto" ]; then
  exit 10
fi
if [ "$1" = "-necesita-etl" ]; then
  echo "BD con datos en: areas_verdes"
  exit 10
fi
exit 0
`)
		etlBin := createFakeBinary(t, binDir, "etl", `echo "ETL_EJECUTADO" >> "`+tmpDir+`/log"; exit 0`)
		apiBin := createFakeBinary(t, binDir, "api", `echo "API_EJECUTADA" >> "`+tmpDir+`/log"; exit 0`)

		cmd := exec.Command("/bin/sh", entrypointPath)
		cmd.Env = append(os.Environ(),
			"MIGRATE_BIN="+migrateBin,
			"ETL_BIN="+etlBin,
			"API_BIN="+apiBin,
			"ETL_DONE_FILE="+etlDone,
			"DATA_DIR="+dataDir,
		)

		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("fallo al ejecutar entrypoint: %v\nSalida: %s", err, string(out))
		}

		if !strings.Contains(string(out), "Catastro completo: no se vuelve a cargar") {
			t.Fatalf("esperaba mensaje de catastro completo, obtuve: %s", string(out))
		}

		if _, err := os.Stat(etlDone); os.IsNotExist(err) {
			t.Fatalf("no se creó .etl-done")
		}

		logBytes, _ := os.ReadFile(filepath.Join(tmpDir, "log"))
		logContent := string(logBytes)
		if strings.Contains(logContent, "ETL_EJECUTADO") {
			t.Fatalf("etl no debió ejecutarse")
		}
		if !strings.Contains(logContent, "API_EJECUTADA") {
			t.Fatalf("api debió ejecutarse")
		}
	})

	t.Run("BaseVacia_Codigo0_EjecutaETL_CreaEtlDone", func(t *testing.T) {
		tmpDir := t.TempDir()
		binDir := filepath.Join(tmpDir, "bin")
		if err := os.MkdirAll(binDir, 0o755); err != nil {
			t.Fatal(err)
		}
		dataDir := filepath.Join(tmpDir, "data")
		etlDone := filepath.Join(dataDir, ".etl-done")

		migrateBin := createFakeBinary(t, binDir, "migrate", `
if [ "$1" = "-necesita-etl" ]; then
  echo "BD vacía"
  exit 0
fi
exit 0
`)
		etlBin := createFakeBinary(t, binDir, "etl", `echo "ETL_EJECUTADO" >> "`+tmpDir+`/log"; exit 0`)
		apiBin := createFakeBinary(t, binDir, "api", `echo "API_EJECUTADA" >> "`+tmpDir+`/log"; exit 0`)

		cmd := exec.Command("/bin/sh", entrypointPath)
		cmd.Env = append(os.Environ(),
			"MIGRATE_BIN="+migrateBin,
			"ETL_BIN="+etlBin,
			"API_BIN="+apiBin,
			"ETL_DONE_FILE="+etlDone,
			"DATA_DIR="+dataDir,
		)

		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("fallo al ejecutar entrypoint: %v\nSalida: %s", err, string(out))
		}

		if _, err := os.Stat(etlDone); os.IsNotExist(err) {
			t.Fatalf("no se creó .etl-done tras carga inicial")
		}

		logBytes, _ := os.ReadFile(filepath.Join(tmpDir, "log"))
		logContent := string(logBytes)
		if !strings.Contains(logContent, "ETL_EJECUTADO") {
			t.Fatalf("etl debió ejecutarse")
		}
		if !strings.Contains(logContent, "API_EJECUTADA") {
			t.Fatalf("api debió ejecutarse")
		}
	})

	t.Run("CatastroIncompleto_EjecutaLoteSinETL", func(t *testing.T) {
		tmpDir := t.TempDir()
		binDir := filepath.Join(tmpDir, "bin")
		if err := os.MkdirAll(binDir, 0o755); err != nil {
			t.Fatal(err)
		}
		dataDir := filepath.Join(tmpDir, "data")
		if err := os.MkdirAll(dataDir, 0o755); err != nil {
			t.Fatal(err)
		}
		etlDone := filepath.Join(dataDir, ".etl-done")
		if err := os.WriteFile(etlDone, []byte("previo"), 0o644); err != nil {
			t.Fatal(err)
		}

		migrateBin := createFakeBinary(t, binDir, "migrate", `
if [ "$1" = "-catastro-incompleto" ]; then
  echo "catastro visible: areas_con_geom=1 zonas_con_sector=1"
  exit 0
fi
if [ "$1" = "-necesita-etl" ]; then
  echo "BD con datos en: areas_verdes (1)"
  exit 10
fi
exit 0
`)
		etlBin := createFakeBinary(t, binDir, "etl", `echo "ETL_EJECUTADO" >> "`+tmpDir+`/log"; exit 0`)
		loteBin := createFakeBinary(t, binDir, "etl-lote", `echo "LOTE_EJECUTADO" >> "`+tmpDir+`/log"; exit 0`)
		apiBin := createFakeBinary(t, binDir, "api", `echo "API_EJECUTADA" >> "`+tmpDir+`/log"; exit 0`)

		cmd := exec.Command("/bin/sh", entrypointPath)
		cmd.Env = append(os.Environ(),
			"MIGRATE_BIN="+migrateBin,
			"ETL_BIN="+etlBin,
			"ETL_LOTE_BIN="+loteBin,
			"API_BIN="+apiBin,
			"ETL_DONE_FILE="+etlDone,
			"DATA_DIR="+dataDir,
		)

		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("fallo al ejecutar entrypoint: %v\nSalida: %s", err, string(out))
		}
		if !strings.Contains(string(out), "sin TRUNCATE") {
			t.Fatalf("esperaba upsert sin TRUNCATE, obtuve: %s", string(out))
		}
		logBytes, _ := os.ReadFile(filepath.Join(tmpDir, "log"))
		logContent := string(logBytes)
		if strings.Contains(logContent, "ETL_EJECUTADO") {
			t.Fatalf("la carga inicial no debe truncar un catastro que ya tiene filas")
		}
		if !strings.Contains(logContent, "LOTE_EJECUTADO") {
			t.Fatalf("etl-lote debió completar el catastro: %s", logContent)
		}
		if !strings.Contains(logContent, "API_EJECUTADA") {
			t.Fatalf("api debió ejecutarse")
		}
	})

	t.Run("ErrorEnNecesitaETL_FallaLoudly_NoEjecutaETL", func(t *testing.T) {
		tmpDir := t.TempDir()
		binDir := filepath.Join(tmpDir, "bin")
		if err := os.MkdirAll(binDir, 0o755); err != nil {
			t.Fatal(err)
		}
		dataDir := filepath.Join(tmpDir, "data")
		etlDone := filepath.Join(dataDir, ".etl-done")

		migrateBin := createFakeBinary(t, binDir, "migrate", `
if [ "$1" = "-necesita-etl" ]; then
  echo "error de conexión fatal" >&2
  exit 1
fi
exit 0
`)
		etlBin := createFakeBinary(t, binDir, "etl", `echo "ETL_EJECUTADO" >> "`+tmpDir+`/log"; exit 0`)
		apiBin := createFakeBinary(t, binDir, "api", `echo "API_EJECUTADA" >> "`+tmpDir+`/log"; exit 0`)

		cmd := exec.Command("/bin/sh", entrypointPath)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		cmd.Env = append(os.Environ(),
			"MIGRATE_BIN="+migrateBin,
			"ETL_BIN="+etlBin,
			"API_BIN="+apiBin,
			"ETL_DONE_FILE="+etlDone,
			"DATA_DIR="+dataDir,
		)

		err := cmd.Run()
		if err == nil {
			t.Fatalf("entrypoint debía fallar")
		}
		if !strings.Contains(stderr.String(), "ERROR al verificar si la base necesita ETL") {
			t.Fatalf("esperaba mensaje de error, obtuve: %s", stderr.String())
		}
		if _, err := os.Stat(etlDone); !os.IsNotExist(err) {
			t.Fatalf(".etl-done no debía existir")
		}
	})

	t.Run("DataNoEscribible_MensajeDeChown", func(t *testing.T) {
		if os.Geteuid() == 0 {
			t.Skip("root ignora el modo 0555; el mensaje se cubre con un usuario sin privilegios")
		}
		tmpDir := t.TempDir()
		dataDir := filepath.Join(tmpDir, "data")
		if err := os.MkdirAll(dataDir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(dataDir, 0o555); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = os.Chmod(dataDir, 0o755) })

		cmd := exec.Command("/bin/sh", entrypointPath)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		cmd.Env = append(os.Environ(), "DATA_DIR="+dataDir)
		if err := cmd.Run(); err == nil {
			t.Fatal("entrypoint debía fallar si /data no es escribible")
		}
		if !strings.Contains(stderr.String(), "no es escribible") || !strings.Contains(stderr.String(), "10001") {
			t.Fatalf("esperaba el aviso de chown, obtuve: %s", stderr.String())
		}
	})
}

func TestDockerEntrypoint_RechazoClavesProduccion(t *testing.T) {
	entrypointPath, err := filepath.Abs(filepath.Join("..", "..", "..", "..", "docker-entrypoint.sh"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(entrypointPath); err != nil {
		t.Skipf("no se encontró docker-entrypoint.sh: %v", err)
	}

	casos := []struct {
		nombre        string
		appEnv        string
		devPassword   string
		dbURL         string
		esperaError   bool
		textoEsperado string
	}{
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
		{
			nombre:        "produccion + clave válida + DATABASE_URL pando-local -> error",
			appEnv:        "produccion",
			devPassword:   "produccion-local-demo-2026",
			dbURL:         "postgres://campus:pando-local@db:5432/campus_verde_produccion",
			esperaError:   true,
			textoEsperado: "no puede ser una clave de laboratorio",
		},
		{
			nombre:        "produccion + clave válida + DATABASE_URL campus-lab -> error",
			appEnv:        "produccion",
			devPassword:   "produccion-local-demo-2026",
			dbURL:         "postgres://campus:campus-lab@db:5432/campus_verde_produccion",
			esperaError:   true,
			textoEsperado: "no puede ser una clave de laboratorio",
		},
		{
			nombre:        "produccion + clave válida + DATABASE_URL corta -> error",
			appEnv:        "produccion",
			devPassword:   "produccion-local-demo-2026",
			dbURL:         "postgres://campus:corta123@db:5432/campus_verde_produccion",
			esperaError:   true,
			textoEsperado: "al menos 16 caracteres",
		},
		{
			nombre:      "produccion + clave válida + DATABASE_URL válida -> ok",
			appEnv:      "produccion",
			devPassword: "produccion-local-demo-2026",
			dbURL:       "postgres://campus:campus-produccion-local@db:5432/campus_verde_produccion",
			esperaError: false,
		},
	}

	for _, tc := range casos {
		t.Run(tc.nombre, func(t *testing.T) {
			tmpDir := t.TempDir()
			binDir := filepath.Join(tmpDir, "bin")
			if err := os.MkdirAll(binDir, 0o755); err != nil {
				t.Fatal(err)
			}
			dataDir := filepath.Join(tmpDir, "data")
			etlDone := filepath.Join(dataDir, ".etl-done")
			migrationsDir := filepath.Join(tmpDir, "migrations")
			if err := os.MkdirAll(migrationsDir, 0o755); err != nil {
				t.Fatal(err)
			}

			migrateBin := createFakeBinary(t, binDir, "migrate", `
if [ "$1" = "-catastro-incompleto" ] || [ "$1" = "-necesita-etl" ]; then
  exit 10
fi
exit 0
`)
			etlBin := createFakeBinary(t, binDir, "etl", `exit 0`)
			apiBin := createFakeBinary(t, binDir, "api", `echo "API_OK" >> "`+tmpDir+`/log"; exit 0`)

			cmd := exec.Command("/bin/sh", entrypointPath)
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			cmd.Env = append(os.Environ(),
				"APP_ENV="+tc.appEnv,
				"CAMPUS_DEV_PASSWORD="+tc.devPassword,
				"DATABASE_URL="+tc.dbURL,
				"MIGRATIONS_DIR="+migrationsDir,
				"MIGRATE_BIN="+migrateBin,
				"ETL_BIN="+etlBin,
				"API_BIN="+apiBin,
				"ETL_DONE_FILE="+etlDone,
				"DATA_DIR="+dataDir,
			)

			runErr := cmd.Run()
			if tc.esperaError {
				if runErr == nil {
					t.Fatalf("se esperaba que entrypoint fallara (env=%q, pass=%q)", tc.appEnv, tc.devPassword)
				}
				if tc.textoEsperado != "" && !strings.Contains(stderr.String(), tc.textoEsperado) {
					t.Fatalf("se esperaba mensaje conteniendo %q, se obtuvo stderr: %s", tc.textoEsperado, stderr.String())
				}
				if strings.Contains(stderr.String(), "pando-local") || strings.Contains(stderr.String(), "campus-lab") {
					t.Fatalf("el entrypoint no debe imprimir la contraseña sensible en stderr: %s", stderr.String())
				}
			} else {
				if runErr != nil {
					t.Fatalf("entrypoint falló inesperadamente: %v\nStderr: %s\nStdout: %s", runErr, stderr.String(), stdout.String())
				}
				logBytes, _ := os.ReadFile(filepath.Join(tmpDir, "log"))
				if !strings.Contains(string(logBytes), "API_OK") {
					t.Fatalf("la API debió haberse ejecutado")
				}
			}
		})
	}
}
