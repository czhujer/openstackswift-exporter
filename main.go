package main

// https://prometheus.io/docs/guides/go-application/
// https://github.com/brancz/prometheus-example-app/blob/master/main.go

import (
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/ncw/swift"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"strconv"
	"time"
)

var (
	addr = flag.String("listen-address", ":8080", "The address to listen on for HTTP requests.")

	// quota and real usage
	swiftAcountQuota = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "swift_account_quota",
			Help: "Quota for OpenStack Swift Account",
		},
	)
	swiftAcountUsed = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "swift_account_used",
			Help: "Used space by containers",
		},
	)

	registry *prometheus.Registry
)

func init() {
	registry = prometheus.NewRegistry()
	registry.MustRegister(swiftAcountQuota)
	registry.MustRegister(swiftAcountUsed)
}

//func ScrapeSwiftAcount(client *client, updatedAfter *time.Time) (error, *time.Time){
//
//}

func main() {
	flag.Parse()

	//var lastUpdateTime *time.Time

	// setup insecure tls
	transport := &http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		MaxIdleConnsPerHost: 2048,
	}
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	// Create a connection
	client := swift.Connection{
		UserName:  "",
		ApiKey:    "",
		AuthUrl:   "https://10.200.52.80:5000/v2.0",
		Tenant:    "cc",
		Transport: transport,
	}
	// Authenticate
	err := client.Authenticate()
	if err != nil {
		panic(err)
	}
	// List all the containers
	//containers, err := client.ContainerNames(nil)
	//if err != nil {
	//	panic(err)
	//}
	//for index := range containers {
	//	fmt.Printf("Found container: %s\n", containers[index])
	//}

	// list account info for exporting
	//info, hdr, err := client.Account()
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Printf("Bytes used: %d\n", info.BytesUsed)
	//fmt.Printf("Quota: %s\n", hdr["X-Account-Meta-Quota-Bytes"])

	go func() {
		//var err error
		for {
			//err, lastUpdateTime = ScrapeSwiftAcount(client)

			// list account info for exporting
			info, hdr, err := client.Account()
			if err != nil {
				log.Println(err)
			} else {
				var currentSwiftAcountUsed float64
				var currentSwiftQuota float64

				currentSwiftAcountUsed = float64(info.BytesUsed)
				swiftAcountUsed.Set(currentSwiftAcountUsed)
				fmt.Printf("Bytes used: %.0f\n", currentSwiftAcountUsed)

				var err error
				currentSwiftQuota, err = strconv.ParseFloat(hdr["X-Account-Meta-Quota-Bytes"], 64)
				if err != nil {
					//fmt.Println(currentSwiftQuota)
				} else {
					swiftAcountQuota.Set(currentSwiftQuota)
				}
				fmt.Printf("Quota: %.0f\n", currentSwiftQuota)
			}

			time.Sleep(3 * time.Second)
		}
	}()

	//fmt.Printf("last update time: %s \n", lastUpdateTime)

	// export metrics endpoint
	handler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	http.Handle("/metrics", handler)
	log.Printf("exposing metrics on %v/metrics\n", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
