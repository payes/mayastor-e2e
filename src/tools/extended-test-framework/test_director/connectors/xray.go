package connectors

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	clientId = "2471F500C6154736A3566E24F621A98E"
	clientSecret = "adbb5a7fa5d2c6a47db1c283f6366480d1321fc1a64ac00d5c2add14e4728700"
	authURL = "https://xray.cloud.xpand-it.com/api/v2/authenticate"
	importURL = "https://xray.cloud.xpand-it.com/api/v2/import/execution"
)

type Info struct {
	Summary string `json:"summary,omitempty"`
	Description string `json:"description,omitempty"`
	Version string `json:"version,omitempty"`
	User string `json:"user,omitempty"`
	Revision string `json:"revision,omitempty"`
	StartDate time.Time `json:"startDate,omitempty"`
	FinishDate time.Time `json:"finishDate,omitempty"`
	TestPlanKey string `json:"testPlanKey,omitempty"`
	TestEnvironments []string `json:"testEnvironments,omitempty"`
}

type Test struct {
	TestKey      string    `json:"testKey,omitempty"`
	Start        time.Time `json:"start,omitempty"`
	Finish       time.Time `json:"finish,omitempty"`
	ActualResult string    `json:"actualResult,omitempty"`
	Status       string    `json:"status,omitempty"`
	Evidence     Evidence `json:"evidence,omitempty"`
	Examples []string `json:"examples,omitempty"`
	Steps    []Step `json:"steps,omitempty"`
	Defects []string `json:"defects,omitempty"`
}

type Evidence struct {
	Data        string `json:"data,omitempty"`
	Filename    string `json:"filename,omitempty"`
	ContentType string `json:"contentType,omitempty"`
}

type Step struct {
	Status       string `json:"status,omitempty"`
	Comment      string `json:"comment,omitempty"`
	ActualResult string `json:"actualResult,omitempty"`
	Evidences    []Evidence `json:"evidences,omitempty"`
}

type Auth struct {
	ClientId string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

func Authorize() *string {
	b, _ := json.Marshal(Auth{
		ClientId:     clientId,
		ClientSecret: clientSecret,
	})
	fmt.Printf("API Request as struct %s\n", string(b))
	req, err := http.NewRequest(http.MethodPost, authURL, bytes.NewBuffer(b))
	if err != nil {
		return nil
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Print(err.Error())
		return nil
	}

	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Print(err.Error())
		return nil
	}
	s := string(bodyBytes)
	if resp.StatusCode != 200 {
		fmt.Print(s)
		return nil
	}
	return &s
}

