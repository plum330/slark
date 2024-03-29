package colour

import "fmt"

func Red(str string) string {
	return fmt.Sprintf("\x1b[31m%s\x1b[0m", str)
}

func Yellow(str string) string {
	return fmt.Sprintf("\x1b[33m%s\x1b[0m", str)
}

func Blue(str string) string {
	return fmt.Sprintf("\x1b[34m%s\x1b[0m", str)
}

func Green(str string) string {
	return fmt.Sprintf("\x1b[32m%s\x1b[0m", str)
}
