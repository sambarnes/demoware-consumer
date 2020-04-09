package metrics

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
	"time"
)

func TestLoadMetricsHandler_CurrentStats(t *testing.T) {
	handler := LoadMetricsHandler{}
	statsCopy := handler.CurrentStats()
	statsCopy.Max = 10
	newCopy := handler.CurrentStats()
	if newCopy.Max != 0 {
		t.Errorf("internal stats modified by external code")
	}
}

func TestCPUMetricsHandler_CurrentStats(t *testing.T) {
	handler := CPUMetricsHandler{stats: CPUUsageStats{Averages: []float64{0}}}
	statsCopy := handler.CurrentStats()
	statsCopy.Averages[0] = 1
	newCopy := handler.CurrentStats()
	if newCopy.Averages[0] != 0 {
		t.Errorf("internal stats modified by external code")
	}
}

func TestKernelMetricsHandler_CurrentStats(t *testing.T) {
	handler := KernelMetricsHandler{}
	statsCopy := handler.CurrentStats()
	statsCopy.MostRecent = time.Now()
	newCopy := handler.CurrentStats()
	zeroTimestamp := time.Time{}
	if newCopy.MostRecent != zeroTimestamp {
		t.Errorf("internal stats modified by external code")
	}
}

func TestLoadStats_Update(t *testing.T) {
	testCases := []struct {
		Metrics     []float64
		ExpectedMax float64
		ExpectedMin float64
	}{
		{
			Metrics:     []float64{0.1708074, 0.7533491, 0.2730651, 0.2999999, 0.9424849, 0.4783437, 0.0882931, 0.284951, 0.5481229, 0.5111811},
			ExpectedMax: 0.9424849,
			ExpectedMin: 0.0882931,
		},
		{
			Metrics:     []float64{0.0601794, 0.7496037, 0.3492028, 0.2957186, 0.6272771, 0.7897101, 0.2885016, 0.7665914, 0.0222295, 0.2740695},
			ExpectedMax: 0.7897101,
			ExpectedMin: 0.0222295,
		},
		{
			Metrics:     []float64{0.8959552, 0.2371776, 0.7706957, 0.6617887, 0.9129505, 0.7265538, 0.6759594, 0.2327206, 0.7030355, 0.9101484},
			ExpectedMax: 0.9129505,
			ExpectedMin: 0.2327206,
		},
		{
			Metrics:     []float64{0.3060088, 0.6598068, 0.1945246, 0.9673805, 0.4790032, 0.0866566, 0.6738122, 0.4391651, 0.1479913, 0.3696593},
			ExpectedMax: 0.9673805,
			ExpectedMin: 0.0866566,
		},
		{
			Metrics:     []float64{0.3242237, 0.369516, 0.6909641, 0.8996092, 0.2083473, 0.091752, 0.8703272, 0.4545293, 0.7116505, 0.2226726},
			ExpectedMax: 0.8996092,
			ExpectedMin: 0.091752,
		},
	}
	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("ValidCase%v", i), func(t *testing.T) {
			stats := LoadStats{}
			for _, m := range testCase.Metrics {
				err := stats.Update(m)
				if err != nil {
					t.Fatalf("unexpected error in stats.Update(): %v", err)
				}
			}
			if stats.n != len(testCase.Metrics) {
				t.Errorf("unexpected stats.n: %v != %v (observed, expected)", stats.n, len(testCase.Metrics))
			}
			if stats.Max != testCase.ExpectedMax {
				t.Errorf("unexpected stats.Max: %v != %v (observed, expected)", stats.Max, testCase.ExpectedMax)
			}
			if stats.Min != testCase.ExpectedMin {
				t.Errorf("unexpected stats.Min: %v != %v (observed, expected)", stats.Min, testCase.ExpectedMin)
			}
		})
	}
}

