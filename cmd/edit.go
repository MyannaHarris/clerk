/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

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
	"github.com/mmerkes/clerk/pkg"
	"github.com/spf13/cobra"
)

// editCmd represents the edit command
var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit a task",
	Long:  `Edit an existing task.`,
	Run: func(cmd *cobra.Command, args []string) {
		clerk.EditTask(id)
	},
}

func init() {
	rootCmd.AddCommand(editCmd)

	editCmd.PersistentFlags().IntVarP(&id, "id", "i", -1, "Id of task to edit")
	editCmd.MarkFlagRequired("id")
}
