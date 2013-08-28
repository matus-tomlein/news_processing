package main

import (
	"fmt"
	"os/exec"
	"github.com/matus-tomlein/news_processing/environment"
	"net/http"
	"io/ioutil"
    "encoding/json"
	"hash/fnv"
	"time"
	"os"
)

type LinkInfo struct {
	Url string
	PageId int `json:"page_id"`
}

func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

func DownloadUrl(url string, pageId int, envType string) {
	fmt.Println("Downloading", url)
	fileName := environment.DownloadedArticlePath(pageId, hash(url), envType)

	cmd := exec.Command("sh", "-c", "curl -L --max-time 60 \"" + url + "\" > " + fileName)
	err := cmd.Run()
	if err != nil {
		fmt.Printf("%s\n", err)
	}
}

func StartDownload(envType string, messages chan string) {
	for {
		resp, err := http.Get("http://calculon.fiit.stuba.sk:60080/links_to_downloads/take.json")
		if err != nil { panic(err) }
		defer resp.Body.Close()
		content, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
			continue
		}

		links := make([]LinkInfo, 0)
		err = json.Unmarshal(content, &links)
		if err != nil { panic(err) }

		for _, link := range links {
			DownloadUrl(link.Url, link.PageId, envType)
		}

		fmt.Println("Batch downloaded")
		time.Sleep(1 * time.Second)

		select {
			case msg := <-messages:
				if msg == "q" { os.Exit(0) }
			default:
		}
	}
}