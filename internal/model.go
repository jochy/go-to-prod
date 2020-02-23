package internal

import (
	"strings"
	"time"
)

type Pipeline struct {
	Name    string
	Desc    string
	Version string
	States  []State
}

type State struct {
	Loggable

	Name        string
	Desc        string
	ComposeFile string `yaml:"compose-file"`
	Checks      []Checker
}

type Checker struct {
	Loggable

	Name  string
	Image string
	Env   []string

	ExitCode int64
}

type Loggable struct {
	start     *time.Time
	end       *time.Time
	tick      int
	Operation string
}

type Valid interface {
	IsValid() bool
}

func (loggable *Loggable) Start() {
	start := time.Now()
	loggable.start = &start
}

func (loggable *Loggable) Stop() {
	stop := time.Now()
	loggable.end = &stop
}

func (loggable *Loggable) IsDone() bool {
	return loggable.start != nil && loggable.end != nil
}

func (loggable *Loggable) IsStarted() bool {
	return loggable.start != nil
}

func (loggable *Loggable) ElapsedPrettyPrint() string {
	if loggable.end != nil {
		return loggable.end.Sub(*loggable.start).Round(time.Millisecond).String()
	} else if loggable.start != nil {
		return time.Now().Sub(*loggable.start).Round(time.Millisecond).String()
	}
	return ""
}

func (loggable *Loggable) Tick() int {
	loggable.tick++
	return loggable.tick
}

func (checker *Checker) IsValid() bool {
	return checker.ExitCode == 0
}

func (state *State) IsValid() bool {
	valid := true
	for _, checker := range state.Checks {
		valid = valid && checker.IsValid()
	}
	return valid
}

func (loggable *Loggable) Status(valid bool) string {
	status := "Pending"
	ope := "Running"

	if loggable.Operation != "" {
		ope = loggable.Operation
	}

	if loggable.IsDone() && valid {
		status = "Valid"
	} else if loggable.IsDone() && !valid {
		status = "Failed"
	} else if loggable.IsStarted() {
		status = ope + " " + strings.Repeat(".", loggable.Tick()%4)
	}
	return status
}
