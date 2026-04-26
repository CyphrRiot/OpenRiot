package logger

import "fmt"

// Console color codes
var (
	Cyan   = "\033[36m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Red    = "\033[31m"
	Reset  = "\033[0m"
)

// Simple colored logging helpers for console output
func Info(msg string) { fmt.Printf("%s[INFO]%s %s\n", Cyan, Reset, msg) }
func Warn(msg string) { fmt.Printf("%s[WARN]%s %s\n", Yellow, Reset, msg) }
func Done(msg string) { fmt.Printf("%s[DONE]%s %s\n", Green, Reset, msg) }
func Fail(msg string) { fmt.Printf("%s[FAIL]%s %s\n", Red, Reset, msg) }
func Ask(msg string)  { fmt.Printf("%s[ASK ]%s %s", Cyan, Reset, msg) }
