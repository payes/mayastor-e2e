package main

import (
	"bytes"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/util/json"
	"log"
	"log_monitor/models"
	"log_monitor/utils"
	"net/http"
	"strings"
)

func ProcessLogLine(logChan chan string) {
	go func() {
		for {
			msg := <-logChan
			checkLine(msg)
		}
	}()
}

func checkLine(s string) {
	if strings.Contains(strings.ToLower(s), "error") {
		sendEvent("FAIL", s)
	} else if strings.Contains(strings.ToLower(s), "warn") {
		sendEvent("WARN", s)
	}
}

func sendEvent(level, line string) {
	p, err := GetPod("test-director", "mayastor-e2e")
	if err != nil {
		log.Println("cannot find test-director pod", err)
		return
	}
	var trs []models.TestRun
	url := "http://" + p.Status.PodIP + utils.TestDirectorTestRunsAPI
	req, err := http.Get(url)
	if err != nil {
		log.Println(err.Error())
		return
	}
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Println(err.Error())
		return
	}
	err = json.Unmarshal(body, &trs)
	if err != nil {
		log.Println(err.Error())
		return
	}
	var testRun models.TestRun
	for _, tr := range trs {
		if tr.Spec.Status == "EXECUTING" {
			testRun = tr
			break
		}
	}
	var e models.Event
	url = "http://" + p.Status.PodIP + utils.TestDirectorEventsAPI
	e = models.Event{
		SourceClass:    "log-monitor",
		SourceInstance: testRun.ID,
		Resource:       "",
		Class:          level,
		Message:        line,
		Data:           nil,
	}
	b, err := json.Marshal(e)
	if err != nil {
		log.Println(err.Error())
		return
	}
	req, err = http.Post(url, "application/json", bytes.NewBuffer(b))
	if err != nil {
		log.Println(err.Error())
		return
	}
	defer req.Body.Close()
}
