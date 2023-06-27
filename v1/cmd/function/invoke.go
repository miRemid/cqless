/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package function

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/miRemid/cqless/v1/pkg/types"
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

var (
	functionProxyAddress   string
	functionInvokeData     string
	functionInvokeDataType string
)

func init() {
	invokeCmd.Flags().StringVar(&functionName, "fn", "", "需要调用的函数名称")
	invokeCmd.Flags().StringVarP(&functionProxyAddress, "address", "a", "127.0.0.1", "代理网关地址")
	invokeCmd.Flags().StringVarP(&functionInvokeData, "data", "d", "", "需要发送的数据")
	invokeCmd.Flags().StringVarP(&functionInvokeDataType, "type", "t", "", "需要发送的数据格式，json或者form")
	invokeCmd.Flags().IntVarP(&httpClientProxyPort, "port", "p", 5567, "调用端口，默认5567")
}

func invoke(cmd *cobra.Command, args []string) {
	// 优先处理配置文件
	var reqBody types.FunctionInvokeRequest
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
	requestURI := fmt.Sprintf(cqless_invoke_api, functionProxyAddress, httpClientProxyPort, reqBody.FunctionName)
	ctx, cancel := context.WithTimeout(context.TODO(), time.Duration(timeout)*time.Second)
	defer cancel()
	var body io.Reader = nil
	var contentType string
	if functionInvokeData != "" && functionInvokeDataType != "" {
		switch strings.ToUpper(functionInvokeDataType) {
		case "JSON":
			var tmpData = make(map[string]interface{})
			if err := json.Unmarshal([]byte(functionInvokeData), &tmpData); err != nil {
				fmt.Printf("提供的数据格式错误: %v\n", err)
				return
			}
			data, err := json.Marshal(tmpData)
			if err != nil {
				fmt.Printf("提供的数据格式错误: %v\n", err)
				return
			}
			body = bytes.NewBuffer(data)
			contentType = "application/json"
		case "FORM":
			// eg: param1=a;param2=b
			params := strings.Split(functionInvokeData, ";")
			form := url.Values{}
			for _, param := range params {
				kv := strings.Split(param, "=")
				form.Add(kv[0], kv[1])
			}
			body = strings.NewReader(form.Encode())
			contentType = "application/x-www-form-urlencoded"
		case "TEXT":
			body = strings.NewReader(functionInvokeData)
			contentType = "text/plain"
		default:
			fmt.Printf("不支持的数据格式: %v，目前仅支持json、form和text", functionInvokeDataType)
			return
		}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURI, body)
	if err != nil {
		fmt.Println(err)
		return
	}
	if contentType != "" {
		req.Header.Add("Content-Type", contentType)
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	fmt.Println(string(data), resp.StatusCode)
}
