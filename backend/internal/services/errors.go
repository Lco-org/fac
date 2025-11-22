package services

import "errors"

var (
	ErrContainerNotFound   = errors.New("container not found")
	ErrProjectNotFound     = errors.New("project not found")
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrInvalidAmount       = errors.New("invalid amount")
)
