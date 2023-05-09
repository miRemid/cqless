/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"net/http"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/miRemid/cqless/pkg/cninetwork"
	"github.com/miRemid/cqless/pkg/cqhttp"
	"github.com/miRemid/cqless/pkg/gateway"
	"github.com/miRemid/cqless/pkg/logger"
	"github.com/miRemid/cqless/pkg/middleware"
	"github.com/miRemid/cqless/pkg/types"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

const NameExpression = "-a-zA-Z_0-9."

var (
	config     = types.GetConfig()
	gatewayCmd = &cobra.Command{
		Use:   "gateway",
		Short: "启动一个网关程序吧！",
	}
)

func init() {
	gatewayCmd.AddCommand(&cobra.Command{
		Use:   "up",
		Short: "Run gateway",
		Long:  "Run gateway",
		RunE:  runGateway,
	})
	gatewayCmd.AddCommand(&cobra.Command{
		Use:   "init",
		Short: "生成配置文件",
		Run:   runInitConfig,
	})
	rootCmd.AddCommand(gatewayCmd)
}

func runInitConfig(cmd *cobra.Command, args []string) {
	fmt.Printf("已生成配置文件至：%s\n", types.DEFAULT_CONFIG_PATH)
}

func runGateway(cmd *cobra.Command, args []string) error {
	logger.InitLogger(config)
	if err := gateway.Init(config); err != nil {
		return err
	}
	if err := cninetwork.Init(config); err != nil {
		return err
	}
	route := gin.New()
	route.Use(gin.Recovery())
	route.Use(middleware.Logger())
	if config.Gateway.EnablePprof {
		log.Info().Msg("开启pprof性能分析")
		pprof.Register(route)
	}
	proxyHandler := gateway.MakeProxyHandler(config.Proxy)

	cqless := route.Group("/cqless")
	{
		cqless.POST("/function", gateway.MakeDeployHandler("", false))
		cqless.DELETE("/function", gateway.MakeRemoveHandler())
		cqless.GET("/function", gateway.MakeInspectHandler())
	}

	function := route.Group("/function")
	if config.Gateway.EnableRateLimit {
		function.Use(middleware.RateLimit(config.Gateway.RateLimit))
	}

	{
		function.POST("/:name", proxyHandler)
		function.POST("/:name/:params", proxyHandler)
	}

	// CQHTTP Websocket
	cq := route.Group("/cqhttp")
	{
		cq.Match([]string{http.MethodGet, http.MethodPost}, "", cqhttp.GetDefaultCQHTTPManager().WebsocketHandler)
	}

	server := &http.Server{
		Addr:           fmt.Sprintf(":%d", config.Gateway.Port),
		Handler:        route,
		ReadTimeout:    config.Gateway.ReadTimeout,
		WriteTimeout:   config.Gateway.WriteTimeout,
		MaxHeaderBytes: http.DefaultMaxHeaderBytes,
	}
	log.Info().Str("msg", fmt.Sprintf("start listen at port: %d", config.Gateway.Port)).Send()
	return server.ListenAndServe()
}
