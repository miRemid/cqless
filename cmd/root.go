/*
Copyright © 2023 miRemid

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
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
	"github.com/miRemid/cqless/pkg/provider"
	"github.com/miRemid/cqless/pkg/proxy"
	"github.com/miRemid/cqless/pkg/resolver"
	"github.com/miRemid/cqless/pkg/types"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "cqless",
	Short: "=w=",
	Long:  `CQLESS命令行版`,
}

var config *types.CQLessConfig

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "up",
		Short: "start api and proxy server",
		Run:   runUP,
	})
	rootCmd.AddCommand(&cobra.Command{
		Use:   "init",
		Short: "generate config files",
		Run:   runInitConfig,
	})

}

func runInitConfig(cmd *cobra.Command, args []string) {
	types.GetConfig()
	fmt.Printf("已生成配置文件至：%s\n", types.DEFAULT_CONFIG_PATH)
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func setupApiServer(ctx context.Context) *http.Server {
	route := gin.New()
	route.Use(gin.Recovery())
	route.Use(middleware.Logger("api"))
	if types.GetConfig().Gateway.EnablePprof {
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
			cq.Match([]string{http.MethodGet, http.MethodPost}, "", cqhttp.WebsocketHandler())
		}
	}
	server := &http.Server{
		Addr:           fmt.Sprintf(":%d", config.Gateway.Port),
		Handler:        route,
		ReadTimeout:    config.Gateway.ReadTimeout,
		WriteTimeout:   config.Gateway.WriteTimeout,
		MaxHeaderBytes: http.DefaultMaxHeaderBytes,
	}
	return server
}

func setupProxyServer(ctx context.Context) *http.Server {
	route := gin.New()
	route.Use(gin.Recovery())
	route.Use(middleware.Logger("proxy"))

	v1 := route.Group("/")
	if config.Gateway.EnableRateLimit {
		v1.Use(middleware.RateLimit(config.Gateway.RateLimit))
	}
	v1.Any("/:funcName/*requestPath", proxy.ReverseHandler())

	server := &http.Server{
		Addr:           fmt.Sprintf(":%d", config.Proxy.Port),
		Handler:        route,
		ReadTimeout:    config.Gateway.ReadTimeout,
		WriteTimeout:   config.Gateway.WriteTimeout,
		MaxHeaderBytes: http.DefaultMaxHeaderBytes,
	}
	return server
}

func runUP(cmd *cobra.Command, args []string) {
	config = types.GetConfig()
	logger.InitLogger(config.Logger, "cqless.log")
	systemLogger := log.Hook(logger.ModuleHook("system"))

	if err := cninetwork.Init(config.Network); err != nil {
		panic(err)
	}

	if err := resolver.Init(config.Resolver); err != nil {
		panic(err)
	}

	if err := gateway.Init(config.Gateway); err != nil {
		panic(err)
	}

	if err := proxy.Init(config.Proxy); err != nil {
		panic(err)
	}

	if err := provider.Init(config.Provider); err != nil {
		panic(err)
	}

	if err := cqhttp.Init(config.CQHTTP); err != nil {
		panic(err)
	}

	if types.DEBUG == "FALSE" {
		gin.SetMode(gin.ReleaseMode)
	}

	ctx := context.Background()
	ctx, done := signal.NotifyContext(ctx, os.Interrupt, os.Kill)

	apiServer := setupApiServer(ctx)
	proxyServer := setupProxyServer(ctx)
	go apiServer.ListenAndServe()
	go proxyServer.ListenAndServe()
	<-ctx.Done()
	done()

	systemLogger.Info().Msg("监听到退出信号")
	apictx, cancelApi := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelApi()
	log.Info().Msg("正在关闭API服务")
	apiServer.Shutdown(apictx)
	proxyctx, cancelProxy := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelProxy()
	log.Info().Msg("正在关闭Proxy服务")
	proxyServer.Shutdown(proxyctx)
}
