package connectors

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	clientId     = "2471F500C6154736A3566E24F621A98E"
	clientSecret = "adbb5a7fa5d2c6a47db1c283f6366480d1321fc1a64ac00d5c2add14e4728700"
	authURL      = "https://xray.cloud.xpand-it.com/api/v1/authenticate"
)

type Auth struct {
	clientId     string
	clientSecret string
}

func authorize() error {
	b, _ := json.Marshal(Auth{
		clientId:     clientId,
		clientSecret: clientSecret,
	})
	req, err := http.NewRequest(http.MethodPost, authURL, bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Print(err.Error())
		return err
	}

	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Print(err.Error())
		return err
	}
	fmt.Printf("API Response as struct %+v\n", bodyBytes)
	return nil
}
