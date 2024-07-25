/*
 *
 * Copyright 2021 The Vitess Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 * /
 */

package exec

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/vitessio/arewefastyet/go/exec"
)

func ExecCmd() *cobra.Command {
	ex, err := exec.NewExec("")
	if err != nil {
		log.Fatal(err)
	}

	cmd := &cobra.Command{
		Use:     "exec",
		Aliases: []string{"e"},
		Short:   "Execute a task",
		Long: `Execute a task based on the given terraform and ansible configuration.
It handles the creation, configuration, and cleanup of the infrastructure.`,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			defer func() {
				if errSuccess := ex.Success(); errSuccess != nil {
					err = errSuccess
				}
			}()

			// prepare
			if err = ex.Prepare(); err != nil {
				return
			}

			// execute
			err = ex.Execute()
			return
		},
	}

	ex.AddToCommand(cmd)
	return cmd
}
