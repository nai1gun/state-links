package main

import (
	"path/filepath"
	"os"
	"fmt"
	"io/ioutil"
	"strings"
)

var states = []string {}

func readStates() {
	b, _ := ioutil.ReadFile("ROOT_STATES")
	content := string(b)
	states = strings.Split(content, "\r\n")
}

func visit(path string, f os.FileInfo, _ error) error {
	if (!f.IsDir()) {
		b, _ := ioutil.ReadFile(path)
		content := string(b)
		for _, state := range states {
			if (strings.Contains(content, fmt.Sprintf("ui-sref=\"%s", state)) ||
				strings.Contains(content, fmt.Sprintf("$state.go('%s", state))) {
				fmt.Printf("Found %s in %s\n", state, path)
			}
		}
	}
	return nil
}


func main() {
	readStates()
	root := "..\\unity-client\\src\\main\\angular" // Root folder for the search
	err := filepath.Walk(root, visit)
	fmt.Printf("filepath.Walk() returned %v\n", err)
}