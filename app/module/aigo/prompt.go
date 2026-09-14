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
	prompt := `You are a strong but fast Go player. First assess the current board internally, using the exact board position, then choose the single strongest and most effective legal move for the current player. Consider the last move and ko. Coordinate correctness and move legality are more important than strategic preference.

Legal-move procedure:
1. Enumerate legal empty points from the exact board; never guess from a pattern or from the last move. If the user message provides rule-filtered move points, treat that list as authoritative and choose only from it.
2. A move is illegal if the target is occupied, outside the board, forbidden by ko, or leaves the newly placed stone's connected group with no liberty after removing any captured adjacent opponent groups. A capturing move is not suicide.
3. A point surrounded by enemy stones is not automatically playable: simulate the placement, remove every adjacent enemy group with zero liberties, and reject the move when your resulting group still has zero liberties. If no enemy group is captured, a zero-liberty placement is always suicide.
4. Check urgent tactics first: capture an opponent group, save a group in danger, answer an immediate threat, connect friendly stones, cut opposing stones, then choose the best territorial and influential move.
5. Before finalizing, re-check the exact cell with layout[y][x] and ensure it appears in the supplied rule-filtered whitelist. The first answer coordinate is x (column), the second is y (row); do not swap them.

Do not enumerate variations, show reasoning, or explain. Return only JSON, as one JSON object with no markdown or extra text: {"action":"move","vertex":[x,y]}.`
	if allowEndGame {
		prompt += ` You may return {"action":"end_game"} only when continuing the game is genuinely no longer meaningful; otherwise always return a move.`
	}
	return prompt
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
		fmt.Fprintf(&b, "Ko restriction: do not play at [%d,%d] on this move; ko stone sign=%d.\n", req.Ko.Vertex[0], req.Ko.Vertex[1], req.Ko.Sign)
	}
	if req.LatestVertex != nil {
		fmt.Fprintf(&b, "Last move: [%d,%d] (context only; it is not automatically a candidate).\n", req.LatestVertex[0], req.LatestVertex[1])
	}
	b.WriteString("Coordinate convention: every vertex is [x,y], never [row,col]. x is the column from left to right; y is the row from top to bottom; the stone at [x,y] is layout[y][x]. Example: [1,2] means layout[2][1], not layout[1][2]. Each row represents y, and characters from left to right represent x=0..size-1. latestVertex and ko.vertex use the same [x,y] convention. Board symbols: X=black, O=white, .=empty. Only an empty '.' point can be a move candidate.\n")
	b.WriteString("Empty points (candidate coordinates before suicide/ko checks): ")
	writeEmptyPoints(&b, req.Layout, size)
	b.WriteByte('\n')
	b.WriteString("Rule-filtered move points (non-suicide, excluding ko): ")
	writeRuleFilteredPoints(&b, req.Layout, size, req.Player, req.Ko)
	b.WriteByte('\n')
	b.WriteString("Immediate capture points (if any): ")
	writeCapturePoints(&b, req.Layout, size, req.Player, req.Ko)
	b.WriteByte('\n')
	b.WriteString("Forbidden empty points (suicide or ko; NEVER choose): ")
	writeForbiddenPoints(&b, req.Layout, size, req.Player, req.Ko)
	b.WriteByte('\n')
	b.WriteString("Coordinate ruler: x=")
	for x := 0; x < size; x++ {
		fmt.Fprintf(&b, "%d", x)
		if x+1 < size {
			b.WriteByte(' ')
		}
	}
	b.WriteString("\n")

	for y := 0; y < size; y++ {
		fmt.Fprintf(&b, "y=%d | ", y)
		for x := 0; x < size; x++ {
			cell := "."
			if y < len(req.Layout) && x < len(req.Layout[y]) {
				switch req.Layout[y][x] {
				case 1:
					cell = "X"
				case -1:
					cell = "O"
				}
			}
			b.WriteString(cell)
			if x+1 < size {
				b.WriteByte(' ')
			}
		}
		b.WriteByte('\n')
	}
	b.WriteString("Before finalizing: choose a point shown as '.', simulate the move including captures, confirm it is not suicide or ko-illegal, then verify that vertex [x,y] points to row y and column x in this diagram.\n")
	b.WriteString("FINAL MOVE CONSTRAINT — choose vertex exactly from this rule-filtered list and from nowhere else: ")
	writeRuleFilteredPoints(&b, req.Layout, size, req.Player, req.Ko)
	b.WriteString(". Never output an occupied, suicide, or ko-forbidden point.\n")
	b.WriteString("FINAL SAFETY CHECK — reject every coordinate listed as forbidden, even if it is empty: ")
	writeForbiddenPoints(&b, req.Layout, size, req.Player, req.Ko)
	b.WriteString(".\n")
	b.WriteString("OUTPUT WHITELIST (machine-readable): {\"allowed_vertices\":[")
	writeRuleFilteredPoints(&b, req.Layout, size, req.Player, req.Ko)
	b.WriteString("]}. The vertex array in your JSON response MUST exactly equal one array in allowed_vertices.")

	return b.String(), nil
}

// writeEmptyPoints 将棋盘中的空点按 [x,y] 格式写入提示词。
func writeEmptyPoints(b *strings.Builder, layout dto.GoLayout, size int) {
	first := true
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			if y >= len(layout) || x >= len(layout[y]) || layout[y][x] != 0 {
				continue
			}
			if !first {
				b.WriteString(", ")
			}
			fmt.Fprintf(b, "[%d,%d]", x, y)
			first = false
		}
	}
	if first {
		b.WriteString("none")
	}
}

