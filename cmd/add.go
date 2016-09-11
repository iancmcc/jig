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
	"github.com/Sirupsen/logrus"
	"github.com/iancmcc/jig/config"
	"github.com/iancmcc/jig/vcs"
	"github.com/spf13/cobra"
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a repository to be tracked by jig",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			logrus.Fatal("Must pass path to a repository to be added to the manifest")
		}
		root, err := config.FindJigRoot()
		if err != nil {
			logrus.Fatal("No jig root found. Use 'jig init' to create one.")
		}
		manifest, err := config.JigRootManifest()
		if err != nil {
			manifest = &config.Manifest{
				Repos: []*config.Repo{},
			}
		}
		target := args[0]
		repo, err := vcs.RepoFromPath(target)
		if err != nil {
			logrus.Fatal("Not a path to a valid git repository")
		}
		shortname, err := vcs.RepoToPath(repo.Repo)
		if err != nil {
			logrus.WithField("uri", repo.Repo).Fatal("Unable to parse repository URI")
		}
		var found bool
		for i, r := range manifest.Repos {
			sname, err := vcs.RepoToPath(r.Repo)
			if err != nil {
				continue
			}
			if sname == shortname {
				manifest.Repos[i] = repo
				found = true
				break
			}
		}
		if !found {
			manifest.Repos = append(manifest.Repos, repo)
		}
		manifest.Save(root)
	},
}

func init() {
	RootCmd.AddCommand(addCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// addCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// addCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
