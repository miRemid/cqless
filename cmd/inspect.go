/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/miRemid/cqless/pkg/httputil"
	"github.com/spf13/cobra"
)

// inspectCmd represents the inspect command
var inspectCmd = &cobra.Command{
	Use:   "inspect",
	Short: "",
	Long:  ``,
	Run:   inspect,
}

func init() {
	functionCmd.AddCommand(inspectCmd)

	inspectCmd.Flags().StringVarP(&functionName, "function-name", "f", "", "inspect given function name container")
}

func inspect(cmd *cobra.Command, args []string) {
	var requestURI = fmt.Sprintf(cqless_function_api, httpClientGatewayAddress, config.Gateway.Port)
	req, err := http.NewRequest(http.MethodGet, requestURI, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	if functionName != "" {
		query := req.URL.Query()
		query.Add("fn", functionName)
		req.URL.RawQuery = query.Encode()
	}
	fmt.Println(req.URL.String())
	resp, err := httpClient.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()
	// var data bytes.Buffer
	// data, _ := io.ReadAll(resp.Body)
	// fmt.Println(string(data))
	var response httputil.Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		fmt.Println(err)
		return
	}
	fns := response.Data.([]interface{})
	tb := table.NewWriter()
	tb.AppendHeader(table.Row{"Name", "Full Name", "ID", "IP Address", "Status"})
	for _, f := range fns {
		fn := f.(map[string]interface{})
		tb.AppendRow(table.Row{fn["name"], fn["full_name"], fn["id"], fn["ip"], fn["status"]})
	}
	fmt.Println(tb.Render())
}
