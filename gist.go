package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/ssh/terminal"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

func login() {
	var username, pstring string
	fmt.Println("Logging in to Github...")
	fmt.Print("Username: ")
	fmt.Scan(&username)
	fmt.Print("Password: ")
	password, _ := terminal.ReadPassword(0)
	fmt.Print("\n")
	pstring = string(password)
	fmt.Println(username, pstring)
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

	// Read info
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Description: ")
	description, _ := reader.ReadString('\n')
	fmt.Print("Public (Y/n): ")
	fmt.Scan(&public)
	if public == "N" || public == "n" {
		pbool = false
	}

	// Create file map
	filemap := make(map[string]GistFile)

	for i := range files {
		filename := files[i]
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
	if len(os.Args) == 1 {
		fmt.Println("Usage: gist file1 file2 .. fileN")
	} else {
		files := os.Args[1:]
		upload(files)
	}
}
