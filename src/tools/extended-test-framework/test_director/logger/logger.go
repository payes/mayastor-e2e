package logger

import (
	log "github.com/sirupsen/logrus"
	"os"
	"test-director/config"
	"time"
)

// Init logger
func InitLogger(cfg *config.Logger) {

	log.SetReportCaller(cfg.ReportCaller)
	switch cfg.Level {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warning":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}

	if cfg.Encoding == "json" {
		log.SetFormatter(&log.JSONFormatter{
			TimestampFormat: time.RFC3339Nano,
		})
	} else {
		log.SetFormatter(&log.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: time.RFC3339Nano,
		})
	}

	if cfg.Output == "file" {
		// You could set this to any `io.Writer` such as a file
		file, err := os.OpenFile("test_director.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			log.SetOutput(file)
		} else {
			log.Error("Failed to log to file, using default stderr")
		}
	} else {
		log.SetOutput(os.Stdout)
	}
}
