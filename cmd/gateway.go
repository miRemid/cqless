/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

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
		Run:   runGateway,
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

func getApiServer(ctx context.Context) *http.Server {
	route := gin.New()
	route.Use(gin.Recovery())
	route.Use(middleware.Logger("api"))
	if config.Gateway.EnablePprof {
		log.Info().Msg("开启pprof性能分析")
		pprof.Register(route)
	}
	v1 := route.Group("/api/v1")
	{
		cqless := v1.Group("/cqless")
		{
			cqless.POST("/function", gateway.MakeDeployHandler("", false))
			cqless.DELETE("/function", gateway.MakeRemoveHandler())
			cqless.GET("/function", gateway.MakeInspectHandler())
		}

		cq := v1.Group("/cqhttp")
		{
			cq.Match([]string{http.MethodGet, http.MethodPost}, "", cqhttp.GetDefaultCQHTTPManager().WebsocketHandler)
		}
	}
	server := &http.Server{
		Addr:           fmt.Sprintf(":%d", config.Gateway.APIPort),
		Handler:        route,
		ReadTimeout:    config.Gateway.ReadTimeout,
		WriteTimeout:   config.Gateway.WriteTimeout,
		MaxHeaderBytes: http.DefaultMaxHeaderBytes,
	}
	return server
}

func getProxyServer(ctx context.Context) *http.Server {
	route := gin.New()
	route.Use(gin.Recovery())
	route.Use(middleware.Logger("proxy"))
	if config.Gateway.EnablePprof {
		log.Info().Msg("开启pprof性能分析")
		pprof.Register(route)
	}
	proxyHandler := gateway.MakeProxyHandler(config.Proxy)

	proxy := route.Group("/")
	if config.Gateway.EnableRateLimit {
		proxy.Use(middleware.RateLimit(config.Gateway.RateLimit))
	}
	proxy.POST("/*proxyPath", proxyHandler)

	server := &http.Server{
		Addr:           fmt.Sprintf(":%d", config.Gateway.Port),
		Handler:        route,
		ReadTimeout:    config.Gateway.ReadTimeout,
		WriteTimeout:   config.Gateway.WriteTimeout,
		MaxHeaderBytes: http.DefaultMaxHeaderBytes,
	}
	return server
}

func runGateway(cmd *cobra.Command, args []string) {
	logger.InitLogger(config)
	if err := gateway.Init(config); err != nil {
		panic(err)
	}
	if err := cninetwork.Init(config); err != nil {
		panic(err)
	}

	if types.DEBUG == "FALSE" {
		gin.SetMode(gin.ReleaseMode)
	}

	ctx := context.Background()
	ctx, done := signal.NotifyContext(ctx, os.Interrupt, os.Kill)
	apiServer := getApiServer(ctx)
	proxyServer := getProxyServer(ctx)

	go apiServer.ListenAndServe()
	go proxyServer.ListenAndServe()
	<-ctx.Done()
	done()

	log.Info().Msg("收到退出信号")
	apictx, cancelApi := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelApi()
	log.Info().Msg("正在关闭API服务")
	apiServer.Shutdown(apictx)
	proxyctx, cancelProxy := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelProxy()
	log.Info().Msg("正在关闭Proxy服务")
	proxyServer.Shutdown(proxyctx)
}
