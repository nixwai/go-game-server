package aigo

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/nixwai/go-game-server/app/module/aigo/dto"
)

// buildSystemPrompt 根据 allowEndGame 构造系统提示词。
func buildSystemPrompt(allowEndGame bool) string {
	if allowEndGame {
		return `You are a Go (Weiqi) game assistant. Given the current board state, respond with ONLY a JSON object: {"action":"move","vertex":[row,col]} to place a stone, or {"action":"end_game"} if the game should end by counting. Do not include any other text.`
	}
	return `You are a Go (Weiqi) game assistant. Given the current board state, you must always return a move. Do not suggest ending the game. Respond with ONLY a JSON object: {"action":"move","vertex":[row,col]}. Do not include any other text.`
}

// buildUserPrompt 将棋局数据序列化为 LLM 可读的文本。
func buildUserPrompt(req dto.AnalyzeRequest) (string, error) {
	size, err := parseSize(req.Size)
	if err != nil {
		return "", err
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Board size: %dx%d\n", size, size)

	playerStr := "Black"
	if req.Player < 0 {
		playerStr = "White"
	}
	fmt.Fprintf(&b, "Current player: %s (%d)\n", playerStr, req.Player)

	if req.Ko != nil {
		fmt.Fprintf(&b, "Ko: sign=%d at [%d,%d]\n", req.Ko.Sign, req.Ko.Vertex[0], req.Ko.Vertex[1])
	}
	if req.LatestVertex != nil {
		fmt.Fprintf(&b, "Last move: [%d,%d]\n", req.LatestVertex[0], req.LatestVertex[1])
	}
	b.WriteString("\n")

	b.WriteString("  ")
	for c := 0; c < size; c++ {
		if c == size-1 {
			fmt.Fprintf(&b, "%d\n", c)
		} else {
			fmt.Fprintf(&b, "%d ", c)
		}
	}

	for r := 0; r < size; r++ {
		fmt.Fprintf(&b, "%d ", r)
		for c := 0; c < size; c++ {
			cell := "."
			if r < len(req.Layout) && c < len(req.Layout[r]) {
				switch req.Layout[r][c] {
				case 1:
					cell = "X"
				case -1:
					cell = "O"
				}
			}
			if c == size-1 {
				b.WriteString(cell + "\n")
			} else {
				b.WriteString(cell + " ")
			}
		}
	}

	return b.String(), nil
}

// parseSize 将 number 或 string 类型的棋盘尺寸统一转为 int。
func parseSize(size any) (int, error) {
	switch v := size.(type) {
	case float64:
		return int(v), nil
	case string:
		n, err := strconv.Atoi(v)
		if err != nil {
			return 0, fmt.Errorf("invalid board size: %s", v)
		}
		return n, nil
	case int:
		return v, nil
	default:
		return 0, fmt.Errorf("unsupported board size type: %T", size)
	}
}

// analyzeResult 是解析 LLM 返回后的内部结果。
type analyzeResult struct {
	Action string
	Vertex *dto.GoVertex
}

// parseAIResponse 将 LLM 返回文本解析为动作和坐标。
func parseAIResponse(content string) (analyzeResult, error) {
	var raw struct {
		Action string `json:"action"`
		Vertex []int  `json:"vertex,omitempty"`
	}
	if err := json.Unmarshal([]byte(content), &raw); err != nil {
		return analyzeResult{}, fmt.Errorf("parse AI response: %w", err)
	}
	switch raw.Action {
	case "move":
		if len(raw.Vertex) != 2 {
			return analyzeResult{}, fmt.Errorf("vertex must have exactly 2 elements")
		}
		v := dto.GoVertex{raw.Vertex[0], raw.Vertex[1]}
		return analyzeResult{Action: "move", Vertex: &v}, nil
	case "end_game":
		return analyzeResult{Action: "end_game"}, nil
	default:
		return analyzeResult{}, fmt.Errorf("unknown action: %s", raw.Action)
	}
}
