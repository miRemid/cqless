/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package function

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/miRemid/cqless/pkg/types"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// invokeCmd represents the invoke command
var invokeCmd = &cobra.Command{
	Use:   "invoke",
	Short: "调用一个函数接口",
	Run:   invoke,
}

func init() {
	invokeCmd.Flags().StringVar(&functionName, "fn", "", "需要调用的函数名称")
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
	requestURI := fmt.Sprintf(cqless_invoke_api, httpClientGatewayAddress, httpClientGatewayPort, reqBody.FunctionName)
	ctx, cancel := context.WithTimeout(context.TODO(), time.Duration(httpTimeout)*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURI, nil)
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
