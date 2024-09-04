package main

import (
	"github.com/alexflint/go-arg"
	"github.com/jakeslee/ikuai-exporter/ikuai"
	"github.com/jakeslee/ikuai-exporter/pkg"
	mper "github.com/kiuber/metrics-pusher/mper"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
)

type Config struct {
	Ikuai               string `arg:"env:IK_URL" help:"iKuai URL" default:"http://10.0.1.253"`
	IkuaiUsername       string `arg:"env:IK_USER" help:"iKuai username" default:"test"`
	IkuaiPassword       string `arg:"env:IK_PWD" help:"iKuai password" default:"test123"`
	Debug               bool   `arg:"env:DEBUG" help:"iKuai 开启 debug 日志" default:"false"`
	InsecureSkip        bool   `arg:"env:SKIP_TLS_VERIFY" help:"是否跳过 iKuai 证书验证" default:"true"`
	PushgatewayUrl      string `arg:"env:PG_URL" help:"pushgateway url" default:""`
	PushgatewayCrontab  string `arg:"env:PG_CRONTAB" help:"pushgateway crontab, default every minute" default:"*/15 * * * * *"`
	PushgatewayJob      string `arg:"env:PG_JOB" help:"pushgateway job" default:"ikuai"`
	PushgatewayUsername string `arg:"env:PG_USERNAME" help:"pushgateway username" default:""`
	PushgatewayPassword string `arg:"env:PG_PASSWORD" help:"pushgateway password" default:""`
}

var (
	version   string
	buildTime string
)

func main() {
	config := &Config{}
	arg.MustParse(config)

	i := ikuai.NewIKuai(config.Ikuai, config.IkuaiUsername, config.IkuaiPassword, config.InsecureSkip, true)

	if config.Debug {
		i.Debug()
	}

	registry := prometheus.NewRegistry()

	exporter := pkg.NewIKuaiExporter(i)
	registry.MustRegister(exporter)

	http.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{Registry: registry}))

	mper.PullPushCrontab(mper.PullPushConfig{
		MetricsUrl:          "http://localhost:9090/metrics",
		PushgatewayUrl:      config.PushgatewayUrl,
		PushgatewayUsername: config.PushgatewayUsername,
		PushgatewayPassword: config.PushgatewayPassword,
		PushgatewayCrontab:  config.PushgatewayCrontab,
	})

	log.Printf("exporter %v started at :9090", version)
	log.Fatal(http.ListenAndServe(":9090", nil))
}
