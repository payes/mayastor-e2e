package main

import (
	"bufio"
	"io"
	v1 "k8s.io/api/core/v1"
	"log"
	"log_monitor/config"
	"os"
)

var app config.AppConfig

func main() {
	client, err := Init()
	if err != nil {
		log.Fatalln(err)
	}

	app.PodMap = make(map[string]v1.Pod)
	logChan := make(chan string)
	defer close(logChan)
	ProcessLogLine(logChan)

	app.Client = client
	app.LogChannel = logChan
	app.PipeReader, app.PipeWriter = io.Pipe()
	if os.Getenv("LOG_REGEX") != "" {
		app.LogRegex = os.Getenv("LOG_REGEX")
	} else {
		app.LogRegex = `level.{0,4}(error|warn)`
	}

	checkForNewFluentdPods()

	if len(app.PodMap) == 0 {
		log.Fatalln("There are no pods")
	}

	buf := bufio.NewReader(app.PipeReader)
	for {
		l, _, err := buf.ReadLine()
		app.LogChannel <- string(l)
		if err != nil {
			log.Fatalln(err, "read a line failed")
		}
	}
}
