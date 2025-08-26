package interpreter

func pointerTo[T any](v T) *T {
	return &v
}