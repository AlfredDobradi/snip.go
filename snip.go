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

func makeRequest(endpoint string, body *bytes.Buffer, headers map[string]string, auth map[string]string) (map[string]interface{}, string) {
	url := "https://api.github.com" + endpoint
	req, sendReqErr := http.NewRequest("POST", url, body)
	panicOnError(sendReqErr)

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

	respBody, sendResErr := ioutil.ReadAll(resp.Body)
	panicOnError(sendResErr)

	var bodyObject map[string]interface{}

	respBodyErr := json.Unmarshal(respBody, &bodyObject)
	panicOnError(respBodyErr)

	return bodyObject, resp.Status
}

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
		NoteURL:      "https://github.com/AlfredDobradi/snip.go",
		ClientID:     "",
		ClientSecret: "",
	}

	bodyBuf, err := createBodyBuffer(body)
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

	respBody, status := makeRequest(endpoint, bodyBuf, headers, auth)

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
	endpoint := "/gists"
	headers := make(map[string]string)

	headers["Accept"] = "application/vnd.github.v3+json"
	headers["Content-Type"] = "application/json"

	if token != "" {
		headers["Authorization"] = fmt.Sprintf("token %s", token)
	}

	respBody, status := makeRequest(endpoint, gistJSON, headers, nil)
	if status == "200 OK" || status == "201 Created" {
		fmt.Println(respBody["html_url"])
	} else {
		fmt.Println("response Status:", status)
		fmt.Println(respBody)
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