func TestCPUUsageStats_Update(t *testing.T) {
	type testCase struct {
		N                int
		CPUCount         int
		Metrics          [][]float64
		ExpectedTotals   []float64
		ExpectedAverages []float64
	}
	s := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(s)
	testCases := make([]testCase, 10)
	for caseNumber, testCase := range testCases {
		// Randomly generate happy-path test cases
		testCase.N = 10 + rng.Intn(40)
		testCase.CPUCount = 1 + rng.Intn(5)
		testCase.Metrics = make([][]float64, testCase.N)
		testCase.ExpectedTotals = make([]float64, testCase.CPUCount)
		for i := 0; i < testCase.N; i++ {
			testCase.Metrics[i] = make([]float64, testCase.CPUCount)
			for j := 0; j < testCase.CPUCount; j++ {
				newMetric := rng.Float64()
				testCase.Metrics[i][j] = newMetric
				testCase.ExpectedTotals[j] += newMetric
			}
			testCase.ExpectedAverages = make([]float64, testCase.CPUCount)
		}
		for i := 0; i < testCase.CPUCount; i++ {
			testCase.ExpectedAverages[i] = testCase.ExpectedTotals[i] / float64(testCase.N)
		}

		// Then run them
		t.Run(fmt.Sprintf("ValidCase%v", caseNumber), func(t *testing.T) {
			stats := CPUUsageStats{}
			for _, metric := range testCase.Metrics {
				err := stats.Update(metric)
				if err != nil {
					t.Fatalf("unexpected error in stats.Update(): %v", err)
				}
			}
			if stats.n != len(testCase.Metrics) {
				t.Errorf("unexpected stats.n: %v != %v (observed, expected)", stats.n, len(testCase.Metrics))
			}
			if !reflect.DeepEqual(stats.totals, testCase.ExpectedTotals) {
				t.Errorf("unexpected stats.totals: %v != %v (observed, expected)", stats.totals, testCase.ExpectedTotals)
			}
			if !reflect.DeepEqual(stats.Averages, testCase.ExpectedAverages) {
				t.Errorf("unexpected stats.Averages: %v != %v (observed, expected)", stats.Averages, testCase.ExpectedAverages)
			}
		})
	}

	// Special bad cases
	t.Run("MismatchArrayLengths", func(t *testing.T) {
		stats := CPUUsageStats{}
		goodFirstMetric := make([]float64, 4)
		badSecondMetric := make([]float64, 1)
		err := stats.Update(goodFirstMetric)
		if err != nil {
			t.Fatalf("unexpected error in stats.Update(): %v", err)
		}
		err = stats.Update(badSecondMetric)
		if err == nil {
			t.Fatal("expected error in stats.Update(), got none")
		}
		if stats.n != 1 {
			t.Errorf("unexpected stats.n: should not increment when an error is raised in stats.Update")
		}
	})
}

func TestKernelUpgradeStats_Update(t *testing.T) {
	testCases := []struct {
		Metrics                   []string
		ExpectedMostRecentUpgrade string
	}{
		{
			Metrics: []string{
				"2020-04-02T11:38:29.886438475-05:00",
				"2020-04-02T11:38:54.887244112-05:00",
				"2020-04-02T11:38:39.886793347-05:00",
				"2020-04-02T11:39:04.887691248-05:00",
				"2020-04-02T11:38:19.886101098-05:00",
				"2020-04-02T11:38:44.886963028-05:00",
				"2020-04-02T11:38:49.887072248-05:00",
				"2020-04-02T11:38:34.886607459-05:00",
				"2020-04-02T11:38:59.887454161-05:00",
				"2020-04-02T11:38:24.886237657-05:00",
			},
			ExpectedMostRecentUpgrade: "2020-04-02T11:39:04.887691248-05:00",
		},
		// TODO: generate some more test cases
	}
	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("ValidCase%v", i), func(t *testing.T) {
			stats := KernelUpgradeStats{}
			for _, metric := range testCase.Metrics {
				err := stats.Update(metric)
				if err != nil {
					t.Fatalf("unexpected error in stats.Update(): %v", err)
				}
			}
			expectedMostRecentUpgrade, err := time.Parse(time.RFC3339, testCase.ExpectedMostRecentUpgrade)
			if err != nil {
				t.Fatalf("unexpected error in parsing testCase.ExpectedMostRecentUpgrade: %v", err)
			}
			if stats.MostRecent != expectedMostRecentUpgrade {
				t.Errorf("unexpected stats.MostRecent: %v != %v (observed, expected)", stats.MostRecent, expectedMostRecentUpgrade)
			}
		})
	}

	// Special bad cases
	t.Run("BadTimestamp", func(t *testing.T) {
		stats := KernelUpgradeStats{}
		err := stats.Update("NO. BAD TIMESTAMP. BAD.")
		if err == nil {
			t.Fatal("expected error in stats.Update(), got none")
		}
		if stats.n != 0 {
			t.Errorf("unexpected stats.n: should not increment when an error is raised in stats.Update")
		}
	})
}
