package main

import (
	"io/ioutil"
	"encoding/json"
	"fmt"
)

type State struct {
	Name string
	Url string
	TenplateUrl string
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

func main() {
	var states = readAllStates()
	fmt.Println(states)
}