/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/miRemid/cqless/cmd/function"
	"github.com/spf13/cobra"
)

// functionCmd represents the function command
var functionCmd = &cobra.Command{
	Use:   "func",
	Short: "function crud",
}

func init() {
	function.Init(functionCmd)
	rootCmd.AddCommand(functionCmd)
}
