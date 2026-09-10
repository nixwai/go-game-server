package aigo_test

import (
	"encoding/json"
	"testing"

	"github.com/nixwai/go-game-server/app/module/aigo/dto"
)

func TestAnalyzeRequestJSONNumberSize(t *testing.T) {
	raw := `{"size":19,"layout":[[0,1,0]],"player":1}`
	var req dto.AnalyzeRequest
	if err := json.Unmarshal([]byte(raw), &req); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if req.Size == nil {
		t.Fatal("size should not be nil")
	}
	size, ok := req.Size.(float64)
	if !ok {
		t.Fatalf("expected float64, got %T", req.Size)
	}
	if size != 19 {
		t.Fatalf("expected 19, got %v", size)
	}
	if len(req.Layout) != 1 || len(req.Layout[0]) != 3 {
		t.Fatalf("unexpected layout: %v", req.Layout)
	}
	if req.Layout[0][1] != 1 {
		t.Fatalf("expected black stone at [0][1], got %d", req.Layout[0][1])
	}
	if req.Player != 1 {
		t.Fatalf("expected player 1, got %d", req.Player)
	}
}

func TestAnalyzeRequestJSONStringSize(t *testing.T) {
	raw := `{"size":"19","layout":[[0,0,0]],"player":-1}`
	var req dto.AnalyzeRequest
	if err := json.Unmarshal([]byte(raw), &req); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	size, ok := req.Size.(string)
	if !ok {
		t.Fatalf("expected string, got %T", req.Size)
	}
	if size != "19" {
		t.Fatalf("expected 19, got %s", size)
	}
	if req.Player != -1 {
		t.Fatalf("expected player -1, got %d", req.Player)
	}
}

func TestAnalyzeRequestWithKoAndLatest(t *testing.T) {
	raw := `{"size":3,"layout":[[0,0,0],[0,0,0],[0,0,0]],"player":1,"ko":{"sign":-1,"vertex":[1,1]},"latestVertex":[2,2]}`
	var req dto.AnalyzeRequest
	if err := json.Unmarshal([]byte(raw), &req); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if req.Ko == nil {
		t.Fatal("ko should not be nil")
	}
	if req.Ko.Sign != -1 {
		t.Fatalf("expected ko sign -1, got %d", req.Ko.Sign)
	}
	if req.Ko.Vertex[0] != 1 || req.Ko.Vertex[1] != 1 {
		t.Fatalf("expected ko vertex [1,1], got %v", req.Ko.Vertex)
	}
	if req.LatestVertex == nil {
		t.Fatal("latestVertex should not be nil")
	}
	if req.LatestVertex[0] != 2 || req.LatestVertex[1] != 2 {
		t.Fatalf("expected latestVertex [2,2], got %v", req.LatestVertex)
	}
}

func TestAnalyzeRequestWithoutOptionalFields(t *testing.T) {
	raw := `{"size":3,"layout":[[0,0,0],[0,0,0],[0,0,0]],"player":1}`
	var req dto.AnalyzeRequest
	if err := json.Unmarshal([]byte(raw), &req); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if req.Ko != nil {
		t.Fatal("ko should be nil when not provided")
	}
	if req.LatestVertex != nil {
		t.Fatal("latestVertex should be nil when not provided")
	}
}

func TestAnalyzeResponseMoveJSON(t *testing.T) {
	resp := dto.AnalyzeResponse{
		Action: "move",
		Vertex: &dto.GoVertex{5, 3},
	}
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var m map[string]any
	json.Unmarshal(data, &m)
	if m["action"] != "move" {
		t.Fatalf("expected action move, got %v", m["action"])
	}
	v, ok := m["vertex"].([]any)
	if !ok || len(v) != 2 {
		t.Fatalf("expected vertex array, got %v", m["vertex"])
	}
}

func TestAnalyzeResponseEndGameJSON(t *testing.T) {
	resp := dto.AnalyzeResponse{Action: "end_game"}
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var m map[string]any
	json.Unmarshal(data, &m)
	if m["action"] != "end_game" {
		t.Fatalf("expected action end_game, got %v", m["action"])
	}
	if _, exists := m["vertex"]; exists {
		t.Fatal("vertex should be omitted for end_game")
	}
}

func TestGameSettingResponseFields(t *testing.T) {
	resp := dto.GameSettingResponse{
		ActiveModelID:  0,
		AllowAIEndGame: true,
		ModelName:      "gpt-4o-mini",
		IsDefault:      true,
		HasAPIKey:      true,
	}
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	jsonStr := string(data)
	if !contains(jsonStr, "active_model_id") {
		t.Fatal("response should contain active_model_id")
	}
	if !contains(jsonStr, "allow_ai_end_game") {
		t.Fatal("response should contain allow_ai_end_game")
	}
	if !contains(jsonStr, "model_name") {
		t.Fatal("response should contain model_name")
	}
	if !contains(jsonStr, "is_default") {
		t.Fatal("response should contain is_default")
	}
	if !contains(jsonStr, "has_api_key") {
		t.Fatal("response should contain has_api_key")
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
