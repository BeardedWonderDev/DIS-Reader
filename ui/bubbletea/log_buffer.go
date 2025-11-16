package bubbletea

import "strings"

type logBuffer struct {
	max   int
	lines []string
}

func newLogBuffer(max int) logBuffer {
	return logBuffer{max: max}
}

func (lb logBuffer) append(line string) logBuffer {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return lb
	}
	lines := append(lb.lines, trimmed)
	if lb.max > 0 && len(lines) > lb.max {
		lines = lines[len(lines)-lb.max:]
	}
	lb.lines = lines
	return lb
}

func (lb logBuffer) tail(n int) []string {
	if n <= 0 || n >= len(lb.lines) {
		return append([]string(nil), lb.lines...)
	}
	return append([]string(nil), lb.lines[len(lb.lines)-n:]...)
}
