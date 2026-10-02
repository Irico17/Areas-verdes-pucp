package database

import "testing"

func TestCatastroIncompleto(t *testing.T) {
	if !CatastroIncompleto(1, 1, 521, 534) {
		t.Fatal("una siembra de un polígono no es el catastro")
	}
	if !CatastroIncompleto(521, 0, 521, 534) {
		t.Fatal("áreas sin sector no alcanzan")
	}
	if !CatastroIncompleto(0, 0, 521, 534) {
		t.Fatal("la base vacía está incompleta")
	}
	if CatastroIncompleto(521, 534, 521, 534) {
		t.Fatal("el catastro publicado está completo")
	}
	if CatastroIncompleto(527, 540, 521, 534) {
		t.Fatal("filas ficticias de más no lo vuelven incompleto")
	}
}
