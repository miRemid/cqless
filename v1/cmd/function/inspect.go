/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package function

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	v1 "github.com/miRemid/cqless/v1/pkg/pb"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// inspectCmd represents the inspect command
var inspectCmd = &cobra.Command{
	Use:   "inspect",
	Short: "get function infomation",
	Run:   inspect,
}

func init() {
	inspectCmd.Flags().StringVar(&functionName, "fn", "", "需要检查的函数名称")
}

func inspect(cmd *cobra.Command, args []string) {
	var reqBody v1.GetFunctionRequest
	if functionConfigPath != "" {
		functionConfigReader.SetConfigFile(functionConfigPath)
		if err := functionConfigReader.ReadInConfig(); err != nil {
			fmt.Printf("read function config faield: %v\n", err)
			os.Exit(1)
		}
		if err := functionConfigReader.Unmarshal(&reqBody, viper.DecoderConfigOption(func(dc *mapstructure.DecoderConfig) {
			dc.TagName = "json"
		})); err != nil {
			fmt.Printf("read function config failed: %v\n", err)
			os.Exit(1)
		}
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
	resp, err := client.GetFunction(ctx, &reqBody)
	if err != nil {
		fmt.Println("get functions failed: ", err)
		os.Exit(2)
	}
	tb := table.NewWriter()
	tb.AppendHeader(table.Row{"Name", "Full Name", "ID", "IP Address", "Status"})
	for _, f := range resp.Functions {
		tb.AppendRow(table.Row{f.Name, f.FullName, f.Id, f.IpAddress, f.Status})
	}
	fmt.Println(tb.Render())
}
