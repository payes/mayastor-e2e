package connectors

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

//Authorization: Basic bWlsYW4uaGFqZWtAbWF5YWRhdGEuaW86YUZmZFlaWjN4b2dqRzBLM1pqYkE4NDY4 - milan.hajek@mayadata.io:aFfdYZZ3xogjG0K3ZjbA8468 - need to create some user for all
const (
	apiToken = "bWlsYW4uaGFqZWtAbWF5YWRhdGEuaW86YUZmZFlaWjN4b2dqRzBLM1pqYkE4NDY4"
	issueURL = "https://mayadata.atlassian.net/rest/api/3/issue/"
)

type JiraTask struct {
	Key    string `json:"key"`
	Fields Fields `json:"fields"`
}

type Fields struct {
	Assignee Assignee `json:"assignee"`
	Status   Status   `json:"status"`
	Name     *string  `json:"summary"`
}

type Assignee struct {
	Email string `json:"emailAddress"`
	Name  string `json:"displayName,omitempty"`
}

type Status struct {
	Description string `json:"description,omitempty"`
	Name        string `json:"name"`
}

func GetJiraTaskDetails(key string) (*JiraTask, error) {
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, issueURL+key, nil)
	if err != nil {
		fmt.Print(err.Error())
		return nil, err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Basic "+apiToken)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Print(err.Error())
		return nil, err
	}

	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Print(err.Error())
		return nil, err
	}

	var jt JiraTask
	err = json.Unmarshal(bodyBytes, &jt)
	if err != nil {
		fmt.Print(err.Error())
		return nil, err
	}
	return &jt, nil
}
