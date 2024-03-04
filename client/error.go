package client

import (
	"fmt"
)

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (err Error) Error() string {
	return fmt.Sprintf("code:%d,Message:%s", err.Code, err.Message)
}
