package main

import (
	"testing"
	"github.com/matus-tomlein/news_processing/page_db"
	"github.com/matus-tomlein/news_processing/environment"
	"encoding/json"
	"os"
	"fmt"
)

func TestDownloadedJson(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	out, err := downloadJson(1, "http://fiit.sk", environment.CurrentTest())
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

func TestDownloadJsonForBadPage(t *testing.T) {
	_, err := downloadJson(12, "http://today.gm", environment.CurrentTest())
	if err == nil {
		t.FailNow()
	}
}

func TestSaveDownloadedJson(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	cacheFolderName, err := download(13, "http://fiit.sk", environment.CurrentTest())
	if err != nil {
		t.Log("Failed to download update", err)
		t.FailNow()
	}
	if _, err := os.Stat(fmt.Sprintf("%s/parsed/%s.json", environment.CachePath(environment.CurrentTest()), cacheFolderName)); os.IsNotExist(err) {
		t.Log("File not found")
		t.FailNow()
	} else {
		err = os.Remove(fmt.Sprintf("%s/parsed/%s.json", environment.CachePath(environment.CurrentTest()), cacheFolderName))
		if err != nil {
			panic(err)
		}
	}
}
