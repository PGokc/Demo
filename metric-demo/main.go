package main

import (
	"fmt"
	"math/rand"
	"time"

	m "code.byted.org/gopkg/metrics/v4"
)

var client m.Client
var metric m.Metric

func init() {
	var err error
	client, err = m.NewClient("metrics.sdk.demo")
	if err != nil {
		panic(fmt.Sprintf("failed to create the client: %v\n", err))
	}

	metric, err = client.NewMetricWithOps("my.metric", []string{"tag0", "tag1", "tag2"},
		m.SetHistogramBucket(m.LinearBuckets(1, 2, 3)),
		m.SetMultiFieldTimer(),
	)
	if err != nil {
		fmt.Printf("failed to create the metric: %v\n, the metric is a noop instance", err)
	}
}

func main() {
	ticker := time.NewTicker(3 * time.Second)
	for _ = range ticker.C {
		err := metric.WithTags(
			m.T{"tag0", "value0"},
			m.T{"tag1", "value1"},
			m.T{"tag2", "value2"},
		).Emit(
			m.IncrCounter(rand.Intn(10)), // Counter类型, 默认后缀:counter
			m.Incr(rand.Intn(10)),        // RateCounter类型，默认后缀：rate
			m.Store(rand.Intn(10)),       // Store类型, 默认后缀：store
			m.Observe(rand.Intn(500)),    // Timer类型，默认后缀:timer
			m.Stat(rand.Intn(500)),       // Histogram类型,默认后缀histogram
		)

		if err != nil {
			fmt.Printf("failed to emit metric: %v\n", err)
		}
	}
}
