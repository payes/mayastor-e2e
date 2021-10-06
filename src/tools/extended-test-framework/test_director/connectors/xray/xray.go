package xray

import (
	"bytes"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	clientId     = "2471F500C6154736A3566E24F621A98E"
	clientSecret = "adbb5a7fa5d2c6a47db1c283f6366480d1321fc1a64ac00d5c2add14e4728700"
	authUrl      = "https://xray.cloud.xpand-it.com/api/v2/authenticate"
	graphqlUrl   = "https://xray.cloud.xpand-it.com/api/v2/graphql"
)

var token *string
var tries = 0

type Info struct {
	Summary          string    `json:"summary,omitempty"`
	Description      string    `json:"description,omitempty"`
	Version          string    `json:"version,omitempty"`
	User             string    `json:"user,omitempty"`
	Revision         string    `json:"revision,omitempty"`
	StartDate        time.Time `json:"startDate,omitempty"`
	FinishDate       time.Time `json:"finishDate,omitempty"`
	TestPlanKey      string    `json:"testPlanKey,omitempty"`
	TestEnvironments []string  `json:"testEnvironments,omitempty"`
}

type Test struct {
	TestKey      string    `json:"testKey,omitempty"`
	Start        time.Time `json:"start,omitempty"`
	Finish       time.Time `json:"finish,omitempty"`
	ActualResult string    `json:"actualResult,omitempty"`
	Status       string    `json:"status,omitempty"`
	Evidence     Evidence  `json:"evidence,omitempty"`
	Examples     []string  `json:"examples,omitempty"`
	Steps        []Step    `json:"steps,omitempty"`
	Defects      []string  `json:"defects,omitempty"`
}

type Evidence struct {
	Data        string `json:"data,omitempty"`
	Filename    string `json:"filename,omitempty"`
	ContentType string `json:"contentType,omitempty"`
}

type Step struct {
	Status       string     `json:"status,omitempty"`
	Comment      string     `json:"comment,omitempty"`
	ActualResult string     `json:"actualResult,omitempty"`
	Evidences    []Evidence `json:"evidences,omitempty"`
}

type Auth struct {
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

func authorize() *string {
	b, _ := json.Marshal(Auth{
		ClientId:     clientId,
		ClientSecret: clientSecret,
	})
	req, err := http.NewRequest(http.MethodPost, authUrl, bytes.NewBuffer(b))
	if err != nil {
		return nil
	}
	req.Header.Add("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error(err.Error())
		return nil
	}

	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err.Error())
		return nil
	}
	s := string(bodyBytes)
	if resp.StatusCode != 200 {
		log.Error(s)
		return nil
	}
	s = s[1 : len(s)-1]
	return &s
}

func sendQuery(s string) string {
	jsonData := map[string]string{
		"query": s,
	}
	jsonValue, _ := json.Marshal(jsonData)
	request, err := http.NewRequest(http.MethodPost, graphqlUrl, bytes.NewBuffer(jsonValue))
	if err != nil {
		log.Errorf("The HTTP request failed with error %s\n", err)
	}
	request.Header.Add("Authorization", "Bearer "+*authorize())
	request.Header.Add("Content-Type", "application/json")
	client := &http.Client{Timeout: time.Second * 10}
	response, err := client.Do(request)
	defer response.Body.Close()
	if err != nil {
		log.Errorf("The HTTP request failed with error %s\n", err)
	}

	data, _ := ioutil.ReadAll(response.Body)
	return string(data)
}
