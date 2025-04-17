package database

import "errors"

var (
	ErrNoRows       = errors.New("ERR_NO_ROWS")
	ErrAlreadyExist = errors.New("ERR_ALREADY_EXISTS")
)
