package dotler_test

import (
	"flag"
	"github.com/ronin13/dotler/dotler"
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
		dotler.ClientTimeout = testURL.clientTimeout
		code := dotler.StartCrawl(testURL.turl)
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
		dotler.ClientTimeout = 1
		code := dotler.StartCrawl("http://wnohang.net")
		if code > 0 {
			b.Fatalf("Benchmark failed")
		}
	}
}

func BenchmarkDotlerWithoutGen(b *testing.B) {
	flag.Lookup("alsologtostderr").Value.Set("false")
	flag.Lookup("gen-graph").Value.Set("true")
	dotler.ClientTimeout = 1
	for n := 0; n < b.N; n++ {
		code := dotler.StartCrawl("http://wnohang.net")
		if code > 0 {
			b.Fatalf("Benchmark failed")
		}
	}
}

func TestMain(m *testing.M) {
	os.Remove("dotler.dot")
	dotler.ParseFlags()
	os.Exit(m.Run())
}
