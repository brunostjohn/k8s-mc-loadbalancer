package cmd

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/brunostjohn/k8s-mc-loadbalancer/pkg/loadbalancer"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the load balancer",
	Long:  "Start the load balancer",
	Run:   runStartCommand,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		v := viper.New()
		v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
		v.AutomaticEnv()
		bindFlags(cmd, v)

		return nil
	},
}

var bindAddress net.IP
var bindPort int
var proxyProtocolReceive bool
var proxyProtocolTrustedCIDRs []string
var keepAliveTimeout int
var filtersRateLimiterRequestLimit int
var filtersRateLimiterWindowLength int
var kubeconfig string
var proxyPollInterval int
var prometheusBindAddress net.IP
var prometheusBindPort int
var enablePrometheus bool

func init() {
	startCmd.Flags().IPVarP(&bindAddress, "bind-address", "b", net.IPv4(0, 0, 0, 0), "The IP address to bind to.")
	startCmd.Flags().IntVarP(&bindPort, "bind-port", "p", 25565, "The port to bind to.")
	startCmd.Flags().BoolVarP(&proxyProtocolReceive, "proxy-protocol.receive", "r", false, "Whether to receive PROXY protocol packets.")
	startCmd.Flags().StringSliceVarP(&proxyProtocolTrustedCIDRs, "proxy-protocol.trusted-cidrs", "t", []string{}, "The CIDRs to trust for PROXY protocol.")
	startCmd.Flags().IntVarP(&keepAliveTimeout, "keep-alive-timeout", "k", 0, "The keep-alive timeout in seconds.")
	startCmd.Flags().IntVarP(&filtersRateLimiterRequestLimit, "filters.rate-limiter.request-limit", "l", 0, "Rate limiter request limit.")
	startCmd.Flags().IntVarP(&filtersRateLimiterWindowLength, "filters.rate-limiter.window-length", "w", 0, "Rate limiter window length.")
	startCmd.Flags().StringVarP(&kubeconfig, "kubeconfig", "c", "", "Path to the kubeconfig file.")
	startCmd.Flags().IntVarP(&proxyPollInterval, "proxy-poll-interval", "i", 120, "The interval in seconds to poll for new proxies.")
	startCmd.Flags().IPVarP(&prometheusBindAddress, "prometheus-bind-address", "a", net.IPv4(0, 0, 0, 0), "The IP address to bind to for Prometheus.")
	startCmd.Flags().IntVarP(&prometheusBindPort, "prometheus-bind-port", "o", 2112, "The port to bind to for Prometheus.")
	startCmd.Flags().BoolVarP(&enablePrometheus, "enable-prometheus", "e", false, "Enable Prometheus metrics.")
}

func runStartCommand(cmd *cobra.Command, args []string) {
	lb, err := loadbalancer.NewMCLoadBalancer(cmd.Context(), loadbalancer.MCLoadBalancerOptions{
		Kubeconfig: kubeconfig,
		Log: &log.Logger,
		WatchFrequency: time.Duration(proxyPollInterval) * time.Second,
		BindAddress: &bindAddress,
		BindPort: bindPort,
		ProxyProtocolReceive: proxyProtocolReceive,
		ProxyProtocolTrustedCIDRs: proxyProtocolTrustedCIDRs,
		KeepAliveTimeout: keepAliveTimeout,
		FiltersRateLimiterRequestLimit: filtersRateLimiterRequestLimit,
		FiltersRateLimiterWindowLength: filtersRateLimiterWindowLength,
		EnablePrometheus: enablePrometheus,
		PrometheusBindAddress: &prometheusBindAddress,
		PrometheusBindPort: prometheusBindPort,
	})
	if err != nil {
		panic(err)
	}

	err = lb.Start(cmd.Context())
	if err != nil {
		panic(err)
	}
}

func bindFlags(cmd *cobra.Command, v *viper.Viper) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		configName := f.Name
		
		if !f.Changed && v.IsSet(configName) {
			val := v.Get(configName)
			cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
		}
	})
}