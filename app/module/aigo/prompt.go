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
		return `You are a strong but fast Go player. First assess the current board internally, then choose the single strongest and most effective legal move for the current player. Prioritize urgent tactics: capture an opponent group, save a group in danger, respond to an immediate threat, connect or cut groups, then choose the move with the best balance of territory and influence. Consider the last move and ko. Do not enumerate variations, show reasoning, or explain. Return only JSON: {"action":"move","vertex":[row,col]}; return {"action":"end_game"} only when continuing is no longer meaningful.`
	}
	return `You are a strong but fast Go player. First assess the current board internally, then choose the single strongest and most effective legal move for the current player. Prioritize urgent tactics: capture an opponent group, save a group in danger, respond to an immediate threat, connect or cut groups, then choose the move with the best balance of territory and influence. Consider the last move and ko. Do not enumerate variations, show reasoning, or explain. You must return only JSON: {"action":"move","vertex":[row,col]}.`
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
	b.WriteString("Coordinates are [row,col]. Board: X=black, O=white, .=empty. Assess this exact position and select the strongest effective legal move for the current player.\n")

	for r := 0; r < size; r++ {
		fmt.Fprintf(&b, "%d: ", r)
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
			b.WriteString(cell)
		}
		b.WriteString("\n")
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

// parseAIResponse 将 LLM 返回文本解析为动作和坐标，并校验坐标是否在棋盘范围内。
func parseAIResponse(content string, size int) (analyzeResult, error) {
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
		row, col := raw.Vertex[0], raw.Vertex[1]
		if row < 0 || row >= size || col < 0 || col >= size {
			return analyzeResult{}, fmt.Errorf("vertex out of bounds: [%d,%d] for size %d", row, col, size)
		}
		v := dto.GoVertex{row, col}
		return analyzeResult{Action: "move", Vertex: &v}, nil
	case "end_game":
		return analyzeResult{Action: "end_game"}, nil
	default:
		return analyzeResult{}, fmt.Errorf("unknown action: %s", raw.Action)
	}
}
