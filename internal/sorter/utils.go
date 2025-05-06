package sorter

import (
	"log"
	"os"
)

// LogExecution logs the execution details to a specified log file.
func LogExecution(message string) {
	logFile, err := os.OpenFile("execution.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer logFile.Close()

	logger := log.New(logFile, "", log.LstdFlags)
	logger.Println(message)
}

// HandleError logs the error message and exits the application if necessary.
func HandleError(err error, exit bool) {
	if err != nil {
		LogExecution(err.Error())
		if exit {
			os.Exit(1)
		}
	}
}
