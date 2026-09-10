package aigo

import (
	"github.com/nixwai/go-game-server/app/module/aigo/dto"
	"github.com/nixwai/go-game-server/app/response"
)

// validateSnapshot 校验棋局数据的结构完整性。
func validateSnapshot(req dto.AnalyzeRequest) error {
	size, err := parseSize(req.Size)
	if err != nil {
		return response.NewError(response.CodeValidation, "棋盘尺寸无效", err)
	}
	if size < 2 || size > 52 {
		return response.NewError(response.CodeValidation, "棋盘尺寸必须在 2-52 之间", nil)
	}
	if len(req.Layout) != size {
		return response.NewError(response.CodeValidation, "棋盘布局行数与尺寸不匹配", nil)
	}
	for _, row := range req.Layout {
		if len(row) != size {
			return response.NewError(response.CodeValidation, "棋盘布局列数与尺寸不匹配", nil)
		}
		for _, cell := range row {
			if cell != 0 && cell != 1 && cell != -1 {
				return response.NewError(response.CodeValidation, "棋盘包含无效棋子标记", nil)
			}
		}
	}
	if req.Player != 1 && req.Player != -1 {
		return response.NewError(response.CodeValidation, "执棋方必须为 1 或 -1", nil)
	}
	if req.Ko != nil {
		if req.Ko.Sign != 1 && req.Ko.Sign != -1 {
			return response.NewError(response.CodeValidation, "劫子标记必须为 1 或 -1", nil)
		}
	}
	return nil
}
