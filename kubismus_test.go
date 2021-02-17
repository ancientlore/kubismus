package kubismus

import (
	"net/http"
	"net/http/httptest"
	"runtime"
	"strconv"
	"testing"
	"time"
)

func TestMetrics(t *testing.T) {
	ops := []Op{COUNT, AVERAGE, SUM}

	for _, op := range ops {
		Define("testMetrics"+op.String(), op, "A test "+op.String())
	}

	// need channels processed
	runtime.Gosched()

	for _, op := range ops {
		m := GetMetrics("testMetrics"+op.String(), op)
		if m == nil {
			t.Fatal(op.String(), ": Metrics is nil")
		}
		if len(m) != cMETRICS {
			t.Errorf("%s Metrics is invalid size: %d != %d", op.String(), len(m), cMETRICS)
		}
		for i, x := range m {
			if x != 0 {
				t.Errorf("%s: There should be no reading at position %d: %g", op.String(), i, x)
			}
		}
		ReleaseMetrics(m)
	}

	for _, op := range ops {
		Metric("testMetrics"+op.String(), 1, 3.14)
	}

	// metrics aggregates each second
	time.Sleep(1100 * time.Millisecond)

	for _, op := range ops {
		m := GetMetrics("testMetrics"+op.String(), op)
		if m == nil {
			t.Fatal(op.String(), ": Metrics is nil")
		}
		if len(m) != cMETRICS {
			t.Errorf("%s Metrics is invalid size: %d != %d", op.String(), len(m), cMETRICS)
		}
		for i, x := range m {
			if i < cMETRICS-1 {
				if x != 0 {
					t.Errorf("%s: There should be no reading at position %d: %g", op.String(), i, x)
				}
			} else {
				if x == 0 {
					t.Errorf("%s: There should be a reading at position %d: %g", op.String(), i, x)
				}
			}
		}
		ReleaseMetrics(m)
	}
}

func TestNote(t *testing.T) {
	n := GetNotes()
	if len(n) > 0 {
		t.Errorf("There shoule be no notes: %d", len(n))
	}
	ReleaseNotes(n)

	Note("OK", "This is a test note")
	runtime.Gosched()

	n = GetNotes()
	if len(n) != 1 {
		t.Errorf("There shoule be ONE note: %d", len(n))
	} else {
		if n[0].Name != "OK" {
			t.Errorf("Got the wrong note: %s", n[0].Name)
		}
	}
	ReleaseNotes(n)
}

func parallelReader(name string, iterations int, wait, done chan struct{}) {
	<-wait
	for i := 0; i < iterations; i++ {
		m := GetMetrics(name, AVERAGE)
		if m != nil {
			ReleaseMetrics(m)
		}
	}
	done <- struct{}{}

}

func parallelWriter(name string, iterations int, wait, done chan struct{}) {
	<-wait
	for i := 0; i < iterations; i++ {
		Metric(name, 1, 3.14159)
	}
	done <- struct{}{}
}

func benchmarkMetrics(b *testing.B, numReaders, numWriters, iterations int) {

	b.StopTimer()
	done := make(chan struct{})
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		wait := make(chan struct{})

		for i := 0; i < numReaders; i++ {
			go parallelReader("reading"+strconv.Itoa(i), iterations, wait, done)
		}

		for i := 0; i < numWriters; i++ {
			go parallelWriter("reading"+strconv.Itoa(i), iterations, wait, done)
		}

		close(wait)

		for i := 0; i < numReaders+numWriters; i++ {
			<-done
		}
	}
}

// 1 reader, 1 writer, 32 iteratons each = 64 operations
func BenchmarkMetricsSameReadWrite1(b *testing.B) {
	benchmarkMetrics(b, 1, 1, 32)
}

// 2 readers, 2 writers, 32 iterations each = 128 operations
func BenchmarkMetricsSameReadWrite2(b *testing.B) {
	benchmarkMetrics(b, 2, 2, 32)
}

// 4 readers, 4 writers, 32 iterations each = 256 operations
func BenchmarkMetricsSameReadWrite4(b *testing.B) {
	benchmarkMetrics(b, 4, 4, 32)
}

// 2 readers, 8 writers, 32 iterations each = 320 operations
func BenchmarkMetrics1(b *testing.B) {
	benchmarkMetrics(b, 2, 8, 32)
}

// 4 readers, 16 writers, 64 iterations each = 1280 operations
func BenchmarkMetrics2(b *testing.B) {
	benchmarkMetrics(b, 4, 16, 64)
}

// 1 reader, 64 writers, 128 iterations each = 8320 operations
func BenchmarkMetrics3(b *testing.B) {
	benchmarkMetrics(b, 1, 64, 128)
}

// 8 readers, 320 writers, 256 iterations each = 83968 operations
func BenchmarkMetrics4(b *testing.B) {
	benchmarkMetrics(b, 8, 320, 256)
}

// 16 readers, 2048 writers, 64 iterations each = 132096 operations
func BenchmarkMetrics5(b *testing.B) {
	benchmarkMetrics(b, 16, 2048, 64)
}

// 16 readers, 4096 writers, 512 iterations each = 2105344 operations
func BenchmarkMetrics6(b *testing.B) {
	benchmarkMetrics(b, 16, 4096, 512)
}

func TestWebSite(t *testing.T) {
	req := httptest.NewRequest("GET", "/web/d3.min.js", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Error("Unable to find d3.min.js")
	}
}
