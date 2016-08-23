// Copyright © 2016 NAME HERE <EMAIL ADDRESS>
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
	"math"
	"os"
	"sync"

	ui "github.com/gizak/termui"
	"github.com/iancmcc/jig/manifest"
	"github.com/iancmcc/jig/vcs"
	"github.com/spf13/cobra"
)

// restoreCmd represents the restore command
var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "A brief description of your command",
	Long:  `A`,
	Run: func(cmd *cobra.Command, args []string) {
		man, err := manifest.FromJSON(os.Stdin)
		if err != nil {
			panic(err)
		}

		/*
			for _, repo := range man.Repos {
				// Ping it to check auth
				c := exec.Command("git", "ls-remote", repo.Repo)
				c.Stdin = nil
				rc := c.Run()
				fmt.Println(rc)

			}
		*/

		err = ui.Init()
		if err != nil {
			panic(err)
		}
		defer ui.Close()

		maxItemsPerColumn := float64(10)
		ncols := int(math.Ceil(float64(len(man.Repos)) / maxItemsPerColumn))

		span := 12 / ncols

		cols := []*ui.Row{}

		var wg sync.WaitGroup
		for _, repo := range man.Repos {

			widget := ui.NewGauge()
			widget.Height = 3
			widget.BarColor = ui.ColorGreen
			widget.BorderLabel = repo.Repo
			cols = append(cols, ui.NewCol(span, 0, widget))
			if len(cols) == ncols {
				ui.Body.AddRows(ui.NewRow(cols...))
				cols = []*ui.Row{}
			}

			wg.Add(1)
			go func(repo manifest.Repo, widget *ui.Gauge) {
				defer wg.Done()
				for prog := range vcs.Git.Clone(context, repo) {
					if prog.IsBegin {
						widget.Label = prog.Message
					}
					widget.Percent = int(float64(prog.Current) / float64(prog.Total) * 100)
					ui.Render(widget)
				}
			}(repo, widget)

		}
		ui.Body.AddRows(ui.NewRow(cols...))
		ui.Body.Align()
		ui.Render(ui.Body)
		wg.Wait()
	},
}

func init() {
	RootCmd.AddCommand(restoreCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// restoreCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// restoreCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
