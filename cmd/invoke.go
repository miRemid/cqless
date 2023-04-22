/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"io"
	"net/http"

	"github.com/miRemid/cqless/pkg/types"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	// 优先处理配置文件
	var reqBody types.FunctionRequest
	if functionConfigPath != "" {
		// 1. read yaml file
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
	} else if functionName == "" {
		fmt.Println("未找到函数名称")
		return
	} else {
		reqBody.FunctionName = functionName
	}
	requestURI := fmt.Sprintf(cqless_invoke_api, httpClientGatewayAddress, config.Gateway.Port, reqBody.FunctionName)

	req, err := http.NewRequest(http.MethodPost, requestURI, nil)
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
