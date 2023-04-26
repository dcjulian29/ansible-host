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
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.szostok.io/version/extension"
)

var (
	cfgFile          string
	folderPath       string
	workingDirectory string

	rootCmd = &cobra.Command{
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
	}
)

func Execute() error {
	workingDirectory, _ = os.Getwd()

	rootCmd.AddCommand(
		extension.NewVersionCobraCmd(
			extension.WithUpgradeNotice("dcjulian29", "ansible-host"),
		),
	)

	return rootCmd.Execute()
}

func init() {
	pwd, _ := os.Getwd()

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ~/.ansible-host.yaml)")
	rootCmd.PersistentFlags().StringVar(&folderPath, "path", pwd, "path to Ansible folder")
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".ansible-host")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

func ensureAnsibleDirectory() {
	if workingDirectory != folderPath {
		if err := os.Chdir(folderPath); err != nil {
			fmt.Fprintln(os.Stderr, "Unable to access path!")
			cobra.CheckErr(err)
		}
	}
}

func ensurefileExists(filename string, errorMsg string) {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) || info.IsDir() {
		cobra.CheckErr(errors.New(errorMsg))
	}
}

func ensureWorkingDirectory() {
	if workingDirectory != folderPath {
		if err := os.Chdir(workingDirectory); err != nil {
			cobra.CheckErr(err)
		}
	}
}

func executeExternalProgram(program string, params ...string) {
	executeExternalProgramEnv(program, []string{""}, params...)
}

func executeExternalProgramEnv(program string, env []string, params ...string) {
	cmd := exec.Command(program, params...)
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Env = append(os.Environ(), env...)

	if err := cmd.Run(); err != nil {
		ensureWorkingDirectory()
		cobra.CheckErr(err)
	}
}

func executeCommand(program string, env []string, params ...string) string {
	cmd := exec.Command(program, params...)
	cmd.Stdin = os.Stdin
	cmd.Env = append(os.Environ(), env...)

	output, err := cmd.CombinedOutput()

	if err != nil {
		ensureWorkingDirectory()
		cobra.CheckErr(string(output[:]))
	}

	return string(output[:])
}
