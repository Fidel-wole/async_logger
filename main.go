package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

// Logger struct
type Logger struct {
	logChannel chan string
	done       chan struct{}
}

// NewLogger initializes a new async logger
func NewLogger(bufferSize int) *Logger {
	l := &Logger{
		logChannel: make(chan string, bufferSize),
		done:       make(chan struct{}),
	}

	// Start a goroutine to process log messages
	go l.processLogs()
	return l
}

// processLogs writes logs to file asynchronously
func (l *Logger) processLogs() {
	// Open a log file
	file, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("failed to open log file: %v", err)
	}
	defer file.Close()

	// Log to both file and terminal
	multiWriter := io.MultiWriter(file, os.Stdout)
	logger := log.New(multiWriter, "", log.LstdFlags)

	for {
		select {
		case msg := <-l.logChannel:
			logger.Println(msg) // Write log to both file & terminal
		case <-l.done:
			close(l.logChannel)
			for msg := range l.logChannel {
				logger.Println(msg) // Flush remaining logs
			}
			return
		}
	}
}

// Log sends a message to the logging channel
func (l *Logger) Log(message string) {
	select {
	case l.logChannel <- message:
	default:
		fmt.Println("Log buffer full, dropping log:", message) // Prevent blocking
	}
}

// Close gracefully shuts down the logger
func (l *Logger) Close() {
	close(l.done)
}

// Read and print log file contents
func printLogFile() {
	file, err := os.Open("app.log")
	if err != nil {
		fmt.Println("Error opening log file:", err)
		return
	}
	defer file.Close()

	fmt.Println("\n--- Log File Contents ---")
	io.Copy(os.Stdout, file) // Print file contents to terminal
	fmt.Println("\n-------------------------")
}

var logger = NewLogger(100)

func test(message string) {
	logger.Log(message)
}

func main() {
	defer logger.Close()
	test("Test log 1")
	test("Test log 2")
	test("Another test log")

	// Give some time for the goroutine to process logs before exit
	time.Sleep(1 * time.Second)

	// Read and print the log file contents
	printLogFile()

	fmt.Println("Done logging messages, exiting...")
}
