package config

import (
	"io"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

type AppConfig struct {
	LogChannel chan string
	Client     kubernetes.Interface
	PodMap     map[string]v1.Pod
	PipeReader *io.PipeReader
	PipeWriter *io.PipeWriter
}
