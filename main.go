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

// list account info for exporting and add to prometheus registry
func getSwiftAcountInfo(client swift.Connection, updatedAfter *time.Time) (error, *time.Time) {

	var err error
	var info swift.Account
	var hdr swift.Headers

	info, hdr, err = client.Account()
	if err != nil {
		log.Println(err)
	} else {
		var currentSwiftAcountUsed float64
		var currentSwiftQuota float64

		currentSwiftAcountUsed = float64(info.BytesUsed)

		// TODO check if BytesUsed value is zero or greater

		swiftAcountUsed.Set(currentSwiftAcountUsed)
		fmt.Printf("Bytes used: %.0f\n", currentSwiftAcountUsed)

		currentSwiftQuota, err = strconv.ParseFloat(hdr["X-Account-Meta-Quota-Bytes"], 64)
		if err != nil {
			panic(err)
		} else {
			swiftAcountQuota.Set(currentSwiftQuota)
		}
		fmt.Printf("Quota: %.0f\n", currentSwiftQuota)
	}
	//TODO fill this vars
	return err, updatedAfter
}

func main() {
	flag.Parse()

	var lastUpdateTime *time.Time

	// setup insecure tls
	transport := &http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		MaxIdleConnsPerHost: 2048,
	}
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	// Create a connection
	client := swift.Connection{
		//TODO rewrite to flags/params
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

	// run main goroutine
	go func() {
		for {
			err, lastUpdateTime = getSwiftAcountInfo(client, lastUpdateTime)

			time.Sleep(3 * time.Second)
		}
	}()

	// export metrics endpoint
	handler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	http.Handle("/metrics", handler)
	log.Printf("exposing metrics on %v/metrics\n", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
