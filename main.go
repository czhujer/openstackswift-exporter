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
	swiftUserNameEnv    string
	swiftPasswordEnv    string
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

	currentSwiftAcountUsed float64
	currentSwiftQuota      float64

	registry *prometheus.Registry

	rootHandlerResponse = `<html>
			<head><title>OpenstackSwift Exporter</title></head>
			<body>
			<h1>OpenstackSwift Exporter</h1>
			<p><a href="/metrics">Metrics</a></p>
			</body>
			</html>`
)

func init() {
	registry = prometheus.NewRegistry()
	registry.MustRegister(swiftAccountQuotaBytes)
	registry.MustRegister(swiftAcountUsedBytes)
}

// list account info and extract values for exporting
func collectSwiftAcountInfo(client swift.Connection) error {

	info, hdr, err := client.Account()
	if err != nil {
		log.Printf("Can't get info from Swift: %s \n", err.Error())
		return err
	}

	currentSwiftAcountUsed = float64(info.BytesUsed)

	//fmt.Printf("Bytes used: %.0f\n", currentSwiftAcountUsed)

	currentSwiftQuota, err = strconv.ParseFloat(hdr["X-Account-Meta-Quota-Bytes"], 64)
	if err != nil {
		log.Printf("Can't parse info from Swift: %s \n", err.Error())
		return err
	}
	//fmt.Printf("Quota: %.0f\n", currentSwiftQuota)

	return err
}

// publish metrics
func publishMetrics() {
	swiftAccountQuotaBytes.Set(currentSwiftQuota)
	swiftAcountUsedBytes.Set(currentSwiftAcountUsed)
}

// check input from cmd args
func checkInputVars() {

	var err error

	_, err = url.ParseRequestURI(swiftAuthUrl)
	if err != nil {
		log.Fatalf("Wrong or empty URL for Swift Endpoint: %s \n", err.Error())
	}

	if swiftUserName == "" {
		logFatalfwithUsage("Empty username for Swift login!\n")
	}

	if swiftPassword == "" {
		logFatalfwithUsage("Empty password for Swift login!\n")
	}

}

// parse and fill input vars
func parseInputVars() {
	flag.StringVar(&addr, "listen-address", ":8080", "The address to listen on for HTTP requests.")
	flag.StringVar(&swiftAuthUrl, "swift-auth-url", "https://10.200.52.80:5000/v2.0", "The URL for Swift connection")
	flag.StringVar(&swiftUserName, "swift-user-name", "", "The username for swift login")
	flag.StringVar(&swiftPassword, "swift-password", "", "The password for swift login")
	flag.StringVar(&swiftTenant, "swift-tenant", "cc", "The tenant name for swift")
	flag.BoolVar(&swiftUseInsecureTLS, "swift-use-insecure-tls", false, "Use InsecureTLS for Swift communication")
	flag.Parse()

	swiftUserNameEnv := os.Getenv("SWIFTUSERNAME")
	swiftUserPassEnv := os.Getenv("SWIFTUSERPASS")

	if swiftUserName == "" && swiftUserNameEnv != "" {
		fmt.Println("INFO: ENV userName used")
		swiftUserName = swiftUserNameEnv
	}

	if swiftPassword == "" && swiftUserPassEnv != "" {
		fmt.Println("INFO: ENV userPass used")
		swiftPassword = swiftUserPassEnv
	}
}

// print error and usage and die
func logFatalfwithUsage(format string, v ...interface{}) {
	log.Printf(format, v...)
	flag.Usage()
	os.Exit(1)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(rootHandlerResponse))
}

func main() {
	var err error

	// parse and fill input vars
	parseInputVars()

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
		log.Fatalf("Can't authenticate to Swift: %s \n", err.Error())
	}

	// run main goroutine
	go func() {
		for {
			err = collectSwiftAcountInfo(client)
			publishMetrics()
			time.Sleep(30 * time.Second)
		}
	}()

	// export metrics endpoint
	metricsHandler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	http.Handle("/metrics", metricsHandler)
	http.HandleFunc("/", rootHandler)

	log.Printf("exposing metrics on %v/metrics\n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
