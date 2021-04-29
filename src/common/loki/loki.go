package loki

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"mayastor-e2e/common/e2e_config"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var g_apiUser string
var g_apiPw string
var g_loki_run_id string
var g_enabled = false
var g_once sync.Once

func SendLokiMarker(text string) {
	g_once.Do(func() {
		g_apiUser = os.Getenv("grafana_api_user")
		g_apiPw = os.Getenv("grafana_api_pw")
		g_loki_run_id = os.Getenv("loki_run_id")

		if g_apiUser != "" && g_apiPw != "" && g_loki_run_id != "" {
			g_enabled = true
		} else if g_apiUser != "" || g_apiPw != "" || g_loki_run_id != "" { // all should be defined or none
			errorStr := "Invalid combination of environment variables"
			if g_apiUser == "" {
				errorStr += ", user is not defined"
			}
			if g_apiPw == "" {
				errorStr += ", password is not defined"
			}
			if g_loki_run_id == "" {
				errorStr += ", loki_run_id is not defined"
			}
			logf.Log.Info("Invalid Loki config", "reason", errorStr)
		}
	})

	if !g_enabled {
		return
	}

	imageTag := e2e_config.GetConfig().ImageTag
	timestamp := strconv.FormatInt(time.Now().UnixNano(), 10)

	logentryJSON := `
	{
		"streams": [
			{
				"stream": {
					"run": "` + g_loki_run_id + `",
					"version": "` + imageTag + `",
					"app": "marker"
				},
				"values": [
					["` + timestamp + `","` + text + `"]
				]
			}
		]
	}`
	compactedBuffer := new(bytes.Buffer)
	err := json.Compact(compactedBuffer, []byte(logentryJSON))
	if err != nil {
		logf.Log.Info("Failed to compact Loki request", "error", err)
		return
	}
	req, err := http.NewRequest("POST", "https://logs-prod-us-central1.grafana.net/loki/api/v1/push", compactedBuffer)
	if err != nil {
		logf.Log.Info("Failed to create Loki marker request", "error", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(g_apiUser, g_apiPw)

	client := &http.Client{}
	client.Timeout = time.Second * 10
	resp, err := client.Do(req)
	if err != nil {
		logf.Log.Info("Failed to send Loki marker", "error", err)
		return
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		logf.Log.Info("Unexpected response from Grafana / Loki", "status code", resp.StatusCode)
	}
	resp.Body.Close()
}
