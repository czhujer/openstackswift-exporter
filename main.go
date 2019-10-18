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
	addr string

	swiftAuthUrl        string
	swiftUserName       string
	swiftPassword       string
	swiftTenant         string
	swiftUseInsecureTLS bool

	// quota and real usage
	swiftAccountQuotaBytes = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "swift_account_quota_bytes",
			Help: "Quota for OpenStack Swift Account",
		},
	)
	swiftAcountUsedBytes = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "swift_account_used_bytes",
			Help: "Used space by containers",
		},
	)

	registry *prometheus.Registry
)

func init() {
	registry = prometheus.NewRegistry()
	registry.MustRegister(swiftAccountQuotaBytes)
	registry.MustRegister(swiftAcountUsedBytes)
}

// list account info for exporting and add to prometheus registry
func getSwiftAcountInfo(client swift.Connection) error {

	info, hdr, err := client.Account()
	if err != nil {
		log.Printf("Can't get info from Swift (%s) \n", err.Error())
		return err
	}

	currentSwiftAcountUsed := float64(info.BytesUsed)

	swiftAcountUsedBytes.Set(currentSwiftAcountUsed)
	//fmt.Printf("Bytes used: %.0f\n", currentSwiftAcountUsed)

	currentSwiftQuota, err := strconv.ParseFloat(hdr["X-Account-Meta-Quota-Bytes"], 64)
	if err != nil {
		log.Printf("Can't parse info from Swift (%s) \n", err.Error())
		return err
	}

	swiftAccountQuotaBytes.Set(currentSwiftQuota)
	//fmt.Printf("Quota: %.0f\n", currentSwiftQuota)

	return err
}

// check input from cmd args
func checkInputVars() {

	var err error

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

}

// print error and usage and die
func logFatalfwithUsage(format string, v ...interface{}) {
	log.Printf(format, v...)
	flag.Usage()
	os.Exit(1)
}

func main() {
	flag.StringVar(&addr, "listen-address", ":8080", "The address to listen on for HTTP requests.")
	flag.StringVar(&swiftAuthUrl, "swift-auth-url", "https://10.200.52.80:5000/v2.0", "The URL for Swift connection")
	flag.StringVar(&swiftUserName, "swift-user-name", "", "The username for swift login")
	flag.StringVar(&swiftPassword, "swift-password", "", "The password for swift login")
	flag.StringVar(&swiftTenant, "swift-tenant", "cc", "The tenant name for swift")
	flag.BoolVar(&swiftUseInsecureTLS, "swift-use-insecure-tls", false, "Use InsecureTLS for Swift communication")
	flag.Parse()

	var err error

	// test input vars
	checkInputVars()

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
			time.Sleep(30 * time.Second)
		}
	}()

	// export metrics endpoint
	handler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	http.Handle("/metrics", handler)
	log.Printf("exposing metrics on %v/metrics\n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
