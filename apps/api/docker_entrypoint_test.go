package main_test

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
	entrypointPath, err := filepath.Abs("docker-entrypoint.sh")
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

		if !strings.Contains(string(out), "BD con datos: no se ejecuta la carga inicial") {
			t.Fatalf("esperaba mensaje de BD con datos, obtuve: %s", string(out))
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
}
