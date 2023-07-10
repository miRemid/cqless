/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package function

import (
	"context"
	"fmt"
	"os"
	"time"

	v1 "github.com/miRemid/cqless/pkg/pb"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "deploy a function",
	Run:   deploy,
}

var (
	deployFunctionName   string
	deployFunctionImage  string
	deployFunctionPort   string
	deployFunctionScheme string
)

func init() {
	deployCmd.Flags().StringVarP(&deployFunctionName, "name", "n", "", "function name")
	deployCmd.Flags().StringVarP(&deployFunctionImage, "image", "i", "", "image name")
	deployCmd.Flags().StringVar(&deployFunctionScheme, "scheme", "http", "function scheme，default http")
	deployCmd.Flags().StringVar(&deployFunctionPort, "fn", "8080", "function watchdog port")
}

func deploy(cmd *cobra.Command, args []string) {
	// 1. read yaml file
	var reqBody v1.Function
	if functionConfigPath != "" {
		functionConfigReader.SetConfigFile(functionConfigPath)
		if err := functionConfigReader.ReadInConfig(); err != nil {
			fmt.Printf("read config file failed: %v\n", err)
			os.Exit(1)
		}
		if err := functionConfigReader.Unmarshal(&reqBody, viper.DecoderConfigOption(func(dc *mapstructure.DecoderConfig) {
			dc.TagName = "json"
		})); err != nil {
			fmt.Printf("read deploy config file failed: %v\n", err)
			os.Exit(1)
		}
	} else {
		if deployFunctionImage == "" || deployFunctionName == "" {
			fmt.Println("invalid request params, please check function image and function name")
			os.Exit(1)
		}
		reqBody.Image = deployFunctionImage
		reqBody.Name = deployFunctionName
		reqBody.WatchDogPort = deployFunctionPort
		reqBody.Scheme = deployFunctionScheme
	}

	conn, err := grpc.Dial(grpcServerAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Println("connect to cqless failed: ", err)
		os.Exit(2)
	}
	defer conn.Close()
	client := v1.NewFunctionServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()
	function, err := client.CreateFunction(ctx, &v1.CreateFunctionRequest{Function: &reqBody})
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
	fmt.Printf("create function: %s!\n", function.Name)
}
