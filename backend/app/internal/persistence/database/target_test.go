package database

import (
	"strings"
	"testing"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/shared/config"
)

func TestConnectionTargetUsesDatabaseURL(t *testing.T) {
	const secret = "clave-super-secreta"
	host, port, dbname, schema, ssl := connectionTarget(config.DatabaseConfig{
		URL:     "postgres://campus:" + secret + "@db.interno:5432/campus_verde?sslmode=require&search_path=public",
		Host:    "localhost",
		Port:    "5432",
		Name:    "areasverdes",
		Schema:  "public",
		SSLMode: "disable",
	})
	if host != "db.interno" || port != "5432" || dbname != "campus_verde" || schema != "public" || ssl != "require" {
		t.Fatalf("destino incorrecto: host=%s port=%s dbname=%s schema=%s ssl=%s", host, port, dbname, schema, ssl)
	}
	joined := strings.Join([]string{host, port, dbname, schema, ssl}, " ")
	if strings.Contains(joined, secret) || strings.Contains(joined, "localhost") || strings.Contains(joined, "areasverdes") {
		t.Fatalf("el destino filtró el secreto o los defaults: %s", joined)
	}
}

func TestConnectionTargetKeywordDSNOmitsPassword(t *testing.T) {
	const secret = "otra-clave"
	host, port, dbname, _, ssl := connectionTarget(config.DatabaseConfig{
		URL:    "host=pg.ejemplo port=5433 user=campus password=" + secret + " dbname=campus_verde sslmode=disable",
		Host:   "localhost",
		Name:   "areasverdes",
		Schema: "public",
	})
	if host != "pg.ejemplo" || port != "5433" || dbname != "campus_verde" || ssl != "disable" {
		t.Fatalf("destino keyword incorrecto: %s %s %s %s", host, port, dbname, ssl)
	}
	if strings.Contains(host+port+dbname+ssl, secret) {
		t.Fatal("el destino incluye la contraseña")
	}
}

func TestConnectionTargetFallsBackWhenURLMissing(t *testing.T) {
	host, port, dbname, _, _ := connectionTarget(config.DatabaseConfig{
		Host: "127.0.0.1",
		Port: "5432",
		Name: "campus_verde",
	})
	if host != "127.0.0.1" || port != "5432" || dbname != "campus_verde" {
		t.Fatalf("fallback incorrecto: %s %s %s", host, port, dbname)
	}
}
