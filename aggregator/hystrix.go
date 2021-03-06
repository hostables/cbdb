package main

import (
	"strconv"
	"time"
)

// A snapshot transcription of the hystrix.stream JSON object
// This is here for legacy support only. Only update if the fields change or
// In the event of an inevitable bug.
type HystrixStream struct {
	// Forgive my ridiculous formatting in this ridiculous object
	CurrentConcurrentExecutionCount int64  `json:"currentConcurrentExecutionCount,int64"`
	CurrentTime          string            `json:"currentTime,string"`
	ErrorPercentage      int64             `json:"errorPercentage,int64"`
	ErrorCount           int64             `json:"errorCount,int64"`
	Group                string            `json:"group,string"`
	IsCircuitBreakerOpen bool              `json:"isCircuitBreakerOpen,bool"`
	LatencyExecute       HystrixHistogram  `json:"latencyExecute,HystrixHistogram"`
	LatencyExecuteMean   int64             `json:"latencyExecute_mean,int64"`
	LatencyTotal         HystrixHistogram  `json:"latencyTotal,HystrixHistogram"`
	LatencyTotalMean     int64             `json:"latencyTotal_mean,int64"`
	Name                 string            `json:"name,string"`
	ReportingHosts       int64             `json:"reportingHosts,int64"`
	RequestCount         int64             `json:"requestCount,int64"`
	RollingCountCollapsedRequests   int64  `json:"rollingCountCollapsedRequests,int64"`
	RollingCountExceptionsThrown    int64  `json:"rollingCountExceptionsThrown,int64"`
	RollingCountFailure             int64  `json:"rollingCountFailure,int64"`
	RollingCountFallbackFailure     int64  `json:"rollingCountFallbackFailure,int64"`
	RollingCountFallbackRejection   int64  `json:"rollingCountFallbackRejection,int64"`
	RollingCountResponseFromCache   int64  `json:"rollingCountResponseFromCache,int64"`
	RollingCountSemaphoreRejected   int64  `json:"rollingCountSemaphoreRejected,int64"`
	RollingCountShortCircuited      int64  `json:"rollingCountShortCircuited,int64"`
	RollingCountSuccess             int64  `json:"rollingCountSuccess,int64"`
	RollingCountThreadPoolRejected  int64  `json:"rollingCountThreadPoolRejected,int64"`
	RollingCountTimeout             int64  `json:"rollingCOuntTimeout,int64"`
	Type                            string `json:"type,string"`
	// Don't blame me for these awful names.
	// I'm preserving the bad names Hystrix uses
	PropertyValueCircuitBreakerEnabled                            bool   `json:"propertyValue_circuitBreakerEnabled,bool"`
	PropertyValueCircuitBreakerErrorThresholdPercentage           int64  `json:"propertyValue_circuitBreakerErrorThresholdPercentage,int64"`
	PropertyValueCircuitBreakerForceOpen                          bool   `json:"propertyValue_circuitBreakerForceOpen,bool"`
	PropertyValueCircuitBreakerForceClosed                        bool   `json:"propertyValue_circuitBreakerForceClosed,bool"`
	PropertyValueCircuitBreakerRequestVolumeThreshold             int64  `json:"propertyValue_circuitBreakerRequestVolumeThreshold,int64"`
	PropertyValueCircuitBreakerSleepWindowInMilliseconds          int64  `json:"propertyValue_circuitBreakerSleepWindowInMilliseconds,int64"`
	PropertyValueExecutionIsolationSemaphoreMaxConcurrentRequests int64  `json:"propertyValue_executionIsolationSemaphoreMaxConcurrentRequests,int64"`
	PropertyValueExecutionIsolationStrategy                       string `json:"propertyValue_executionIsolationStrategy,string"`
	PropertyValueExecutionIsolationThreadPoolKeyOverride          string `json:"propertyValue_executionIsolationThreadPoolKeyOverride,string"`
	PropertyValueExecutionIsolationThreadTimeoutInMilliseconds    int64  `json:"propertyValue_executionIsolationThreadTimeoutInMilliseconds,string"`
	PropertyValueFallbackIsolationSemaphoreMaxConcurrentRequests  int64  `json:"propertyValue_fallbackIsolationSeampahoreMaxConcurrentRequests,int64"`
	PropertyValueMetricsRollingStatisticalWindowInMilliseconds    int64  `json:"propertyValue_metricsRollingStatisticalWindowInMilliseconds,int64"`
	PropertyValueRequestCacheEnabled                              bool   `json:"propertyValue_requestCacheEnabled,bool"`
	PropertyValueRequestLogEnabled                                bool   `json:"propertyValue_requestLogEnabled,bool"`
}

// A snapshot transcription of the histogram objects hystrix.stream JSON object
// This is here for legacy support only. Only update if the fields change or
// In the event of an inevitable bug.
type HystrixHistogram struct {
	//minimum
	Percentile0   int64 `json:"0,int64"`
	Percentile25  int64 `json:"25,int64"`
	//median
	Percentile50  int64 `json:"50,int64"`
	Percentile75  int64 `json:"75,int64"`
	Percentile90  int64 `json:"90,int64"`
	Percentile95  int64 `json:"95,int64"`
	Percentile99  int64 `json:"99,int64"`
	Percentile995 int64 `json:"99.5,int64"`
	//maximum
	Percentile100 int64 `json:"100,int64"`
}


func (h HystrixHistogram) ToLatencyHistogram(mean int64) LatencyHistogram {
	return LatencyHistogram {
		Mean: mean,
		Median: h.Percentile50,
		Min: h.Percentile0,
		Max: h.Percentile100,
		Percentile25: h.Percentile25,
		Percentile75: h.Percentile75,
		Percentile90: h.Percentile90,
		Percentile95: h.Percentile95,
		Percentile99: h.Percentile99,
		Percentile995: h.Percentile995,
		// FIXME: there's actually a way to calculate this from EWMA
		// Unfortunately, the closest we have is an estimate between 99.5 and 100. We'll take it
		Percentile999: (h.Percentile100 + h.Percentile995) / 2,
	}
}

func (h HystrixStream) ToCircuitBreaker() (CircuitBreaker, error) {
	var breakerCount BreakerCount
	if h.IsCircuitBreakerOpen {
		breakerCount = BreakerCount{OpenCount: 1, ClosedCount: 0}
	} else {
		breakerCount = BreakerCount{OpenCount: 0, ClosedCount: 1}
	}

	var currentTime time.Time

	// This is how I parse the time.
	parsedTime, err := strconv.Atoi(h.CurrentTime)
	if err != nil {
		return CircuitBreaker{}, err
	} else {
		// Split hystrix ms encoded unix time into s and ns
		currentTime = time.Unix(int64(parsedTime / 1000), int64((parsedTime % 1000) * 1000))
	}

	return CircuitBreaker {
		Name: h.Group + h.Name,
		SuccessCount: h.RollingCountSuccess,
		FailCount: 1,
		FallbackCount: 1,
		ShortCircuitCount: 1,
		WindowDuration: 1,
		CurrentTime: currentTime,
		BreakerStatus: breakerCount,
		Latency: h.LatencyTotal.ToLatencyHistogram(h.LatencyTotalMean),
	}, nil
}


