package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/alfreddobradi/snip.go/internal/auth"
	"github.com/alfreddobradi/snip.go/internal/gist"
)

func main() {
	isLogin := flag.Bool("login", false, "Login mode")

	if len(os.Args) == 1 {
		fmt.Println("Usage: gist file1 file2 .. fileN")
	} else {
		flag.Parse()

		if *isLogin {
			auth.Login()
		} else {
			files := os.Args[1:]
			gist.Upload(files)
		}
	}
}
