package interpreter

import "log/slog"

type Interpreter struct {
	log *slog.Logger
}

func New() *Interpreter {
	return &Interpreter{}
}
