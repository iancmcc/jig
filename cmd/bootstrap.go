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

	"github.com/spf13/cobra"
)

var (
	alias   = "cdj"
	jigroot = ""
)

var bootstrap = `
%s() {
	CDDIR="$@"
	if [ -z "$CDDIR" ]; then
		cd $(%s jig root)
	else
		cd $(%s jig ls -n1 $CDDIR)
	fi
}
`

// bootstrapCmd represents the bootstrap command
var bootstrapCmd = &cobra.Command{
	Use:   "bootstrap",
	Short: "Install jig tools into your shell",
	Run: func(cmd *cobra.Command, args []string) {
		if jigroot != "" {
			jigroot = fmt.Sprintf("JIGROOT=%s", jigroot)
		}
		fmt.Printf(bootstrap, alias, jigroot, jigroot)
	},
}

func init() {
	RootCmd.AddCommand(bootstrapCmd)
	bootstrapCmd.Flags().StringVarP(&alias, "cd-command", "c", "cdj", "The command name to use for changing dirs")
	bootstrapCmd.Flags().StringVarP(&jigroot, "with-jigroot", "j", "", "Use a custom jig root for this evaluation")
}
