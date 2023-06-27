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

	"github.com/go-playground/validator/v10"
	"github.com/miRemid/cqless/pkg/httputil"
	"github.com/miRemid/cqless/pkg/types"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "部署",
	Run:   deploy,
}

var (
	deployFunctionName   string
	deployFunctionImage  string
	deployFunctionPort   string
	deployFunctionScheme string
)

func init() {
	deployCmd.Flags().StringVarP(&deployFunctionName, "name", "n", "", "函数名称")
	deployCmd.Flags().StringVarP(&deployFunctionImage, "image", "i", "", "容器镜像名称")
	deployCmd.Flags().StringVar(&deployFunctionScheme, "scheme", "http", "容器服务协议，默认http")
	deployCmd.Flags().StringVar(&deployFunctionPort, "fn", "8080", "函数服务监听端口")
	deployCmd.Flags().IntVarP(&httpClientGatewayPort, "port", "p", 5566, "调用端口，默认5566")
}

func deploy(cmd *cobra.Command, args []string) {

	// 1. read yaml file
	var reqBody types.FunctionCreateRequest
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
		if deployFunctionImage == "" || deployFunctionName == "" {
			fmt.Println("请求参数错误，确实函数名称或容器镜像名称")
			return
		}
		reqBody.Image = deployFunctionImage
		reqBody.Name = deployFunctionName
		reqBody.WatchDogPort = deployFunctionPort
		reqBody.Scheme = deployFunctionScheme
	}
	var buffer bytes.Buffer
	if err := json.NewEncoder(&buffer).Encode(reqBody); err != nil {
		fmt.Printf("序列化请求失败: %v\n", err)
		return
	}
	validate := validator.New()
	if err := validate.Struct(reqBody); err != nil {
		fmt.Println(err)
		return
	}
	requestURI := getApiRequestURI()
	ctx, cancel := context.WithTimeout(context.TODO(), time.Duration(httpTimeout)*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURI, &buffer)
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
	var response httputil.Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		fmt.Println(err)
		return
	}
	if response.Code != httputil.StatusOK {
		fmt.Println("创建函数失败: ", response.Message)
	}
}
