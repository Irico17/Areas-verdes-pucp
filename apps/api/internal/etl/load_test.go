package etl

import "testing"

func strp(s string) *string { return &s }

func TestDuplicadosPorIndiceConservaMenorId(t *testing.T) {
	areas := []Record{
		{FeatureID: "AV-0001", Codigo: strp("D 20")},
		{FeatureID: "AV-0002", Codigo: strp("G 13")},
		{FeatureID: "AV-0003", Codigo: strp("D 20")}, // repite el índice 0
		{FeatureID: "AV-0004", Codigo: nil},
		{FeatureID: "AV-0005", Codigo: strp("  ")},   // en blanco, no cuenta
		{FeatureID: "AV-0006", Codigo: strp("G 13")}, // repite el índice 1
	}

	dup := duplicadosPorIndice(areas)

	if len(dup) != 2 {
		t.Fatalf("duplicados = %v, se esperaban 2", dup)
	}
	if dup[2] != "D 20" {
		t.Fatalf("índice 2 (AV-0003) debía renombrarse desde %q, quedó %q", "D 20", dup[2])
	}
	if dup[5] != "G 13" {
		t.Fatalf("índice 5 (AV-0006) debía renombrarse desde %q, quedó %q", "G 13", dup[5])
	}
	if _, marcado := dup[0]; marcado {
		t.Fatalf("índice 0 (AV-0001, menor id) no debía marcarse para renombrar")
	}
	if _, marcado := dup[1]; marcado {
		t.Fatalf("índice 1 (AV-0002, menor id) no debía marcarse para renombrar")
	}
}

func TestDuplicadosPorIndiceSinRepetidos(t *testing.T) {
	areas := []Record{
		{FeatureID: "AV-0001", Codigo: strp("A 1")},
		{FeatureID: "AV-0002", Codigo: strp("A 2")},
	}
	if dup := duplicadosPorIndice(areas); len(dup) != 0 {
		t.Fatalf("duplicados = %v, se esperaba ninguno", dup)
	}
}
