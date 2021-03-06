package server

import (
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/common/expfmt"
)

func TestMetricDisabled(t *testing.T) {
	t.Log("Testing Metrics Enabled")
	LoadConfig("template-push-proxy.json")
	platform := "junk"
	pushType := PUSH_TYPE_MESSAGE
	CfgPP.AndroidPushSettings[0].AndroidApiKey = platform
	CfgPP.EnableMetrics = false
	Start()
	time.Sleep(time.Second * 2)
	defer func() {
		Stop()
		time.Sleep(time.Second * 2)
	}()

	incrementBadRequest()
	incrementNotificationTotal(platform, pushType)
	incrementSuccess(platform, pushType)
	incrementRemoval(platform, pushType, "not registered")
	incrementFailure(platform, pushType, "error")
	observerNotificationResponse(PUSH_NOTIFY_APPLE, 1)
	observerNotificationResponse(PUSH_NOTIFY_ANDROID, 1)
	observeServiceResponse(1)

	resp, err := http.Get("http://localhost:8066/metrics")
	if err != nil {
		t.Fatalf("service should not return an http error")
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("service should return a parsable response")
	}
	if !strings.Contains(string(data), "404 page not found") {
		t.Fatalf("service should return a 404")
	}
}

func TestMetricEnabled(t *testing.T) {
	t.Log("Testing Metrics Enabled")
	LoadConfig("template-push-proxy.json")
	platform := "junk"
	pushType := PUSH_TYPE_MESSAGE
	CfgPP.AndroidPushSettings[0].AndroidApiKey = platform
	CfgPP.EnableMetrics = true
	Start()
	time.Sleep(time.Second * 2)
	defer func() {
		Stop()
		time.Sleep(time.Second * 2)
	}()

	incrementBadRequest()
	incrementNotificationTotal(platform, pushType)
	incrementSuccess(platform, pushType)
	incrementRemoval(platform, pushType, "not registered")
	incrementFailure(platform, pushType, "error")
	observerNotificationResponse(PUSH_NOTIFY_APPLE, 1)
	observerNotificationResponse(PUSH_NOTIFY_ANDROID, 1)
	observeServiceResponse(1)

	resp, err := http.Get("http://localhost:8066/metrics")
	if err != nil {
		t.Fatalf("failed to get metrics endpoint - %s", err.Error())
	}
	defer resp.Body.Close()

	parser := &expfmt.TextParser{}
	metrics, _ := parser.TextToMetricFamilies(resp.Body)

	counters := []string{metricSuccessName, metricFailureName, metricFailureWithReasonName, metricRemovalName, metricBadRequestName, metricNotificationsTotalName}
	for _, cn := range counters {
		if m, ok := metrics[cn]; !ok {
			t.Fatalf("metric not found. name: %s", cn)
		} else {
			val := m.Metric[0].Counter.Value
			result := float64(1)

			if cn == metricFailureName {
				result = float64(2)
			}

			if val == nil {
				t.Fatalf("no metric value. name: %s", cn)
			}
			if *val != result {
				t.Fatalf("metric value does not match. mame: %s, got: %v, expected: %v",
					cn, *val, result)
			}
		}
	}

	histograms := []string{metricAPNSResponseName, metricFCMResponseName, metricServiceResponseName, metricNotificationResponseName}
	for _, hn := range histograms {
		if m, ok := metrics[hn]; !ok {
			t.Fatalf("metric not found. name: %s", hn)
		} else {
			val := m.Metric[0].Histogram.SampleCount
			if val == nil {
				t.Fatalf("no metric value. name: %s", hn)
			}
			if *val != 1 {
				t.Fatalf("metric value does not match. mame: %s, got: %v, expected: %v",
					hn, *val, 1)
			}
		}
	}
}
