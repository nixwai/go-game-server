package model

import "errors"

var (
	// ErrNotFound 表示查询不到指定记录。
	ErrNotFound = errors.New("record not found")
	// ErrDuplicate 表示记录违反唯一约束。
	ErrDuplicate = errors.New("record already exists")
)
