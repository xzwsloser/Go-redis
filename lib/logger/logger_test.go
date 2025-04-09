package logger

import "testing"

func TestLogger(t *testing.T) {
	Debug("debug")
	Info("info: %d", 111)
	Warn("warn: %s", "hello")
	Error("error: %s", "error")
	Error("error: %s", "error")
	Error("error: %s", "error")
	Error("error: %s", "error")
	Error("error: %s", "error")
	Error("error: %s", "error")
	Error("error: %s", "error")
}
