package auth

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/alfreddobradi/snip.go/internal/request"
	"golang.org/x/crypto/ssh/terminal"
)

// Authorization represents auth data sent to Github
type authorization struct {
	Scopes       []string `json:"scopes"`
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	Note         string   `json:"note"`
	NoteURL      string   `json:"note_url"`
}

func login() {
	var twofa string
	var pstring, username string
	fmt.Println("Logging in to Github...")

	fmt.Print("Username: ")
	_, err := fmt.Scan(&username)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Print("Password: ")
	password, err := terminal.ReadPassword(0)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	fmt.Print("\n")
	pstring = string(password)

	fmt.Print("Two factor auth (optional): ")
	_, err = fmt.Scanln(&twofa)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	scopes := []string{"gist"}

	body := authorization{
		Scopes:       scopes,
		Note:         "Gist.go",
		NoteURL:      "https://github.com/AlfredDobradi/gist.go",
		ClientID:     "",
		ClientSecret: "",
	}

	bodyBuf, err := request.CreateBodyBuffer(body)
	if err != nil {
		panic(err)
	}

	endpoint := "/authorizations"
	headers := make(map[string]string)

	headers["Accept"] = "application/vnd.github.v3+json"
	headers["Content-Type"] = "application/json"

	if twofa != "" {
		headers["X-GitHub-OTP"] = twofa
	}

	auth := make(map[string]string)
	auth["username"] = username
	auth["password"] = pstring

	respBody, status := request.Make(endpoint, bodyBuf, headers, auth)

	if status == "201 Created" {
		token := respBody["token"].(string)
		err := ioutil.WriteFile("auth.dat", []byte(token), 0644)
		if err != nil {
			panic(err)
		}
	} else {
		fmt.Println("response Status:", status)
		fmt.Println(respBody)
	}
}
