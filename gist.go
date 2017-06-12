package main

import (
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/ssh/terminal"
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
	description string              `json:"description"`
	public      bool                `json:"public"`
	files       map[string]GistFile `json:"files"`
}

type GistFile struct {
	content string `json:"content"`
}

func upload() {
	filemap := make(map[string]GistFile)
	filemap["file.txt"] = GistFile{content: "This is an example content"}

	gist := Gist{description: "example", public: false, files: filemap}
	fmt.Print(gist)
	jsongist, err := json.Marshal(gist)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(string(jsongist))
	}
}

func main() {
	upload()
}
