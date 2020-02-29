package internal

import (
	"fmt"
)

var firstDisplay = true
var pipeline Pipeline

func SetPipeline(pip Pipeline) {
	pipeline = pip
}

func LogMessage(msg string) {
	removeSummary()
	fmt.Printf("\033[0m%v\r\n", msg)
	displaySummary()
}

func removeSummary() {
	count := 0
	for _, state := range pipeline.States {
		count += len(state.Checks) + 1
	}
	for count >= 0 {
		fmt.Printf("\033[F\033[K")
		count--
	}
}

func displaySummary() {
	fmt.Printf("\033[0m=======================================================\r\n")
	for index, _ := range pipeline.States {
		renderData := printState(&pipeline.States[index])
		fmt.Printf("%v [%v] %v (%v)\r\n", renderData[3], renderData[2], renderData[0], renderData[1])

		for idx, _ := range pipeline.States[index].Checks {
			renderData = printCheckers(&pipeline.States[index].Checks[idx])
			fmt.Printf("%v   -- [%v] %v (%v)\r\n", renderData[3], renderData[2], renderData[0], renderData[1])
		}
	}
}

func UpdateSummary() {
	if !firstDisplay {
		removeSummary()
	}
	displaySummary()
	firstDisplay = false
}

func printState(state *State) []string {
	return []string{state.Name, state.ElapsedPrettyPrint(), state.Status(state.IsValid()), state.Color(state.IsValid())}
}

func printCheckers(checker *Checker) []string {
	var statusPrecision = ""
	if !checker.IsValid() {
		statusPrecision = " (exit code = " + fmt.Sprintf("%v", checker.ExitCode) + ")"
	}
	return []string{checker.Name, checker.ElapsedPrettyPrint(), checker.Status(checker.IsValid()) + statusPrecision, checker.Color(checker.IsValid())}
}