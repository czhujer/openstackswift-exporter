package main

import (
	"flag"
	"net/http"
	"net/http/httptest"
	"testing"
)

// https://golang.org/pkg/net/http/httptest/#pkg-examples
// https://blog.questionable.services/article/testing-http-handlers-go/
// https://quii.gitbook.io/learn-go-with-tests/build-an-application/http-server
// https://medium.com/@hau12a1/golang-capturing-log-println-and-fmt-println-output-770209c791b4

//func TestName(t *testing.T) {
//	name := t.Name()
//	fmt.Println(name)
//}

//func TestMain(m *testing.M) {
//	// call flag.Parse() here if TestMain uses flags
//	os.Exit(m.Run())
//}

func TestWithoutParams(t *testing.T) {
	flag.Parse()
	parsed := flag.Parsed()

	if parsed == false {
		t.Errorf("Parsing flags failed")
	}
}

func TestCheckInputVars(t *testing.T) {
	//checkInputVars()
	//
	//_, err := url.ParseRequestURI(swiftAuthUrl)
	//if err != nil {
	//	t.Log("ParseRequestURI failed")
	//}
}

//func TestCollectSwiftAcountInfo(t *testing.T) {
//	_, _, err := collectSwiftAcountInfo()
//}

func TestRootHandler(t *testing.T) {
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(rootHandler)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check the response body is what we expect.
	expected := `<html>
			<head><title>OpenstackSwift Exporter</title></head>
			<body>
			<h1>OpenstackSwift Exporter</h1>
			<p><a href="/metrics">Metrics</a></p>
			</body>
			</html>`

	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}

//func TestMetricsHandler(t *testing.T) {1
//	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
//	// pass 'nil' as the third parameter.
//	req, err := http.NewRequest("GET", "/metrics", nil)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
//	rr := httptest.NewRecorder()
//	handler := http.Handler(metricsHandler)
//
//	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
//	// directly and pass in our Request and ResponseRecorder.
//	handler.ServeHTTP(rr, req)
//
//	// Check the status code is what we expect.
//	if status := rr.Code; status != http.StatusOK {
//		t.Errorf("handler returned wrong status code: got %v want %v",
//			status, http.StatusOK)
//	}
//
//	// Check the response body is what we expect.
//	expected := `<html>
//			<head><title>OpenstackSwift Exporter</title></head>
//			<body>
//			<h1>OpenstackSwift Exporter</h1>
//			<p><a href="/metrics">Metrics</a></p>
//			</body>
//			</html>`
//
//	if rr.Body.String() != expected {
//		t.Errorf("handler returned unexpected body: got %v want %v",
//			rr.Body.String(), expected)
//	}
//}
