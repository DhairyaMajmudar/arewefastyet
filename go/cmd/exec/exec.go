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
	"github.com/spf13/cobra"
	"github.com/vitessio/arewefastyet/go/exec"
	"log"
)

func ExecCmd() *cobra.Command {
	ex, err := exec.NewExec()
	if err != nil {
		log.Fatal(err)
	}

	cmd := &cobra.Command{
		Use: "exec",
		Aliases: []string{"e"},
		Short: "Execute a task",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := ex.Infra.Prepare(); err != nil {
				return err
			}
			defer ex.InfraConfig.Close()

			out, err := ex.Infra.Create("device_public_ip")
			if err != nil {
				return err
			}
			log.Println(out["device_public_ip"])

			ex.Infra.Run(ex.AnsibleConfig)
			return nil
		},
	}

	ex.InfraConfig.AddToPersistentCommand(cmd)
	ex.AnsibleConfig.AddToPersistentCommand(cmd)
	ex.Infra.AddToCommand(cmd)
	return cmd
}