/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package function

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/miRemid/cqless/pkg/httputil"
	"github.com/miRemid/cqless/pkg/types"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// inspectCmd represents the inspect command
var inspectCmd = &cobra.Command{
	Use:   "inspect",
	Short: "检查",
	Run:   inspect,
}

func init() {
	inspectCmd.Flags().StringVar(&functionName, "fn", "", "需要检查的函数名称")
}

func inspect(cmd *cobra.Command, args []string) {
	var reqBody types.FunctionInspectRequest
	if functionConfigPath != "" {
		functionConfigReader.SetConfigFile(functionConfigPath)
		if err := functionConfigReader.ReadInConfig(); err != nil {
			fmt.Printf("读取部署文件配置失败: %v\n", err)
			return
		}
		if err := functionConfigReader.Unmarshal(&reqBody, viper.DecoderConfigOption(func(dc *mapstructure.DecoderConfig) {
			dc.TagName = "json"
		})); err != nil {
			fmt.Printf("读取部署文件配置失败: %v\n", err)
			return
		}
	} else {
		reqBody.FunctionName = functionName
	}

	requestURI := getApiRequestURI()
	ctx, cancel := context.WithTimeout(context.TODO(), time.Duration(httpTimeout)*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURI, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	if reqBody.FunctionName != "" {
		query := req.URL.Query()
		query.Add("fn", functionName)
		req.URL.RawQuery = query.Encode()
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	var response httputil.Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		fmt.Printf("处理查询数据错误: %v", err)
		return
	}
	if fns, ok := response.Data.([]interface{}); ok {
		tb := table.NewWriter()
		tb.AppendHeader(table.Row{"Name", "Full Name", "ID", "IP Address", "Status"})
		for _, f := range fns {
			fn := f.(map[string]interface{})
			tb.AppendRow(table.Row{fn["name"], fn["full_name"], fn["id"], fn["ip"], fn["status"]})
		}
		fmt.Println(tb.Render())
	}
}
