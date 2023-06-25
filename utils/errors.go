package utils

import "fmt"

type ClientLeftError struct {
	Username string
}

func (e *ClientLeftError) Error() string {
	return fmt.Sprintf("%s left the chat", e.Username)
}

type NotEnoughArgs struct {
	CmdName string
}

func (e *NotEnoughArgs) Error() string {
	return fmt.Sprintf("Command %s requires argument after '::'", e.CmdName)
}
