package utils

import "fmt"

func Error(msg any) {
	fmt.Printf("\033[91m [ERROR] %s \033[0m\n", msg)
}

func Info(msg any) {
	fmt.Printf("\033[94m [INFO] %s \033[0m\n", msg)
}

func Warn(msg any) {
	fmt.Printf("\033[93m [WARN] %s \033[0m\n", msg)
}

func Debug(msg any) {
	fmt.Printf("\033[36m [DEBUG] %s \033[0m\n", msg)
}

func ErrorF(msg string, args ...any) {
	fmt.Printf("\033[91m [ERROR] %s \033[0m\n", fmt.Sprintf(msg, args...))
}

func InfoF(msg string, args ...any) {
	fmt.Printf("\033[94m [INFO] %s \033[0m\n", fmt.Sprintf(msg, args...))
}

func WarnF(msg string, args ...any) {
	fmt.Printf("\033[93m [WARN] %s \033[0m\n", fmt.Sprintf(msg, args...))
}

func DebugF(msg string, args ...any) {
	fmt.Printf("\033[36m [DEBUG] %s \033[0m\n", fmt.Sprintf(msg, args...))
}
