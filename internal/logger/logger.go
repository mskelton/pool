package logger

import (
	"fmt"

	"github.com/fatih/color"
)

var (
	infoPrefix    = color.BlueString("[INFO]")
	successPrefix = color.GreenString("[SUCCESS]")
	errorPrefix   = color.RedString("[ERROR]")
	warningPrefix = color.YellowString("[WARNING]")
)

func Info(format string, args ...interface{}) {
	fmt.Printf("%s %s\n", infoPrefix, fmt.Sprintf(format, args...))
}

func Success(format string, args ...interface{}) {
	fmt.Printf("%s %s\n", successPrefix, fmt.Sprintf(format, args...))
}

func Error(format string, args ...interface{}) {
	fmt.Fprintf(color.Error, "%s %s\n", errorPrefix, fmt.Sprintf(format, args...))
}

func Warning(format string, args ...interface{}) {
	fmt.Printf("%s %s\n", warningPrefix, fmt.Sprintf(format, args...))
}

func Plain(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}
