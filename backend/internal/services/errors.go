package services

import "errors"

var (
	ErrContainerNotFound      = errors.New("container not found")
	ErrProjectNotFound        = errors.New("project not found")
	ErrProjectCodeExists      = errors.New("project code already exists")
	ErrEmployeeNotFound       = errors.New("employee not found")
	ErrEmployeeNumberExists   = errors.New("employee number already exists")
	ErrMemberNotFound         = errors.New("member not found")
	ErrMembershipNumberExists = errors.New("membership number already exists")
	ErrInsufficientBalance    = errors.New("insufficient balance")
	ErrInvalidAmount          = errors.New("invalid amount")
)
