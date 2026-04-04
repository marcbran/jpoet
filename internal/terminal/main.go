package terminal

import (
	"fmt"
	"os"
)

var noColor bool

func SetNoColor(v bool) {
	noColor = v
}

func paint(seq, text string) string {
	if noColor {
		return text
	}
	return seq + text + "\033[0m"
}

func Space() {
	fmt.Fprintln(os.Stderr, "")
}

func Info(text string) {
	_, _ = fmt.Fprintln(os.Stderr, paint("\033[34m", text))
}

func Infof(text string, a ...any) {
	Info(fmt.Sprintf(text, a...))
}

func Warn(text string) {
	_, _ = fmt.Fprintln(os.Stderr, paint("\033[33m", text))
}

func Warnf(text string, a ...any) {
	Warn(fmt.Sprintf(text, a...))
}

func Fail(text string) {
	_, _ = fmt.Fprintln(os.Stderr, paint("\033[31m", text))
}

func Failf(text string, a ...any) {
	Fail(fmt.Sprintf(text, a...))
}

func Success(text string) {
	_, _ = fmt.Fprintln(os.Stderr, paint("\033[32m", text))
}

func Successf(text string, a ...any) {
	Success(fmt.Sprintf(text, a...))
}
