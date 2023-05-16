/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package function

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/miRemid/cqless/pkg/httputil"
	"github.com/miRemid/cqless/pkg/types"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// rmCmd represents the rm command
var rmCmd = &cobra.Command{
	Use:   "remove",
	Short: "删除目标函数",
	Run:   remove,
}

func init() {
	rmCmd.Flags().StringVar(&functionName, "fn", "", "需要删除的函数名称")
	rmCmd.Flags().IntVarP(&httpClientGatewayPort, "port", "p", 5566, "调用端口，默认5566")
}

func remove(cmd *cobra.Command, args []string) {

	var reqBody types.FunctionRemoveRequest
	// 优先处理配置文件
	if functionConfigPath != "" {
		// 1. read yaml file
		functionConfigReader.SetConfigFile(functionConfigPath)
		// functionConfigReader.AddConfigPath(functionConfigPath)
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

	var buffer bytes.Buffer
	if err := json.NewEncoder(&buffer).Encode(reqBody); err != nil {
		fmt.Println(err)
		return
	}
	requestURI := getApiRequestURI()
	ctx, cancel := context.WithTimeout(context.TODO(), time.Duration(httpTimeout)*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, requestURI, &buffer)
	if err != nil {
		fmt.Println(err)
		return
	}
	query := req.URL.Query()
	query.Add("namespace", functionNamespace)
	req.URL.RawQuery = query.Encode()

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
	if response.Code != httputil.StatusOK {
		fmt.Println(response.Message)
	}
}
