// Copyright 2017 Raghavendra Prabhu.
// Refer to LICENSE for more

// Package dotler http related functions.
package dotler

import (
	"github.com/golang/glog"

	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"time"
)

// Returns content from a url.
// Uses a timeout on http.Client.
// Does not panic, crawling can fail for some pages, doesn't
// mean we throw crawler with bath water. (to use the pun).
func getContent(url *url.URL) (string, error) {
	client := &http.Client{
		Timeout: time.Duration(ClientTimeout) * time.Second,
	}
	resp, err := client.Get(url.String())
	if err != nil {
		glog.Infof("Failed to fetch due to %+v", err)
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		// Can happy, don't panic here, try crawling others
		glog.Infof("Failed to read response %+v", err)
		return "", err
	}
	return string(body), nil
}

// What we consider as a static asset
func isStatic(url string) bool {
	// https://developer.mozilla.org/en-US/docs/Web/HTML/Supported_media_formats
	// https://developer.mozilla.org/en-US/docs/Web/HTML/Element/img#Supported_image_formats
	static := regexp.MustCompile(STATICTYPES)
	return static.MatchString(url)

}
