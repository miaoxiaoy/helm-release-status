package utils

import (
	"fmt"
)

type Logger struct {
	PrintHeader bool
}

func (l *Logger) Info(msg string) {
	if l.PrintHeader {
		fmt.Printf("\033[1;34;40m[INFO] %s\033[0m\n", msg)
	} else {
		fmt.Printf("\033[1;34;40m%s\033[0m\n", msg)
	}

}
func (l *Logger) Success(msg string) {
	if l.PrintHeader {
		fmt.Printf("\033[1;32;40m[INFO] %s\033[0m\n", msg)
	} else {
		fmt.Printf("\033[1;32;40m%s\033[0m\n", msg)
	}
}
func (l *Logger) Warning(msg string) {
	if l.PrintHeader {
		fmt.Printf("\033[1;33;40m[WARN] %s\033[0m\n", msg)
	} else {
		fmt.Printf("\033[1;33;40m%s\033[0m\n", msg)
	}
}

func (l *Logger) Error(msg string) {
	if l.PrintHeader {
		fmt.Printf("\033[1;31;40m[ERROR] %s\033[0m\n", msg)
	} else {
		fmt.Printf("\033[1;31;40m%s\033[0m\n", msg)
	}
}

func (l *Logger) Critical(msg string) {
	if l.PrintHeader {
		fmt.Printf("\033[1;37;41m[CRITICAL] %s\033[0m\n", msg)
	} else {
		fmt.Printf("\033[1;37;41m%s\033[0m\n", msg)
	}
}

var DefaultLogger = &Logger{PrintHeader: true}
var NoHeaderLogger = &Logger{PrintHeader: false}

