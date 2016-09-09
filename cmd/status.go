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
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"text/tabwriter"

	"github.com/Sirupsen/logrus"
	"github.com/iancmcc/jig/config"
	"github.com/iancmcc/jig/vcs"
	"github.com/spf13/cobra"
)

var (
	statall bool
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Print status of the repositories in your manifest",
	Run: func(cmd *cobra.Command, args []string) {
		root, err := config.FindClosestJigRoot("")
		if err != nil {
			logrus.Fatal("No jig root found. Use 'jig init' to create one.")
		}
		manifest, err := config.DefaultManifest("")
		if err != nil {
			return
		}
		statuschan := make(chan *vcs.Status)
		var (
			wg     sync.WaitGroup
			maxlen int
		)
		for _, r := range manifest.Repos {
			dir, err := vcs.RepoToPath(r.Repo)
			if err != nil {
				logrus.WithField("repo", r.Repo).Error("Unable to parse repo")
				continue
			}
			l := len(dir)
			if maxlen < l {
				maxlen = l
			}
			wg.Add(1)
			go func(repo *config.Repo, dir string) {
				defer wg.Done()
				log := logrus.WithField("repo", repo.Repo)
				dir = filepath.Join(root, dir)
				dir, err = filepath.Abs(dir)
				if err != nil {
					log.WithError(err).Error("Unable to get status for repo")
					return
				}
				stat, err := vcs.Git.Status(repo, dir)
				if err != nil {
					log.WithError(err).Error("Unable to get status for repo")
					return
				}
				statuschan <- stat
			}(r, dir)
		}
		go func() {
			wg.Wait()
			close(statuschan)
		}()

		w := tabwriter.NewWriter(os.Stdout, 0, 5, 4, ' ', 0)
		fmt.Fprintf(w, "Repo\tRef (Orig)\tStaged\tUnstaged\tUntracked\n")
		branched := []*vcs.Status{}
		changed := []*vcs.Status{}
		ordinary := []*vcs.Status{}
		print := func(stat *vcs.Status) {
			var (
				unstaged, untracked, staged string
			)
			if stat.Staged {
				staged = "*"
			}
			if stat.Unstaged {
				unstaged = "*"
			}
			if stat.Untracked {
				untracked = "*"
			}
			var orig string
			if stat.Branch != stat.OrigRef {
				orig = fmt.Sprintf(" (%s)", stat.OrigRef)
			}
			fmt.Fprintf(w, "%s\t%s%s\t%s\t%s\t%s\n", stat.Repo, stat.Branch, orig, staged, unstaged, untracked)
		}
		for stat := range statuschan {
			ischanged := stat.Staged || stat.Unstaged || stat.Untracked
			isbranched := stat.Branch != stat.OrigRef
			if isbranched && ischanged {
				print(stat)
				continue
			}
			if isbranched {
				branched = append(branched, stat)
				continue
			}
			if ischanged {
				changed = append(changed, stat)
				continue
			}
			ordinary = append(ordinary, stat)
		}
		for _, stat := range changed {
			print(stat)
		}
		for _, stat := range branched {
			print(stat)
		}
		if statall {
			for _, stat := range ordinary {
				print(stat)
			}
		}
		w.Flush()
	},
}

func init() {
	RootCmd.AddCommand(statusCmd)
	statusCmd.PersistentFlags().BoolVarP(&statall, "all", "a", false, "Show status for all repositories, not just those with chnages")
}
