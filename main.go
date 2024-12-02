package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/melbahja/goph"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	host     = flag.String("h", "", "host 默认为空")
	user     = flag.String("u", "root", "user 默认为root")
	password = flag.String("p", "", "password 默认为空")
	gateway  = flag.String("g", "192.168.203.254", "ping的目标")
	httpPort = flag.String("P", ":9100", "port 默认为9100")
)

var (
	pingAvg = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "pingAvgCollectedByGO",
			Help: "pingAvg collected by prometheus go client",
		},
		[]string{"type"},
	)
	pingMax = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "pingMaxCollectedByGO",
			Help: "pingMax collected by prometheus go client",
		},
		[]string{"type"},
	)
)

func init() {
	prometheus.MustRegister(pingAvg)
	prometheus.MustRegister(pingMax)
}

func main() {
	//获取命令行参数
	flag.Parse()

	go func() {
		for {
			//登陆测试机
			res, err := sshTo(*user, *host, *gateway)
			if err != nil {
				log.Panic(err)
			}
			fmt.Println(time.Now())
			fmt.Println(res)
			avg := strings.Split(res, " ")[0]
			max := strings.Split(res, " ")[1]
			avgF, err := strconv.ParseFloat(avg, 64)
			if err != nil {
				log.Panic(err)
			}
			maxF, err := strconv.ParseFloat(max, 64)
			if err != nil {
				log.Panic(err)
			}
			pingAvg.With(prometheus.Labels{"type": "avg"}).Set(avgF)
			pingMax.With(prometheus.Labels{"type": "max"}).Set(maxF)

			if avgF > 80 {
				webhook(avg, max)
			}

			time.Sleep(300 * time.Second)
		}
	}()

	http.Handle("/metrics", promhttp.Handler())
	log.Panic(http.ListenAndServe(*httpPort, nil))

}

func sshTo(user string, host string, gateway string) (string, error) {
	key, _ := goph.Key("/root/.ssh/id_rsa", "")
	client, err := goph.New(user, host, key)
	if err != nil {
		return "", err
	}

	defer client.Close()
	cmd := fmt.Sprintf("ping -c 5 %s | tail -1 | awk -F'/' '{print $5,$6}'", gateway)
	res, err := client.Run(cmd)
	if err != nil {
		return "", err
	}
	resN := strings.Split(string(res), "\n")[0]
	return string(resN), nil
}

func webhook(avg string, max string) {
	cmd := exec.Command("sh", "/root/curlWebhook.sh", *gateway, avg, max)
	err := cmd.Run()
	if err != nil {
		log.Panic(err)
	}
}
