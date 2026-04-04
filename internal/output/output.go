package output

import (
	"fmt"
	"os"
)

const (
	red    = "\033[0;31m"
	green  = "\033[0;32m"
	yellow = "\033[1;33m"
	cyan   = "\033[0;36m"
	bold   = "\033[1m"
	reset  = "\033[0m"
)

var appName = "anvil"

func SetAppName(name string) { appName = name }
func AppName() string        { return appName }

func Info(msg string, args ...any) {
	fmt.Printf("%s[%s]%s %s\n", green, appName, reset, fmt.Sprintf(msg, args...))
}

func Warn(msg string, args ...any) {
	fmt.Printf("%s[%s]%s %s\n", yellow, appName, reset, fmt.Sprintf(msg, args...))
}

func Error(msg string, args ...any) {
	fmt.Fprintf(os.Stderr, "%s[%s]%s %s\n", red, appName, reset, fmt.Sprintf(msg, args...))
}

func Bold(s string) string   { return bold + s + reset }
func Green(s string) string  { return green + s + reset }
func Yellow(s string) string { return yellow + s + reset }
func Cyan(s string) string   { return cyan + s + reset }
func Red(s string) string    { return red + s + reset }
