package page_db

import (
	"os"
	"bufio"
	"strings"
	"github.com/matus-tomlein/news_processing/environment"
)

type AdsFiltering struct {
	Blacklist [][]string
}

func (ads *AdsFiltering) Init(envType string) {
	f, err := os.Open(environment.AdsFilteringTxtPath(envType))
	if err != nil {
		panic(err)
	}
	r := bufio.NewReader(f)
	for {
		line, _, err := r.ReadLine()
		if err != nil {
			break
		}
		blacklistItems := make([]string, 0)
		arr := strings.Split(string(line), "*")
		for _, str := range arr {
			if str != "" {
				blacklistItems = append(blacklistItems, str)
			}
		}
		if len(blacklistItems) > 0 {
			ads.Blacklist = append(ads.Blacklist, blacklistItems)
		}
	}
}

func (ads *AdsFiltering) MatchUrl(url string) (bool) {
	url = strings.ToLower(url)
	for _, blacklistItems := range ads.Blacklist {
		subUrl := url
		for index, item := range blacklistItems {
			res := strings.Index(subUrl, item)
			if res != -1 {
				if index == len(blacklistItems) - 1 {
					return true
				}
				if index + 1 >= len(subUrl) - 1 {
					break
				}
				subUrl = subUrl[index + 1:len(subUrl) - 1]
			}
		}
	}
	return false
}