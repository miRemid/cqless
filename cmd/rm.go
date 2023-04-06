/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/miRemid/cqless/pkg/httputil"
	"github.com/miRemid/cqless/pkg/types"
	"github.com/spf13/cobra"
)

// rmCmd represents the rm command
var rmCmd = &cobra.Command{
	Use:   "rm",
	Short: "remove function",
	Run:   remove,
}

func init() {
	rootCmd.AddCommand(rmCmd)

	rmCmd.Flags().StringVarP(&functionName, "function-name", "n", "", "function name")
}

func remove(cmd *cobra.Command, args []string) {
	var buffer bytes.Buffer
	var reqBody types.FunctionRemoveRequest
	reqBody.FunctionName = functionName
	if err := json.NewEncoder(&buffer).Encode(reqBody); err != nil {
		fmt.Println(err)
		return
	}
	requestURI := fmt.Sprintf(cqless_function_api, httpClientGatewayAddress, config.Gateway.Port)
	req, err := http.NewRequest(http.MethodDelete, requestURI, &buffer)
	if err != nil {
		fmt.Println(err)
		return
	}
	req.URL.Query().Add("namespace", functionNamespace)
	resp, err := httpClient.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()
	var response httputil.Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(response.Message)
}
