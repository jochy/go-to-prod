// Copyright © 2020 DUBOIS ALEXANDRE ad.alexandre.dubois@gmail.com
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	g2p "go-to-prod/internal"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os/exec"
	"strings"
	"time"
)

var pipeline g2p.Pipeline
var stateTable *widgets.Table
var grid *ui.Grid
var docker *client.Client

// checkCmd represents the check command
var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Test all your steps during a deployment",
	Long:  `This will play all checks described inside deployment steps and will print the result`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := ui.Init(); err != nil {
			log.Fatalf("failed to initialize termui: %v", err)
		}
		cli, err := client.NewEnvClient()
		if err != nil {
			fmt.Println("Unable to create docker client")
			panic(err)
		}
		docker = cli
		docker.NegotiateAPIVersion(context.Background())

		defer ui.Close()

		file, _ := cmd.Flags().GetString("file")
		pipeline = readDescriptor(file)

		go processPipeline()

		display()
		uiEvents := ui.PollEvents()
		ticker := time.NewTicker(time.Second).C
		for {
			select {
			case e := <-uiEvents:
				switch e.ID {
				case "q", "<C-c>":
					return
				}
			case <-ticker:
				display()
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
	checkCmd.Flags().StringP("file", "f", viper.GetString("G2P_STATE_FILE"), "File describing all states that will be checked")
	_ = cobra.MarkFlagRequired(checkCmd.Flags(), "file")
}

func readDescriptor(filename string) g2p.Pipeline {
	var config g2p.Pipeline
	source, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(source, &config)
	if err != nil {
		panic(err)
	}
	return config
}

func processPipeline() {
	for index, _ := range pipeline.States {
		processState(&pipeline.States[index])
	}
}

func processState(state *g2p.State) {
	uid, _ := uuid.NewRandom()
	id := "g2p_" + strings.ReplaceAll(uid.String(), "-", "")
	state.Start()

	state.Operation = "Deploying"
	cmd := exec.Command("docker-compose", "-f", state.ComposeFile, "-p", id, "up", "-d")
	err := cmd.Run()
	if err != nil {
		_ = stopState(state, id)
		panic(err)
	}

	state.Operation = "Running tests"
	network := findNetwork(err, id)
	// Fixme : try to do better thant this
	time.Sleep(10 * time.Second)

	for checkerIndex, _ := range state.Checks {
		checker := &state.Checks[checkerIndex]
		checker.Start()
		runChecker(checker, network, id, state)
		checker.Stop()
	}

	state.Operation = "Undeploying"
	if stopState(state, id) != nil {
		panic("Unable to stop compose")
	}

	state.Stop()
}

func runChecker(checker *g2p.Checker, network types.NetworkResource, id string, state *g2p.State) {
	cont, err := docker.ContainerCreate(context.Background(),
		&container.Config{
			Env:   checker.Env,
			Image: checker.Image,
		},
		&container.HostConfig{
			NetworkMode: container.NetworkMode(network.Name),
		},
		nil,
		id+"_"+strings.ReplaceAll(checker.Name, " ", "_")+"_0")
	if err != nil {
		_ = stopState(state, id)
		panic(err)
	}
	err = docker.ContainerStart(context.Background(), cont.ID, types.ContainerStartOptions{})
	if err != nil {
		_ = stopState(state, id)
		panic(err)
	}
	statusCh, errCh := docker.ContainerWait(context.Background(), cont.ID, container.WaitConditionNextExit)
	select {
	case err := <-errCh:
		if err != nil {
			_ = stopState(state, id)
			panic(err)
		}
	case status := <-statusCh:
		checker.ExitCode = status.StatusCode
	}
	_ = docker.ContainerRemove(context.Background(), cont.ID, types.ContainerRemoveOptions{})
}

func findNetwork(err error, id string) types.NetworkResource {
	var network types.NetworkResource
	networks, err := docker.NetworkList(context.Background(), types.NetworkListOptions{})
	for _, net := range networks {
		if strings.Contains(net.Name, id) {
			network = net
			break
		}
	}
	return network
}

func stopState(state *g2p.State, id string) error {
	cmd := exec.Command("docker-compose", "-f", state.ComposeFile, "-p", id, "down")
	err := cmd.Run()
	return err
}

func display() {
	if stateTable == nil {
		stateTable = widgets.NewTable()
	}
	stateTable.Title = "Running pipeline state check for " + pipeline.Name
	stateTable.Rows = [][]string{}
	for index, _ := range pipeline.States {
		tmp := [][]string{printState(&pipeline.States[index])}
		for idx, _ := range pipeline.States[index].Checks {
			tmp = append(tmp, printCheckers(&pipeline.States[index].Checks[idx]))
		}
		stateTable.Rows = append(stateTable.Rows, tmp...)
	}

	if grid == nil {
		grid = ui.NewGrid()
	}
	termWidth, termHeight := ui.TerminalDimensions()
	grid.SetRect(0, 0, termWidth, termHeight)
	grid.Set(
		ui.NewRow(1.0,
			ui.NewCol(1.0, stateTable),
		),
	)

	ui.Render(grid)
}

func printState(state *g2p.State) []string {
	return []string{state.Name, state.ElapsedPrettyPrint(), state.Status(state.IsValid())}
}

func printCheckers(checker *g2p.Checker) []string {
	return []string{"     " + checker.Name, checker.ElapsedPrettyPrint(), checker.Status(checker.IsValid())}
}
