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
	"fmt"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

// checkCmd represents the check command
var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Test all your steps during a deployment",
	Long:  `This will play all checks described inside deployment steps and will print the result`,
	Run: func(cmd *cobra.Command, args []string) {
		file, _ := cmd.Flags().GetString("file")
		pipeline := readDescriptor(file)
		defer display()
		fmt.Printf("%s",pipeline)
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
	checkCmd.Flags().StringP("file", "f", viper.GetString("G2P_STATE_FILE"), "File describing all states that will be checked")
	_ = cobra.MarkFlagRequired(checkCmd.Flags(), "file")
}

func readDescriptor(filename string) Pipeline {
	var config Pipeline
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

func display() {
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	p := widgets.NewParagraph()
	p.Text = "Hello World!"
	p.SetRect(0, 0, 25, 5)

	ui.Render(p)

	for e := range ui.PollEvents() {
		if e.Type == ui.KeyboardEvent {
			break
		}
	}
}

type Pipeline struct {
	Name string
	Desc string
	Version string
	States []State
}

type State struct {
	Name string
	Desc string
	Components []Container
	Checks []Container
}

type Container struct {
	Name          string
	Image         string
	Environment   map[string]string
}