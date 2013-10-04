package web_server

import (
    "fmt"
    "strconv"
    "net/http"
    "net/url"
    "os"
    "io/ioutil"
	"github.com/matus-tomlein/news_processing/environment"
	"github.com/matus-tomlein/news_processing/article_downloader"
)

func handler(w http.ResponseWriter, r *http.Request) {
	values, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		fmt.Println(err)
		return
	}

	if _, ok := values["url"]; !ok {
		fmt.Println("Wrong arguments")
		return
	}
	if _, ok := values["page_id"]; !ok {
		fmt.Println("Wrong arguments")
		return
	}

	url := values["url"][0]
	pageId, err := strconv.Atoi(values["page_id"][0])
    if err != nil {
		fmt.Println(err)
		return
	}
	filename := environment.DownloadedArticlePath(pageId, article_downloader.Hash(url), environment.Current())

	// Download if isnt downloaded yet
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		article_downloader.DownloadUrl(url, pageId, environment.Current())
	}

	// Read from disk
	b, err := ioutil.ReadFile(filename)
    if err != nil {
		fmt.Println(err)
		return
	}
	w.Write(b)
}

func Serve() {
    http.HandleFunc("/article", handler)
    http.ListenAndServe(":8080", nil)
}