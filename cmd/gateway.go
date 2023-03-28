/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/miRemid/cqless/pkg/cninetwork"
	"github.com/miRemid/cqless/pkg/gateway"
	"github.com/miRemid/cqless/pkg/middleware"
	"github.com/miRemid/cqless/pkg/types"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

const NameExpression = "-a-zA-Z_0-9."

var (
	route  *mux.Router
	config = types.GetConfig()
	cni    *cninetwork.CNIManager
)

func init() {
	route = mux.NewRouter()
	cni = new(cninetwork.CNIManager)
}

// gatewayCmd represents the gateway command
var gatewayCmd = &cobra.Command{
	Use:   "gateway",
	Short: "setup CQLESS gateway for CQLESS provider",
	Long: `gateway is the entry for the CQLESS function container, gateway will proxy 
	the cli invoke or message event from go-cqhttp to the deployed function`,
}

func init() {
	gatewayCmd.AddCommand(&cobra.Command{
		Use:   "up",
		Short: "Run gateway",
		Long:  "Run gateway",
		RunE:  runGateway,
	})
	rootCmd.AddCommand(gatewayCmd)
}

func printEndpoints(r *mux.Router) {
	if err := r.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		methods, err := route.GetMethods()
		if err != nil {
			return err
		}
		path, err := route.GetPathTemplate()
		if err != nil {
			return nil
		}
		log.Debug().Str("path", path).Strs("methods", methods).Send()
		return nil
	}); err != nil {
		log.Error().Err(err).Send()
	}
}

func runGateway(cmd *cobra.Command, args []string) error {

	if err := gateway.Init(config); err != nil {
		return err
	}

	proxyHandler := gateway.MakeProxyHandler(config.Proxy)

	route.Use(middleware.Logger)

	route.HandleFunc("/cqless/function", gateway.MakeDeployHandler(cni, "", false)).Methods(http.MethodPost)
	route.HandleFunc("/cqless/function", gateway.MakeRemoveHandler(cni)).Methods(http.MethodDelete)
	// route.HandleFunc("/cqless/function", gateway.MakeDeployHandler(cni, "", false)).Methods(http.MethodPut)

	route.HandleFunc("/function/{name:["+NameExpression+"]+}", proxyHandler).Methods(http.MethodGet)
	route.HandleFunc("/function/{name:["+NameExpression+"]+}/", proxyHandler).Methods(http.MethodGet)
	route.HandleFunc("/function/{name:["+NameExpression+"]+}/{params:.*}", proxyHandler).Methods(http.MethodGet)

	printEndpoints(route)

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
