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

type Authorization struct {
	Scopes        []string `json:"scopes"`
	Client_id     string   `json:"client_id"`
	Client_secret string   `json:"client_secret"`
	Note          string   `json:"note"`
	Note_url      string   `json:"note_url"`
}

func contains(haystack []string, needle string) bool {
	res, _ := regexp.MatchString("X-GitHub-OTP", haystack[0])
	return res
}

func makeRequest(body *bytes.Buffer, headers map[string]string) {
	fmt.Print(body, headers)
}

func login() {
	var twofa string
	var pstring, username string
	fmt.Println("Logging in to Github...")

	fmt.Print("Username: ")
	fmt.Scan(&username)

	fmt.Print("Password: ")
	password, _ := terminal.ReadPassword(0)
	fmt.Print("\n")
	pstring = string(password)

	fmt.Print("Two factor auth (optional): ")
	fmt.Scanln(&twofa)

	scopes := []string{"gist"}

	body := Authorization{
		Scopes:        scopes,
		Note:          "Gist.go",
		Note_url:      "https://github.com/AlfredDobradi/gist.go",
		Client_id:     "",
		Client_secret: "",
	}
	bodyJson, _ := json.Marshal(body)
	bodyBuf := bytes.NewBuffer(bodyJson)
	req, _ := http.NewRequest("POST", "https://api.github.com/authorizations", bodyBuf)
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

	res, _ := ioutil.ReadAll(resp.Body)
	var bodyObject interface{}
	json.Unmarshal(res, &bodyObject)
	m := bodyObject.(map[string]interface{})

	if resp.Status == "201 Created" {
		token := m["token"].(string)
		err := ioutil.WriteFile("auth.dat", []byte(token), 0644)
		if err != nil {
			panic(err)
		}
	} else if resp.Status == "403 Forbidden" {
		fmt.Println("response Status:", resp.Status)
		fmt.Println(m)
	}
}

type Gist struct {
	Description string              `json:"description"`
	Public      bool                `json:"public"`
	Files       map[string]GistFile `json:"files"`
}

type GistFile struct {
	Content string `json:"content"`
}

func upload(files []string) {
	var public string = "y"
	var pbool bool = true

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
	description, _ := reader.ReadString('\n')
	fmt.Print("Public (Y/n): ")
	fmt.Scanln(&public)
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
		content, _ := ioutil.ReadFile(files[i])
		trimmed := strings.Trim(string(content), "\n")
		filemap[filename] = GistFile{Content: trimmed}
	}

	gist := Gist{Description: description, Public: pbool, Files: filemap}
	gistJson, err := json.Marshal(gist)
	if err != nil {
		fmt.Println(err)
	} else {
		// POST /gists
		url := "https://api.github.com/gists"
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(gistJson))
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

		body, _ := ioutil.ReadAll(resp.Body)
		var bodyObject interface{}
		json.Unmarshal(body, &bodyObject)
		m := bodyObject.(map[string]interface{})

		if resp.Status == "200" || resp.Status == "201 Created" {
			fmt.Println(m["html_url"])
		} else {
			fmt.Println("response Status:", resp.Status)
			fmt.Println(m)
		}
	}
}

func main() {
	isLogin := flag.Bool("login", false, "Login mode")

	if len(os.Args) == 1 {
		fmt.Println("Usage: gist file1 file2 .. fileN")
	} else {
		flag.Parse()

		if *isLogin == true {
			login()
		} else {
			files := os.Args[1:]
			upload(files)
		}
	}
}
