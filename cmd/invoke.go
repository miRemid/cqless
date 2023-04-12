/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/miRemid/cqless/pkg/types"
	"github.com/spf13/cobra"
)

// invokeCmd represents the invoke command
var invokeCmd = &cobra.Command{
	Use:   "invoke",
	Short: "invoke a function",
	Run:   invoke,
}

func init() {
	functionCmd.AddCommand(invokeCmd)

	invokeCmd.Flags().StringVarP(&functionName, "function-name", "f", "", "function name")
}

func invoke(cmd *cobra.Command, args []string) {
	var buffer bytes.Buffer
	var reqBody types.FunctionInvokeRequest
	reqBody.FunctionName = functionName
	if err := json.NewEncoder(&buffer).Encode(reqBody); err != nil {
		fmt.Println(err)
		return
	}
	requestURI := fmt.Sprintf(cqless_invoke_api, httpClientGatewayAddress, config.Gateway.Port, functionName)
	req, err := http.NewRequest(http.MethodPost, requestURI, &buffer)
	if err != nil {
		fmt.Println(err)
		return
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	fmt.Println(string(data))
}
