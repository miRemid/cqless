/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package function

import (
	"context"
	"fmt"
	"os"
	"time"

	v1 "github.com/miRemid/cqless/v1/pkg/pb"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// rmCmd represents the rm command
var rmCmd = &cobra.Command{
	Use:   "remove",
	Short: "删除目标函数",
	Run:   remove,
}

func init() {
	rmCmd.Flags().StringVar(&functionName, "fn", "", "function name needs to be deleted")
}

func remove(cmd *cobra.Command, args []string) {

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
			fmt.Printf("read config file failed: %v\n", err)
			os.Exit(1)
		}
	} else if functionName == "" {
		fmt.Println("function name can not be empty!")
		os.Exit(1)
	} else {
		reqBody.Name = functionName
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
	_, err = client.DeleteFunction(ctx, &v1.DeleteFunctionRequest{Name: reqBody.Name})
	if err != nil {
		fmt.Println("delete function failed: ", err)
		os.Exit(2)
	}
}
