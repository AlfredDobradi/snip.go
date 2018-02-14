package snip

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

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
