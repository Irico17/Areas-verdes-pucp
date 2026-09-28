package ioc_test

import (
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/cmd/ioc"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/routes"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/shared/config"
)

func TestBuildContainer(t *testing.T) {
	container, err := ioc.BuildContainer()
	if err != nil {
		t.Fatalf("BuildContainer() failed: %v", err)
	}

	err = container.Invoke(func(router *routes.Router, engine *gin.Engine, cfg *config.Config) {
		if router == nil {
			t.Fatal("expected router to be provided")
		}
		if engine == nil {
			t.Fatal("expected engine to be provided")
		}
		if cfg == nil {
			t.Fatal("expected cfg to be provided")
		}
		router.Setup()
	})
	if err != nil {
		t.Fatalf("container.Invoke() failed: %v", err)
	}
}
