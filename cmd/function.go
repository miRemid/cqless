/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/miRemid/cqless/pkg/types"
	"github.com/spf13/cobra"
)

// functionCmd represents the function command
var functionCmd = &cobra.Command{
	Use:   "func",
	Short: "",
	Long:  ``,
}

func init() {
	functionCmd.Run = functionCmd.HelpFunc()
	rootCmd.AddCommand(functionCmd)

	functionCmd.PersistentFlags().IntVar(&httpClientReadTimeout, "--read-timeout", 30, "http request read timeout")
	functionCmd.PersistentFlags().IntVar(&httpClientWriteTimeout, "--write-timeout", 30, "http request write timeout")
	functionCmd.PersistentFlags().StringVarP(&httpClientGatewayAddress, "--address", "a", "127.0.0.1", "gateway address")
	functionCmd.PersistentFlags().StringVarP(&httpClientGatewayConfigPath, "--config", "c", types.DEFAULT_CONFIG_PATH, "config path")
	functionCmd.PersistentFlags().StringVar(&functionNamespace, "--namespace", types.DEFAULT_FUNCTION_NAMESPACE, "function namespace")
}
