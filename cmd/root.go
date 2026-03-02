/*
Copyright © 2023 Julian Easterling julian@julianscorner.com

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/dcjulian29/ansible-host/cmd/create"
	"github.com/dcjulian29/ansible-host/cmd/execute"
	"github.com/dcjulian29/ansible-host/cmd/fact"
	"github.com/dcjulian29/go-toolbox/color"
	"github.com/spf13/cobra"
	"go.szostok.io/version/extension"
)

var rootCmd = &cobra.Command{
	Use:   "ansible-host",
	Short: "Provision, ping, play runbooks, and execute command utilizing Ansible",
	Long: `Provision, ping, play runbooks, and execute command utilizing Ansible:

This tool can be used for configuring and maintaining a computer system in a network environment.
Typically involved tasks are: installing updates, managing user accounts, configuring security
settings, and ensuring that the system is running smoothly. Depending on the type of host and the
specific requirements, managing a host can be a complex and time-consuming task. This tools aims to
simplify that.

Provisioning is a process of setting up and configuring computer systems or servers to meet
specific requirements. This typically involves installing and configuring software packages,
setting up user accounts and permissions, and configuring security settings. Provisioning allows
organizations to quickly and easily deploy new servers and services on-demand, speeding up the
process of application development and deployment. It also ensures that systems are configured
consistently and according to best practices, reducing the risk of errors and vulnerabilities.

Runbooks are a set of detailed and repeatable procedures tasks to standardize and automate common
tasks and processes. These procedures may include steps for infrastructure and hardware validation,
troubleshooting issues, and more. By using runbooks, one can ensure that tasks are completed
consistently and efficiently, reducing the risk of errors and downtime.`,
	SilenceErrors: true,
	SilenceUsage:  true,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}

		return nil
	},
}

func Execute() {
	rootCmd.AddCommand(
		extension.NewVersionCobraCmd(
			extension.WithUpgradeNotice("dcjulian29", "ansible-host"),
		),
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "\n"+color.Fatal(err.Error()))
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(create.NewCommand())
	rootCmd.AddCommand(execute.NewCommand())
	rootCmd.AddCommand(fact.NewCommand())
}
