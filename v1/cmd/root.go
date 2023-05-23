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
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/miRemid/cqless/v1/pkg/cninetwork"
	"github.com/miRemid/cqless/v1/pkg/cqhttp"
	"github.com/miRemid/cqless/v1/pkg/gateway"
	"github.com/miRemid/cqless/v1/pkg/logger"
	"github.com/miRemid/cqless/v1/pkg/middleware"
	v1 "github.com/miRemid/cqless/v1/pkg/pb"
	"github.com/miRemid/cqless/v1/pkg/provider"
	"github.com/miRemid/cqless/v1/pkg/proxy"
	"github.com/miRemid/cqless/v1/pkg/resolver"
	"github.com/miRemid/cqless/v1/pkg/types"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "cqless",
	Short: "=w=",
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
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "get current version",
		Run:   runGetVersion,
	})
}

func runInitConfig(cmd *cobra.Command, args []string) {
	types.GetConfig()
	fmt.Printf("generate config files：%s\n", types.DEFAULT_CONFIG_PATH)
}

func runGetVersion(cmd *cobra.Command, args []string) {
	fmt.Printf("version：%s\n", types.CQLESS_VERSION)
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func setupApiServer(ctx context.Context) (*http.Server, net.Listener, *grpc.Server) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	gw := gateway.New()
	if err := gw.Init(config.Gateway); err != nil {
		panic(err)
	}
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	lis, err := net.Listen("tcp", config.Gateway.Address)
	if err != nil {
		panic(err)
	}
	grpcServer := grpc.NewServer()
	v1.RegisterFunctionServiceServer(grpcServer, gw)
	if err := v1.RegisterFunctionServiceHandlerFromEndpoint(ctx, mux, config.Gateway.Address, opts); err != nil {
		panic(err)
	}
	mux.HandlePath("POST", "/cqhttp", cqhttp.WebsocketHandler())
	server := &http.Server{
		Addr:           config.Gateway.HTTPAddress,
		Handler:        mux,
		ReadTimeout:    config.Gateway.ReadTimeout,
		WriteTimeout:   config.Gateway.WriteTimeout,
		MaxHeaderBytes: http.DefaultMaxHeaderBytes,
	}
	return server, lis, grpcServer
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
		Addr:           config.Proxy.Address,
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

	apiServer, lis, grpcServer := setupApiServer(ctx)
	proxyServer := setupProxyServer(ctx)
	systemLogger.Info().Str("address", config.Gateway.HTTPAddress).Msg("start HTTP API server")
	go apiServer.ListenAndServe()
	systemLogger.Info().Str("address", config.Gateway.Address).Msg("start GRPC API server")
	go grpcServer.Serve(lis)
	systemLogger.Info().Str("address", config.Proxy.Address).Msg("start HTTP Proxy server")
	go proxyServer.ListenAndServe()
	<-ctx.Done()
	done()

	systemLogger.Info().Msg("detect close signal")
	apictx, cancelApi := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelApi()
	systemLogger.Info().Msg("Closing HTTP API server...")
	apiServer.Shutdown(apictx)
	systemLogger.Info().Msg("Closing GRPC API server...")
	grpcServer.GracefulStop()

	proxyctx, cancelProxy := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelProxy()
	systemLogger.Info().Msg("Closing Proxy server...")
	proxyServer.Shutdown(proxyctx)
}
