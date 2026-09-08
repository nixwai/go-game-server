package response_test

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/nixwai/go-game-server/app/response"
)

type assertErr struct{}

func (assertErr) Error() string { return "secret stack detail" }

func TestWriteErrorHidesInternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Set("request_id", "req-1")
	response.WriteError(c, response.NewError(response.CodeInternal, "服务器内部错误", assertErr{}))
	if recorder.Code != 200 {
		t.Fatalf("status: %d", recorder.Code)
	}
	if recorder.Body.String() == "" || contains(recorder.Body.String(), "secret") {
		t.Fatalf("unexpected body: %s", recorder.Body.String())
	}
}

func contains(value, needle string) bool {
	for i := 0; i+len(needle) <= len(value); i++ {
		if value[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
