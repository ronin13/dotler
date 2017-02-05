package main

import (
	"flag"
	"os"
	"testing"
)

var urlTests = []struct {
	turl          string
	clientTimeout uint
}{
	{"http://wnohang.net", 10},
	{"https://gobyexample.com/", 20},
	{"https://blog.golang.org", 20},
	{"https://blog.wnohang.net", 20},
}

func TestDotler(t *testing.T) {
	flag.Lookup("alsologtostderr").Value.Set("false")
	for _, testURL := range urlTests {
		clientTimeout = testURL.clientTimeout
		code := startCrawl(testURL.turl)
		if code != 0 {
			t.Fatalf("Testing failed on %s", testURL.turl)
		}
		if _, err := os.Stat("dotler.dot"); os.IsNotExist(err) {
			t.Fatalf("Test failed on %s: result file does not exist", testURL.turl)
		}
		if os.Remove("dotler.dot") != nil {
			t.Fatalf("Test failed: failed to remove file")
		}
	}
}

func BenchmarkDotler(b *testing.B) {
	flag.Lookup("alsologtostderr").Value.Set("false")
	for n := 0; n < b.N; n++ {
		clientTimeout = 1
		code := startCrawl("http://wnohang.net")
		if code > 0 {
			b.Fatalf("Benchmark failed")
		}
	}
}

func BenchmarkDotlerWithoutGen(b *testing.B) {
	flag.Lookup("alsologtostderr").Value.Set("false")
	clientTimeout = 1
	genGraph = false
	for n := 0; n < b.N; n++ {
		code := startCrawl("http://wnohang.net")
		if code > 0 {
			b.Fatalf("Benchmark failed")
		}
	}
}

func TestMain(m *testing.M) {
	os.Remove("dotler.dot")
	parseFlags()
	os.Exit(m.Run())
}
