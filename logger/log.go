package logger

import (
	"fmt"
	"log"
	"os"
	"sync"
)

type Logger struct {
	logChannel chan LogMessage
	fileLogger *log.Logger
	console    *log.Logger
	logFile    *os.File
	done       chan struct{}
}

type LogMessage struct {
	Level   string
	Message string
}

var (
	instance *Logger
	once     sync.Once
)

// Init initializes the global logger instance.
func Init(logFilePath string) error {
	var initErr error
	once.Do(func() {
		instance = &Logger{
			logChannel: make(chan LogMessage, 100),
			done:       make(chan struct{}),
		}

		instance.logFile, initErr = os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if initErr != nil {
			return
		}

		instance.fileLogger = log.New(instance.logFile, "", log.LstdFlags|log.Lshortfile)
		instance.console = log.New(os.Stdout, "", log.LstdFlags)

		go instance.processLogs()
	})
	return initErr
}

// processLogs writes log messages from the channel to the console and file.
func (l *Logger) processLogs() {
	defer close(l.logChannel) // Ensure the channel is closed when the goroutine exits

	for {
		select {
		case logMsg := <-l.logChannel:
			formattedMessage := "[" + logMsg.Level + "] " + logMsg.Message
			l.console.Println(formattedMessage)
			l.fileLogger.Println(formattedMessage)
		case <-l.done:
			return
		}
	}
}

// Log accepts a format string and arguments, formatting them into a single log message.
func Log(level string, format string, args ...interface{}) {
	if instance != nil {
		message := fmt.Sprintf(format, args...) // Format the log message
		instance.logChannel <- LogMessage{Level: level, Message: message}
	}
}

// Info logs informational messages with formatting.
func Info(format string, args ...interface{}) {
	Log("INFO", format, args...)
}

// Warn logs warning messages with formatting.
func Warn(format string, args ...interface{}) {
	Log("WARN", format, args...)
}

// Error logs error messages with formatting.
func Error(format string, args ...interface{}) {
	Log("ERROR", format, args...)
}

func Fatal(format string, args ...interface{}) {
	Log("FATAL", format, args...)
	os.Exit(1)
}

// Close cleans up resources and ensures all logs are written before exiting.
func Close() {
	if instance != nil {
		// Signal the processLogs goroutine to stop
		close(instance.done)

		// Drain the logChannel and process remaining messages
		for logMsg := range instance.logChannel {
			formattedMessage := "[" + logMsg.Level + "] " + logMsg.Message
			instance.console.Println(formattedMessage)
			instance.fileLogger.Println(formattedMessage)
		}

		// Close the log file
		instance.logFile.Close()
		instance = nil
	}
}
