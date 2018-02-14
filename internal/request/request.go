package request

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

// CreateBodyBuffer creates a body buffer
func CreateBodyBuffer(body interface{}) (*bytes.Buffer, error) {
	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	bodyBuf := bytes.NewBuffer(bodyJSON)

	return bodyBuf, nil
}

func Make(endpoint string, body *bytes.Buffer, headers map[string]string, auth map[string]string) (map[string]interface{}, string) {
	url := "https://api.github.com" + endpoint
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	if auth != nil {
		req.SetBasicAuth(auth["username"], auth["password"])
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}

	var bodyObject map[string]interface{}

	err = json.Unmarshal(respBody, &bodyObject)
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}

	return bodyObject, resp.Status
}
