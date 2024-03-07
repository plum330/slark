package noop

import "io"

type writer struct{}

func (w *writer) Write([]byte) (int, error) {
	return 0, nil
}

func Writer() io.Writer {
	return &writer{}
}
