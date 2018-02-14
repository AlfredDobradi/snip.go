package gist

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/alfreddobradi/snip.go/internal/request"
)

// Gist represents the gist structure sent to Github
type Gist struct {
	Description string          `json:"description"`
	Public      bool            `json:"public"`
	Files       map[string]File `json:"files"`
}

// File is a single file object
type File struct {
	Content string `json:"content"`
}

// Upload sends the request to create Gist from file
func Upload(files []string) {
	public := "y"
	pbool := true

	var token string

	if _, err := os.Stat("auth.dat"); err == nil {
		dat, err := ioutil.ReadFile("auth.dat")
		if err != nil {
			log.Fatalf("Error while opening auth file: %v", err)
		}
		token = string(dat)
	}

	// Read info
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Description: ")
	description, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	fmt.Print("Public (Y/n): ")
	_, err = fmt.Scanln(&public)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	if public == "N" || public == "n" {
		pbool = false
	}

	// Create file map
	filemap := make(map[string]File)

	for i := range files {
		filename := files[i]
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			panic(err)
		}
		content, err := ioutil.ReadFile(files[i])
		if err != nil {
			log.Fatalf("Error: %v", err)
		}
		filemap[filename] = File{Content: strings.Trim(string(content), "\n")}
	}

	gist := Gist{Description: description, Public: pbool, Files: filemap}
	gistJSON, err := request.CreateBodyBuffer(gist)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	// POST /gists
	endpoint := "/gists"
	headers := make(map[string]string)

	headers["Accept"] = "application/vnd.github.v3+json"
	headers["Content-Type"] = "application/json"

	if token != "" {
		headers["Authorization"] = fmt.Sprintf("token %s", token)
	}

	respBody, status := request.Make(endpoint, gistJSON, headers, nil)
	if status == "200 OK" || status == "201 Created" {
		fmt.Println(respBody["html_url"])
	} else {
		fmt.Println("response Status:", status)
		fmt.Println(respBody)
	}
}
