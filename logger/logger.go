package logger

import (
	"io"
	"log"
	"os"
	"path/filepath"
)

var (
	Logger *log.Logger
	LogFile *os.File
)

func Init(logFilePath string) error {
	dir := filepath.Dir(logFilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	LogFile = file

	multiWriter := io.MultiWriter(os.Stdout, file)
	Logger = log.New(multiWriter, "", log.Ldate|log.Ltime|log.Lshortfile)
	
	Logger.Println("Logger initialized successfully")
	return nil
}

func Close() {
	if LogFile != nil {
		LogFile.Close()
	}
}