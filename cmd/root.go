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

// Package cmd is the root command package for the ansible-host CLI. It
// wires together every subcommand, configures the Cobra root command,
// and provides the [Execute] entry point invoked by main.
//
// ansible-host is designed for managing real (non-development) hosts
// using Ansible. Unlike ansible-dev, which targets Vagrant-managed
// virtual machines, ansible-host operates against inventory-defined
// network hosts for provisioning, fact gathering, and ad-hoc execution.
//
// Subcommands are organized into individual packages under cmd/ and
// registered during init:
//
//   - create:    scaffold or generate host-related Ansible artifacts.
//   - execute:   run ad-hoc Ansible commands against inventory hosts.
//   - fact:      gather and display Ansible facts for hosts.
//   - inventory: display or inspect the Ansible host inventory.
//   - ping:      verify host reachability via Ansible ping.
//   - provision: apply roles and playbooks to configure hosts.
//   - restore:   install dependencies from requirements files.
//   - runbook:   execute predefined runbook playbooks against hosts.
package cmd

import (
	"fmt"
	"os"

	"github.com/dcjulian29/ansible-host/cmd/create"
	"github.com/dcjulian29/ansible-host/cmd/execute"
	"github.com/dcjulian29/ansible-host/cmd/fact"
	"github.com/dcjulian29/ansible-host/cmd/inventory"
	"github.com/dcjulian29/ansible-host/cmd/ping"
	"github.com/dcjulian29/ansible-host/cmd/provision"
	"github.com/dcjulian29/ansible-host/cmd/restore"
	"github.com/dcjulian29/ansible-host/cmd/runbook"
	"github.com/dcjulian29/go-toolbox/color"
	"github.com/spf13/cobra"
	"go.szostok.io/version/extension"
)

// rootCmd is the top-level Cobra command for the ansible-host CLI. When
// invoked without a subcommand it prints the help text, which includes
// an extended description of provisioning concepts and runbook workflows.
// Both SilenceErrors and SilenceUsage are enabled so that error
// formatting is handled exclusively by [Execute].
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

Runbooks are a set of detailed and repeatable procedures/tasks to standardize and automate common
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

// Execute is the main entry point for the CLI, called from main.main.
// It adds the auto-generated "version" subcommand (provided by
// go.szostok.io/version) with an upgrade notice for the
// "dcjulian29/ansible-host" GitHub repository, then delegates to
// [cobra.Command.Execute].
//
// If execution returns an error, the error message is printed to stderr
// using [color.Fatal] formatting and the process exits with code 1.
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

// init registers all subcommands on the root command during package
// initialization. Each subcommand is provided by a dedicated package
// under cmd/ and exposes a NewCommand factory function that returns a
// configured [cobra.Command].
func init() {
	rootCmd.AddCommand(create.NewCommand())
	rootCmd.AddCommand(execute.NewCommand())
	rootCmd.AddCommand(fact.NewCommand())
	rootCmd.AddCommand(inventory.NewCommand())
	rootCmd.AddCommand(ping.NewCommand())
	rootCmd.AddCommand(provision.NewCommand())
	rootCmd.AddCommand(restore.NewCommand())
	rootCmd.AddCommand(runbook.NewCommand())
}
