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
const systemSeparator = `\`

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

func (states States) Get(stateName string) *State {// return pointer because struct object return cannot be nil
	for _, state := range states {
		if (state.Name == stateName) {
			return &state
		}
	}
	return nil
}

type Views map[string]View

type View struct {
	TemplateUrl string
	Controller string
	ControllerAs string
}

type FilesMap map[string]string

type Controllers map[string]string //map[controllerName, content]

type Templates map[string]string //map[templateUrl, content]

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
			relativePath := strings.TrimPrefix(absPath, rootFolder + systemSeparator)
			filesMap[relativePath] = content
		}
		return nil
	})
	return filesMap
}

func findReferencedStates(content string, reg *regexp.Regexp) StringSet {
	submatch := reg.FindAllStringSubmatch(content, -1)
	var states = StringSet {}
	for _, submatchItem := range submatch {
		states.Put(submatchItem[1])
	}
	return states
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
			var findInState func(State)
			findInState = func(st State) {
				var controllers = findStateControllers(st, filesMap)
				if (len(controllers) == 0) {
					fmt.Printf("Couldn't find any controllers for state '%s'\n", st.Name)
				} else {
					for _, content := range controllers {
						referencedStates := findReferencedStates(content,
							regexp.MustCompile(`\$state\.go\('(?P<State>.+)'`))
						for _, reference := range referencedStates {
							if (allStates.Contains(reference) && !stateNames.Contains(reference)) {
								stateNames.Put(reference)
								newState := allStates.Get(reference)
								findInState(*newState) //find recursively
							}
						}
					}
				}
			}
			findInState(state)
		}
	}
	return stateNames
}

func findStateTemplates(state State, filesMap FilesMap) Templates {
	var templates = Templates{}
	getContent := func(filesMap FilesMap, templateUrl string) {
		systemTemplateUrl := fixSeparator(templateUrl)
		templateContent := filesMap[systemTemplateUrl]
		templates[systemTemplateUrl] = templateContent
	}
	if (state.TemplateUrl != "") {
		getContent(filesMap, state.TemplateUrl)
	}
	for _, view := range state.Views {
		if (view.TemplateUrl != "") {
			getContent(filesMap, view.TemplateUrl)
		}
	}
	return templates
}

func fixSeparator(path string) string {
	return strings.Replace(path, "/", systemSeparator, -1)
}

func findUiSref(allStates States, stateNames StringSet, filesMap FilesMap) StringSet {
	return findInTemplates(allStates, stateNames, filesMap, regexp.MustCompile(`ui-sref="(?P<State>.+?)[\("]`))
}


func findStateRef(allStates States, stateNames StringSet, filesMap FilesMap) StringSet {
	return findInTemplates(allStates, stateNames, filesMap, regexp.MustCompile(`.state\('(?P<State>.+?)'`))
}

func findInTemplates(allStates States, stateNames StringSet, filesMap FilesMap, reg *regexp.Regexp) StringSet {
	for _, state := range allStates {
		if (stateNames.Contains(state.Name)) {
			findStateTemplates(state, filesMap)
			var findInState func(State)
			findInState = func(st State) {
				var templates = findStateTemplates(st, filesMap)
				if (len(templates) == 0) {
					fmt.Printf("Couldn't find any templates for state '%s'\n", st.Name)
				} else {
					for _, content := range templates {
						referencedStates := findReferencedStates(content, reg)
						for _, reference := range referencedStates {
							if (allStates.Contains(reference) && !stateNames.Contains(reference)) {
								stateNames.Put(reference)
								newState := allStates.Get(reference)
								findInState(*newState) //find recursively
							}
						}
					}
				}
			}
			findInState(state)
		}
	}
	return stateNames
}

func main() {
	var states = readAllStates()
	fixStatesController(states)
	var stateNamesToKeep = readRootStateNamesToKeep()
	checkAndAppendChildStates(states, stateNamesToKeep)

	filesMap := readAllFiles()
	findStateGo(states, stateNamesToKeep, filesMap)
	findUiSref(states, stateNamesToKeep, filesMap)
	findStateRef(states, stateNamesToKeep, filesMap)

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