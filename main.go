package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/ncw/swift"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	//"strconv"
	//"strings"
	//"time"
)

var (
	addr     = flag.String("listen-address", ":8080", "The address to listen on for HTTP requests.")
	registry *prometheus.Registry
)

func init() {
	registry = prometheus.NewRegistry()
}

func main() {
	flag.Parse()

	// setup insecure tls
	transport := &http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		MaxIdleConnsPerHost: 2048,
	}
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	// Create a connection
	c := swift.Connection{
		UserName:  "",
		ApiKey:    "",
		AuthUrl:   "https://10.200.52.80:5000/v2.0",
		Tenant:    "cc",
		Transport: transport,
	}
	// Authenticate
	err := c.Authenticate()
	if err != nil {
		panic(err)
	}
	// List all the containers
	//containers, err := c.ContainerNames(nil)
	//if err != nil {
	//	panic(err)
	//}
	//for index := range containers {
	//	fmt.Printf("Found container: %s\n", containers[index])
	//}

	// list account info for exporting
	info, hdr, err := c.Account()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Bytes used: %d\n", info.BytesUsed)
	fmt.Printf("Quota: %s\n", hdr["X-Account-Meta-Quota-Bytes"])

	// export metrics endpoint
	handler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	http.Handle("/metrics", handler)
	log.Printf("exposing metrics on %v/metrics\n", *addr)
	//log.Fatal(http.ListenAndServe(*addr, nil))
}
