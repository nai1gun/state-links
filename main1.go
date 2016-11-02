package main

import (
	"io/ioutil"
	"encoding/json"
	"fmt"
	"bufio"
	"strings"
	"os"
	"sort"
)

type State struct {
	Name string
	Url string
	TemplateUrl string
	Controller string
	ControllerAs string
	Views Views
}

type Views map[string]View

type View struct {
	TenplateUrl string
	Controller string
	ControllerAs string
}

func readAllStates() []State {
	b, _ := ioutil.ReadFile("all_states.json")
	var states []State
	if err := json.Unmarshal(b, &states); err != nil {
		panic(err)
	}
	return states
}

func fixStatesController(states []State) []State {
	for i, state := range states {
		splitted := strings.Split(state.Controller, " as ")
		if (len(splitted) == 2) {
			state.Controller = splitted[0]
			state.ControllerAs = splitted[1]
			states[i] = state
		}
	}
	return states
}

func readRootStateNamesToKeep() map[string]string {
	file, _ := os.Open("root_states.csv")
	scanner := bufio.NewScanner(file)
	var states = map[string]string {}
	for scanner.Scan() {
		line := scanner.Text()
		rows := strings.Split(line, ";")
		isRoot := rows[3] == "N"
		if (isRoot) {
			states[rows[2]] = rows[2]
		}
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}
	return states
}

func checkAndAppendChildStates(allStates []State, rootStateNames map[string]string) map[string]string {
	for _, state := range allStates {
		_, contains := rootStateNames[state.Name]
		if (state.Name != "" && !contains) {
			parentState := strings.Split(state.Name, ".")[0]
			_, contains2 := rootStateNames[parentState]
			if (contains2) {
				rootStateNames[state.Name] = state.Name
			}
		}
	}
	return rootStateNames
}

func getRemainingStateNames(allStates []State, stateNames map[string]string) []string {
	var otherStateNames = []string {}
	for  _, state := range allStates {
		_, contains := stateNames[state.Name]
		if (!contains) {
			otherStateNames = append(otherStateNames, state.Name)
		}
	}
	return otherStateNames
}

func main() {
	var states = readAllStates()
	fixStatesController(states)
	var stateNamesToKeep = readRootStateNamesToKeep()
	checkAndAppendChildStates(states, stateNamesToKeep)
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