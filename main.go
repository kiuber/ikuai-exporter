package main

import (
	"fmt"
	"github.com/alexflint/go-arg"
	"github.com/jakeslee/ikuai-exporter/ikuai"
	"github.com/jakeslee/ikuai-exporter/pkg"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/client_golang/prometheus/push"
	"github.com/prometheus/common/expfmt"
	"github.com/robfig/cron/v3"
	"log"
	"net/http"
	"os"
	"time"
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

	if config.PushgatewayUrl != "" {
		c := cron.New(cron.WithSeconds(), cron.WithChain(cron.SkipIfStillRunning(cron.DefaultLogger)), cron.WithLogger(
			cron.VerbosePrintfLogger(log.New(os.Stdout, "crontab: ", log.LstdFlags))), cron.WithLocation(time.UTC))

		log.Printf("pushgateway crontab spec: %s", config.PushgatewayCrontab)
		c.AddFunc(config.PushgatewayCrontab, func() {
			log.Printf("push to %s, job: %s", config.PushgatewayUrl, config.PushgatewayJob)

			pusher := push.New(config.PushgatewayUrl, config.PushgatewayJob).
				Collector(exporter).Client(http.DefaultClient).
				BasicAuth(config.PushgatewayUsername, config.PushgatewayPassword).
				Format(expfmt.FmtProtoDelim)
			if err := pusher.Push(); err != nil {
				fmt.Println("could not push completion time to PushGateway: ", err)
			} else {
				log.Printf("push done")
			}
		})
		c.Start()
	}

	log.Printf("exporter %v started at :9090", version)

	log.Fatal(http.ListenAndServe(":9090", nil))
}
