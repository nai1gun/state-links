package main

import (
	"path/filepath"
	"os"
	"fmt"
	"io/ioutil"
	"strings"
	"regexp"
)

var states = []string {}
var links = map[string]string {}

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

func getUiSrefLinks(path string, f os.FileInfo, _ error) error {
	if (!f.IsDir()) {
		b, _ := ioutil.ReadFile(path)
		content := string(b)
		if (strings.HasSuffix(f.Name(), ".html") && strings.Contains(content, "ui-sref=")) {
			re := regexp.MustCompile("\"([^\"]+)\"")
			splitted := strings.Split(content, "ui-sref=")
			for _, part := range splitted {
				rm := re.FindStringSubmatch(part)
				if len(rm) != 0 {
					links[rm[1]] = rm[1] + "(ui-sref)"
				}
			}
		}
	}
	return nil
}

func getHtmlStateLinks(path string, f os.FileInfo, _ error) error {
	if (!f.IsDir()) {
		b, _ := ioutil.ReadFile(path)
		content := string(b)
		if (strings.HasSuffix(f.Name(), ".html") && strings.Contains(content, ".state(")) {
			re := regexp.MustCompile("\\('([^\"]+)'")
			splitted := strings.Split(content, ".state")
			for _, part := range splitted {
				rm := re.FindStringSubmatch(part)
				if len(rm) != 0 {
					links[rm[1]] = rm[1] + "(html-state)"
				}
			}
		}
	}
	return nil
}

func getStateGoLinks(path string, f os.FileInfo, _ error) error {
	if (!f.IsDir()) {
		b, _ := ioutil.ReadFile(path)
		content := string(b)
		if (strings.HasSuffix(f.Name(), ".js") && !strings.HasSuffix(f.Name(), "_test.js") && strings.Contains(content, "$state.go")) {
			re := regexp.MustCompile("\\(\\s?'([^']+)'")
			splitted := strings.Split(content, "$state.go")
			for _, part := range splitted {
				rm := re.FindStringSubmatch(part)
				if len(rm) != 0 {
					links[rm[1]] = rm[1] + "(state)"
				}
			}
		}
	}
	return nil
}

func main() {
	readStates()
	root := "..\\unity-client\\src\\main\\angular" // Root folder for the search
	filepath.Walk(root, getUiSrefLinks)
	filepath.Walk(root, getStateGoLinks)
	filepath.Walk(root, getHtmlStateLinks)
	for _, link := range links {
		fmt.Println(link)
	}
}