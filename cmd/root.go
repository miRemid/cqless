/*
Copyright © 2023 miRemid

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
	"net/http"
	"os"

	"github.com/miRemid/cqless/pkg/types"
	"github.com/spf13/cobra"
)

var (
	config     = types.GetConfig()
	httpClient *http.Client
)

var ()

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "cqless",
	Short: "Container based serverless platform for CQHTTP",
	Long:  `CQLESS is a serverless platform for CQHTTP.`,
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().IntVar(&httpClientReadTimeout, "--read-timeout", 30, "http request read timeout")
	rootCmd.PersistentFlags().IntVar(&httpClientWriteTimeout, "--write-timeout", 30, "http request write timeout")
	rootCmd.PersistentFlags().StringVarP(&httpClientGatewayAddress, "--address", "a", "127.0.0.1", "gateway address")
	rootCmd.PersistentFlags().StringVarP(&httpClientGatewayConfigPath, "--config", "c", types.DEFAULT_CONFIG_PATH, "config path")
	rootCmd.PersistentFlags().StringVar(&functionNamespace, "--namespace", types.DEFAULT_FUNCTION_NAMESPACE, "function namespace")
}

func Execute() {

	httpClient = &http.Client{}

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