// writeRuleFilteredPoints 将明显非法的自杀点和劫点从候选点中剔除。
func writeRuleFilteredPoints(b *strings.Builder, layout dto.GoLayout, size int, player dto.PlayerSign, ko *dto.KoInfo) {
	first := true
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			if !isPromptMoveLegal(layout, size, x, y, player, ko) {
				continue
			}
			writePromptPoint(b, &first, x, y)
		}
	}
	if first {
		b.WriteString("none")
	}
}

// writeCapturePoints 将当前落子可以立即提子的点写入提示词。
func writeCapturePoints(b *strings.Builder, layout dto.GoLayout, size int, player dto.PlayerSign, ko *dto.KoInfo) {
	first := true
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			if !isPromptMoveLegal(layout, size, x, y, player, ko) || promptCaptureCount(layout, size, x, y, player) == 0 {
				continue
			}
			writePromptPoint(b, &first, x, y)
		}
	}
	if first {
		b.WriteString("none")
	}
}

func writeForbiddenPoints(b *strings.Builder, layout dto.GoLayout, size int, player dto.PlayerSign, ko *dto.KoInfo) {
	first := true
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			if y >= len(layout) || x >= len(layout[y]) || layout[y][x] != 0 || isPromptMoveLegal(layout, size, x, y, player, ko) {
				continue
			}
			writePromptPoint(b, &first, x, y)
		}
	}
	if first {
		b.WriteString("none")
	}
}

func writePromptPoint(b *strings.Builder, first *bool, x, y int) {
	if !*first {
		b.WriteString(", ")
	}
	fmt.Fprintf(b, "[%d,%d]", x, y)
	*first = false
}

func isPromptMoveLegal(layout dto.GoLayout, size, x, y int, player dto.PlayerSign, ko *dto.KoInfo) bool {
	if x < 0 || x >= size || y < 0 || y >= size || y >= len(layout) || x >= len(layout[y]) || layout[y][x] != 0 {
		return false
	}
	if ko != nil && ko.Vertex[0] == x && ko.Vertex[1] == y {
		return false
	}
	board := clonePromptLayout(layout, size)
	board[y][x] = dto.GoSign(player)
	for _, point := range promptNeighbors(x, y, size) {
		if board[point[1]][point[0]] != -dto.GoSign(player) {
			continue
		}
		group, liberties := promptGroup(board, size, point[0], point[1])
		if len(liberties) == 0 {
			for _, stone := range group {
				board[stone[1]][stone[0]] = 0
			}
		}
	}
	_, liberties := promptGroup(board, size, x, y)
	return len(liberties) > 0
}

func promptCaptureCount(layout dto.GoLayout, size, x, y int, player dto.PlayerSign) int {
	if y >= len(layout) || x >= len(layout[y]) || layout[y][x] != 0 {
		return 0
	}
	board := clonePromptLayout(layout, size)
	board[y][x] = dto.GoSign(player)
	captured := 0
	for _, point := range promptNeighbors(x, y, size) {
		if board[point[1]][point[0]] != -dto.GoSign(player) {
			continue
		}
		group, liberties := promptGroup(board, size, point[0], point[1])
		if len(liberties) != 0 {
			continue
		}
		captured += len(group)
		for _, stone := range group {
			board[stone[1]][stone[0]] = 0
		}
	}
	return captured
}

func clonePromptLayout(layout dto.GoLayout, size int) dto.GoLayout {
	board := make(dto.GoLayout, size)
	for y := 0; y < size; y++ {
		board[y] = make([]dto.GoSign, size)
		if y < len(layout) {
			copy(board[y], layout[y])
		}
	}
	return board
}

func promptNeighbors(x, y, size int) [][2]int {
	neighbors := make([][2]int, 0, 4)
	if x > 0 {
		neighbors = append(neighbors, [2]int{x - 1, y})
	}
	if x+1 < size {
		neighbors = append(neighbors, [2]int{x + 1, y})
	}
	if y > 0 {
		neighbors = append(neighbors, [2]int{x, y - 1})
	}
	if y+1 < size {
		neighbors = append(neighbors, [2]int{x, y + 1})
	}
	return neighbors
}

func promptGroup(board dto.GoLayout, size, startX, startY int) ([][2]int, [][2]int) {
	sign := board[startY][startX]
	stones := make([][2]int, 0)
	liberties := make([][2]int, 0)
	seenStones := map[[2]int]bool{}
	seenLiberties := map[[2]int]bool{}
	queue := [][2]int{{startX, startY}}
	for len(queue) > 0 {
		point := queue[0]
		queue = queue[1:]
		if seenStones[point] {
			continue
		}
		seenStones[point] = true
		stones = append(stones, point)
		for _, neighbor := range promptNeighbors(point[0], point[1], size) {
			value := board[neighbor[1]][neighbor[0]]
			switch value {
			case 0:
				if !seenLiberties[neighbor] {
					seenLiberties[neighbor] = true
					liberties = append(liberties, neighbor)
				}
			case sign:
				if !seenStones[neighbor] {
					queue = append(queue, neighbor)
				}
			}
		}
	}
	return stones, liberties
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
		x, y := raw.Vertex[0], raw.Vertex[1]
		if x < 0 || x >= size || y < 0 || y >= size {
			return analyzeResult{}, fmt.Errorf("vertex out of bounds: [%d,%d] for size %d", x, y, size)
		}
		v := dto.GoVertex{x, y}
		return analyzeResult{Action: "move", Vertex: &v}, nil
	case "end_game":
		return analyzeResult{Action: "end_game"}, nil
	default:
		return analyzeResult{}, fmt.Errorf("unknown action: %s", raw.Action)
	}
}
