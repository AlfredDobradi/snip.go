package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"

	"golang.org/x/crypto/ssh/terminal"
)

// Authorization represents auth data sent to Github
type Authorization struct {
	Scopes       []string `json:"scopes"`
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	Note         string   `json:"note"`
	NoteURL      string   `json:"note_url"`
}

func contains(haystack []string, needle string) bool {
	res, err := regexp.MatchString("X-GitHub-OTP", haystack[0])
	panicOnError(err)
	return res
}

func createBodyBuffer(body interface{}) (*bytes.Buffer, error) {
	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	bodyBuf := bytes.NewBuffer(bodyJSON)

	return bodyBuf, nil
}

// func makeRequest(endpoint string, body *bytes.Buffer, headers map[string]string) {
	
// }

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}

func login() {
	var twofa string
	var pstring, username string
	fmt.Println("Logging in to Github...")

	fmt.Print("Username: ")
	_, unameErr := fmt.Scan(&username)
	panicOnError(unameErr)

	fmt.Print("Password: ")
	password, err := terminal.ReadPassword(0)
	panicOnError(err)
	fmt.Print("\n")
	pstring = string(password)

	fmt.Print("Two factor auth (optional): ")
	_, twoFaErr := fmt.Scanln(&twofa)
	panicOnError(twoFaErr)

	scopes := []string{"gist"}

	body := Authorization{
		Scopes:       scopes,
		Note:         "Gist.go",
		NoteURL:      "https://github.com/AlfredDobradi/gist.go",
		ClientID:     "",
		ClientSecret: "",
	}

	bodyBuf, err := createBodyBuffer(body)
	if err != nil {
		panic(err)
	}

	req, authReqErr := http.NewRequest("POST", "https://api.github.com/authorizations", bodyBuf)
	panicOnError(authReqErr)

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Content-Type", "application/json")
	if twofa != "" {
		req.Header.Set("X-GitHub-OTP", twofa)
	}
	req.SetBasicAuth(username, pstring)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	res, authResErr := ioutil.ReadAll(resp.Body)
	panicOnError(authResErr)
	var bodyObject map[string]interface{}
	authResUnmarshalErr := json.Unmarshal(res, &bodyObject)
	panicOnError(authResUnmarshalErr)

	if resp.Status == "201 Created" {
		token := bodyObject["token"].(string)
		err := ioutil.WriteFile("auth.dat", []byte(token), 0644)
		if err != nil {
			panic(err)
		}
	} else if resp.Status == "403 Forbidden" {
		fmt.Println("response Status:", resp.Status)
		fmt.Println(bodyObject)
	}
}

// Gist represents the gist structure sent to Github
type Gist struct {
	Description string              `json:"description"`
	Public      bool                `json:"public"`
	Files       map[string]GistFile `json:"files"`
}

// GistFile is a single file object
type GistFile struct {
	Content string `json:"content"`
}

func upload(files []string) {
	public := "y"
	pbool := true

	var token string

	if _, err := os.Stat("auth.dat"); err == nil {
		dat, err := ioutil.ReadFile("auth.dat")
		if err != nil {
			panic(err)
		}
		token = string(dat)
	}

	// Read info
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Description: ")
	description, descErr := reader.ReadString('\n')
	panicOnError(descErr)
	fmt.Print("Public (Y/n): ")
	_, publicErr := fmt.Scanln(&public)
	panicOnError(publicErr)
	if public == "N" || public == "n" {
		pbool = false
	}

	// Create file map
	filemap := make(map[string]GistFile)

	for i := range files {
		filename := files[i]
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			panic(err)
		}
		content, fileErr := ioutil.ReadFile(files[i])
		panicOnError(fileErr)
		trimmed := strings.Trim(string(content), "\n")
		filemap[filename] = GistFile{Content: trimmed}
	}

	gist := Gist{Description: description, Public: pbool, Files: filemap}
	gistJSON, bodyErr := createBodyBuffer(gist)
	panicOnError(bodyErr)

	// POST /gists
	url := "https://api.github.com/gists"
	req, sendReqErr := http.NewRequest("POST", url, gistJSON)
	panicOnError(sendReqErr)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Content-Type", "application/json")

	if token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("token %s", token))
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	body, sendResErr := ioutil.ReadAll(resp.Body)
	panicOnError(sendResErr)

	var bodyObject interface{}
	respBodyErr := json.Unmarshal(body, &bodyObject)
	panicOnError(respBodyErr)
	m := bodyObject.(map[string]interface{})

	if resp.Status == "200" || resp.Status == "201 Created" {
		fmt.Println(m["html_url"])
	} else {
		fmt.Println("response Status:", resp.Status)
		fmt.Println(m)
	}

}

func main() {
	isLogin := flag.Bool("login", false, "Login mode")

	if len(os.Args) == 1 {
		fmt.Println("Usage: gist file1 file2 .. fileN")
	} else {
		flag.Parse()

		if *isLogin {
			login()
		} else {
			files := os.Args[1:]
			upload(files)
		}
	}
}
