package connectors

import (
	"encoding/json"
	"errors"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	apiToken = "ZTJlK3Rlc3QtZGlyZWN0b3JAbWF5YWRhdGEuaW86RXRFS0UyQUZSU21DWXFwd0dkbUkzRkZG"
	issueURL = "https://mayadata.atlassian.net/rest/api/3/issue/"
)

type JiraTask struct {
	Key    string  `json:"key"`
	Id     *string `json:"id"`
	Fields Fields  `json:"fields"`
}

type Fields struct {
	Assignee  Assignee   `json:"assignee"`
	Status    Status     `json:"status"`
	Name      *string    `json:"summary"`
	IssueType *IssueType `json:"issuetype"`
}

type Assignee struct {
	Email string `json:"emailAddress"`
	Name  string `json:"displayName,omitempty"`
}

type Status struct {
	Description string `json:"description,omitempty"`
	Name        string `json:"name"`
}

type IssueType struct {
	Name string `json:"name"`
}

func GetJiraTaskDetails(key string) (*JiraTask, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest(http.MethodGet, issueURL+key, nil)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Basic "+apiToken)
	resp, err := client.Do(req)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, errors.New(string(bodyBytes))
	}

	var jt JiraTask
	err = json.Unmarshal(bodyBytes, &jt)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	return &jt, nil
}
