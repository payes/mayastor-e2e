package main

import (
	v1 "k8s.io/api/core/v1"
	"log"
	"log_monitor/utils"
	"time"
)

func initFluentdPodMap() map[string]v1.Pod {
	pods, err := ListPods(utils.FluentdNS)
	if err != nil {
		log.Fatalln(err, "cannot list pods inside namespace")
	}
	podMap := make(map[string]v1.Pod)
	for _, p := range pods {
		c := 0
		for c < 4 {
			if PodStatus(&p) == "Running" {
				podMap[p.Name] = p
				break
			} else if PodStatus(&p) == "Pending" || PodStatus(&p) == "ContainerCreating" {
				c++
				time.Sleep(5 * time.Second)
			} else {
				break
			}
		}
	}
	return podMap
}

func checkForNewFluentdPods() {
	go func() {
		for {
			podNew := initFluentdPodMap()
			for k, _ := range podNew {
				if _, ok := app.PodMap[k]; !ok {
					app.PodMap[k] = podNew[k]
					execTailPodCommand(podNew[k])
				}
			}
			time.Sleep(5 * time.Minute)
		}
	}()
}
