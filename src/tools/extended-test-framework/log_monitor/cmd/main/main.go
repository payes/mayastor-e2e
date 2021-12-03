package main

import (
	"bufio"
	"io"
	"log"
	"log_monitor/config"
)

var app config.AppConfig

func main() {
	client, err := Init()
	if err != nil {
		log.Fatalln(err)
	}
	logChan := make(chan string)
	defer close(logChan)
	ProcessLogLine(logChan)
	app.Client = client
	app.LogChannel = logChan
	app.PodMap = initPodMap()
	app.PipeReader, app.PipeWriter = io.Pipe()

	checkForNewPods()

	if len(app.PodMap) == 0 {
		log.Fatalln("There are no pods")
	}

	for _, v := range app.PodMap {
		execTailPodCommand(v)
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
