package core

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// LogCommand logs a command to the log file
func LogCommand(logFile, cwd, containerName string, args []string) {
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	fmt.Fprintf(f, "[%s] PWD: %s | Container: %s | Command: %s\n",
		timestamp, cwd, containerName, strings.Join(args, " "))
}
