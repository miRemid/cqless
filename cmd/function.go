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
	Short: "函数相关命令",
	Long:  `调用网关执行函数相关命令，包括创建、获取、执行等`,
}

func init() {
	function.Init(functionCmd)
	rootCmd.AddCommand(functionCmd)
}
