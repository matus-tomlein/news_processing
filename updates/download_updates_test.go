package main

import (
	"testing"
	"github.com/matus-tomlein/news_processing/page_db"
	"encoding/json"
)

func TestDownloadedJson(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	out, err := downloadJson(1, "http://fiit.sk", "test")
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	links := make([]page_db.UpdateLink, 0)
	err = json.Unmarshal(out, &links)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	if len(links) == 0 {
		t.Log("No links downloaded")
		t.FailNow()
	}
}

func TestDownlaodJsonForBadPage(t *testing.T) {
	_, err := downloadJson(12, "http://today.gm", "test")
	if err == nil {
		t.FailNow()
	}
}
