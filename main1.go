package main

import (
	"io/ioutil"
	"encoding/json"
	"fmt"
	"bufio"
	"strings"
	"os"
	"sort"
	"path/filepath"
	"regexp"
)

const rootFolder = `..\unity-client\src\main\angular`

type StringSet map[string]string

func (s StringSet) Put(element string) {
	s[element] = element
}

func (s StringSet) Contains(element string) bool {
	_, contains := s[element]
	return contains
}

func (s StringSet) String() string {
	var str string = "["
	for item := range s {
		str += item + " "
	}
	str += "]"
	return str
}

type State struct {
	Name string
	Url string
	TemplateUrl string
	Controller string
	ControllerAs string
	Views Views
}

type States []State

func (states States) Contains(stateName string) bool {
	for _, state := range states {
		if (state.Name == stateName) {
			return true
		}
	}
	return false
}

type Views map[string]View

type View struct {
	TenplateUrl string
	Controller string
	ControllerAs string
}

type FilesMap map[string]string

type Controllers map[string]string //map[controllerName, content]

func readAllStates() States {
	b, _ := ioutil.ReadFile("all_states.json")
	var states States
	if err := json.Unmarshal(b, &states); err != nil {
		panic(err)
	}
	return states
}

func fixStatesController(states States) States {
	for i, state := range states {
		splitted := strings.Split(state.Controller, " as ")
		if (len(splitted) == 2) {
			state.Controller = splitted[0]
			state.ControllerAs = splitted[1]
			states[i] = state
		}
		for viewName, view := range state.Views {
			splitted2 := strings.Split(view.Controller, " as ")
			if (len(splitted2) == 2) {
				view.Controller = splitted2[0]
				view.ControllerAs = splitted2[1]
				state.Views[viewName] = view
			}
		}
	}
	return states
}

func readRootStateNamesToKeep() StringSet {
	file, _ := os.Open("root_states.csv")
	scanner := bufio.NewScanner(file)
	var stateNames = StringSet {}
	for scanner.Scan() {
		line := scanner.Text()
		rows := strings.Split(line, ";")
		isRoot := rows[3] == "N"
		if (isRoot) {
			stateNames.Put(rows[2])
		}
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}
	return stateNames
}

func checkAndAppendChildStates(allStates States, rootStateNames StringSet) StringSet {
	for _, state := range allStates {
		_, contains := rootStateNames[state.Name]
		if (state.Name != "" && !contains) {
			parentState := strings.Split(state.Name, ".")[0]
			_, contains2 := rootStateNames[parentState]
			if (contains2) {
				rootStateNames.Put(state.Name)
			}
		}
	}
	return rootStateNames
}

func getRemainingStateNames(allStates States, stateNames StringSet) []string {
	var otherStateNames = []string {}
	for  _, state := range allStates {
		if (!stateNames.Contains(state.Name)) {
			otherStateNames = append(otherStateNames, state.Name)
		}
	}
	return otherStateNames
}

func hasAnySuffix(str string, patterns... string) bool {
	for _, pattern := range patterns {
		if (strings.HasSuffix(str, pattern)) {
			return true
		}
	}
	return false
}

func hasNoSuffix(str string, patterns... string) bool {
	for _, pattern := range patterns {
		if (strings.HasSuffix(str, pattern)) {
			return false
		}
	}
	return true
}

func readAllFiles() FilesMap {
	var filesMap = FilesMap {}
	filepath.Walk(rootFolder, func(absPath string, f os.FileInfo, _ error) error {
		if (!f.IsDir() && hasAnySuffix(absPath, ".html", ".js") && hasNoSuffix(absPath, "_test.html")) {
			b, _ := ioutil.ReadFile(absPath)
			content := string(b)
			relativePath := strings.TrimPrefix(absPath, rootFolder + `\`)
			filesMap[relativePath] = content
		}
		return nil
	})
	return filesMap
}

func findControllerContent(filesMap FilesMap, controllerName string) string {
	for _, content := range filesMap {
		if (strings.Contains(content, ".controller('" + controllerName + "'")) {
			return content
		}
	}
	return ""
}

func findStateControllers(state State, filesMap FilesMap) Controllers {
	var controllers = Controllers{}
	getContent := func(filesMap FilesMap, controller string) {
		controllerContent := findControllerContent(filesMap, controller)
		if (controllerContent != "") {
			controllers[controller] = controllerContent
		} else {
			fmt.Printf("Couldn't find controller content for '%s'\n", controller)
		}
	}
	if (state.Controller != "") {
		getContent(filesMap, state.Controller)
	}
	for _, view := range state.Views {
		if (view.Controller != "") {
			getContent(filesMap, view.Controller)
		}
	}
	return controllers
}

func findStateGo(allStates States, stateNames StringSet, filesMap FilesMap) StringSet {
	for _, state := range allStates {
		if (stateNames.Contains(state.Name)) {
			var controllers = findStateControllers(state, filesMap)
			if (len(controllers) == 0) {
				fmt.Printf("Couldn't find any controllers for state '%s'\n", state.Name)
			} else {
				for _, content := range controllers {
					goReferenceStates := findStateGosForControllerContent(content)
					if (len(goReferenceStates) > 0) {
						for _, reference := range goReferenceStates {
							if (allStates.Contains(reference)) {
								stateNames.Put(reference)
							}
						}
					}
				}
			}
		}
	}
	return stateNames
}

func findStateGosForControllerContent (content string) StringSet {
	re := regexp.MustCompile(`\$state\.go\('(?P<State>.*)'`)
	submatch := re.FindAllStringSubmatch(content, -1)
	var states = StringSet {}
	for _, submatchItem := range submatch {
		states.Put(submatchItem[1])
	}
	return states
}

func main() {
	var states = readAllStates()
	fixStatesController(states)
	var stateNamesToKeep = readRootStateNamesToKeep()
	checkAndAppendChildStates(states, stateNamesToKeep)

	filesMap := readAllFiles()
	findStateGo(states, stateNamesToKeep, filesMap)

	var stateNamesList = []string {}
	for s := range stateNamesToKeep {
		stateNamesList = append(stateNamesList, s)
	}
	sort.Strings(stateNamesList)
	otherStateNames := getRemainingStateNames(states, stateNamesToKeep)
	sort.Strings(otherStateNames)
	fmt.Println("\nThese states should NOT be removed:")
	fmt.Println(stateNamesList)
	fmt.Println("\nRemaining states:")
	fmt.Println(otherStateNames)
}