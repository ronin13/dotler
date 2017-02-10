// Copyright 2017 Raghavendra Prabhu.
// Refer to LICENSE for more

package dotler

// Responsible for post-processing such as generation of SVG
// and display of image after crawling is done.
import (
	"github.com/golang/glog"

	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

func postProcess() int {

	var err error
	var graphPipe io.WriteCloser

	glog.Infof("Generating svg from dot file")
	cmdLine := strings.Split(fmt.Sprintf("-T%s -o dotler.%s", graphFormat, graphFormat), " ")
	graphIt := exec.Command("dot", cmdLine...)
	graphIt.Stdout = os.Stdout
	graphIt.Stderr = os.Stderr

	graphPipe, err = graphIt.StdinPipe()
	panicCrawl(err)

	panicCrawl(graphIt.Start())

	_, err = graphPipe.Write([]byte(crawlGraph.String()))
	panicCrawl(err)

	panicCrawl(graphPipe.Close())

	err = graphIt.Wait()
	if err != nil {
		glog.Fatalf("dotler.svg generation failed!")
		return 1
	}
	if showProg != "" {
		glog.Infof("Displaying image with %s", showProg)
		destFile := fmt.Sprintf("dotler.%s", graphFormat)
		showIt := exec.Command(showProg, destFile)
		showIt.Stdout = os.Stdout
		showIt.Stderr = os.Stderr
		panicCrawl(showIt.Start())
		err = showIt.Wait()
		if err != nil {
			glog.Fatalf("Display of image with %s failed", showProg)
			return 1
		}
	}
	return 0
}
