package main

// https://prometheus.io/docs/guides/go-application/
// https://github.com/brancz/prometheus-example-app/blob/master/main.go

import (
	"crypto/tls"
	"flag"
	"github.com/ncw/swift"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

var (
	addr = flag.String("listen-address", ":8080", "The address to listen on for HTTP requests.")

	swiftAuthUrl        string
	swiftUserName       string
	swiftPassword       string
	swiftTenant         string
	swiftUseInsecureTLS bool

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
func getSwiftAcountInfo(client swift.Connection) error {

	var err error
	var info swift.Account
	var hdr swift.Headers

	info, hdr, err = client.Account()
	if err != nil {
		log.Fatalf("Can't get info from Swift (%s) \n", err.Error())
	} else {
		var currentSwiftAcountUsed float64
		var currentSwiftQuota float64

		currentSwiftAcountUsed = float64(info.BytesUsed)

		swiftAcountUsed.Set(currentSwiftAcountUsed)
		//fmt.Printf("Bytes used: %.0f\n", currentSwiftAcountUsed)

		currentSwiftQuota, err = strconv.ParseFloat(hdr["X-Account-Meta-Quota-Bytes"], 64)
		if err != nil {
			log.Fatalf("Can't parse info from Swift (%s) \n", err.Error())
		} else {
			swiftAcountQuota.Set(currentSwiftQuota)
		}
		//fmt.Printf("Quota: %.0f\n", currentSwiftQuota)
	}
	return err
}

// print error and usage and die
func logFatalfwithUsage(format string, v ...interface{}) {
	log.Printf(format, v...)
	flag.Usage()
	os.Exit(1)
}

func main() {
	flag.StringVar(&swiftAuthUrl, "swift-auth-url", "https://10.200.52.80:5000/v2.0", "The URL for Swift connection")
	flag.StringVar(&swiftUserName, "swift-user-name", "", "The username for swift login")
	flag.StringVar(&swiftPassword, "swift-password", "", "The password for swift login")
	flag.StringVar(&swiftTenant, "swift-tenant", "cc", "The tenant name for swift")
	flag.BoolVar(&swiftUseInsecureTLS, "swift-use-insecure-tls", false, "Use InsecureTLS for Swift communication")
	flag.Parse()

	var lastUpdateTime *time.Time
	var err error

	// test input vars
	_, err = url.ParseRequestURI(swiftAuthUrl)
	if err != nil {
		log.Fatalf("Wrong or empty URL for Swift Endpoint (%s) \n", err.Error())
	}

	if swiftUserName == "" {
		logFatalfwithUsage("Empty username for Swift login!\n")
	}

	if swiftPassword == "" {
		logFatalfwithUsage("Empty password for Swift login!\n")
	}

	// setup tls
	transport := &http.Transport{}
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: swiftUseInsecureTLS}

	// Create a connection
	client := swift.Connection{
		UserName:  swiftUserName,
		ApiKey:    swiftPassword,
		AuthUrl:   swiftAuthUrl,
		Tenant:    swiftTenant,
		Transport: transport,
	}
	// Authenticate
	err = client.Authenticate()
	if err != nil {
		log.Fatalf("Can't authenticate to Swift (%s) \n", err.Error())
	}

	// run main goroutine
	go func() {
		for {
			err = getSwiftAcountInfo(client)
			time.Sleep(3 * time.Second)
		}
	}()

	// export metrics endpoint
	handler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	http.Handle("/metrics", handler)
	log.Printf("exposing metrics on %v/metrics\n", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
