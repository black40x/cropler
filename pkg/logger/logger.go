package logger

import (
	"fmt"
	"time"
)

const (
	LogInfoColor    = "\033[1;34m%s\033[0m"
	LogNoticeColor  = "\033[1;36m%s\033[0m"
	LogWarningColor = "\033[1;33m%s\033[0m"
	LogErrorColor   = "\033[1;31m%s\033[0m"
	LogDebugColor   = "\033[0;36m%s\033[0m"
	LogNone         = "%s"
)

func Log(str string) {
	fmt.Printf(LogNone, str)
}

func LogError(str string) {
	fmt.Printf(LogErrorColor, str)
}

func LogNotice(str string) {
	fmt.Printf(LogNoticeColor, str)
}

func LogDebug(str string) {
	fmt.Printf(LogDebugColor, str)
}

func LogWarning(str string) {
	fmt.Printf(LogWarningColor, str)
}

func LogInfo(str string) {
	fmt.Printf(LogInfoColor, str)
}

func CurrTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}
