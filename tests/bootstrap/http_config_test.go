package bootstrap_test

import (
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/nixwai/go-game-server/app/bootstrap"
	"github.com/nixwai/go-game-server/app/config"
)

func TestNewAllowsLLMResponseToFinishBeforeWriteTimeout(t *testing.T) {
	llmTimeout := 20 * time.Second
	cfg := config.Config{HTTPAddr: ":0", GoLLMTimeout: llmTimeout}
	app := bootstrap.New(cfg, http.NewServeMux())

	server := reflect.ValueOf(app).Elem().FieldByName("httpServer")
	writeTimeout := time.Duration(server.Elem().FieldByName("WriteTimeout").Int())
	if writeTimeout <= llmTimeout {
		t.Fatalf("write timeout %s must exceed LLM timeout %s", writeTimeout, llmTimeout)
	}
}
