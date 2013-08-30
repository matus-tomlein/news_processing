package environment

import "fmt"

func cachePath(envType string) (string) {
	if envType == "production" {
		return "/home/tomlein/webcache"
	} else if envType == "pi" {
		return "/data/matus/crawled"
	}
	return "/Users/matus/Programming/gocode/bin/data/news_processing"
}

func appPath(envType string) (string) {
	if envType == "production" {
		return "/home/tomlein/go/src/github.com/matus-tomlein/news_processing"
	}
	return "/Users/matus/Programming/gocode/src/github.com/matus-tomlein/news_processing"
}

func PageDbTemplatePath(envType string) (string) {
	return fmt.Sprintf("%s/template.db", appPath(envType))
}

func PageDbPath(pageId int, envType string) (string) {
	return fmt.Sprintf("%s/databases/page_%d.db", cachePath(envType), pageId)
}

func UpdateJsonPath(updateId int, envType string) (string) {
	if envType == "production" {
		return fmt.Sprintf("http://localhost/updates/%d/links.json", updateId)
	}
	return fmt.Sprintf("http://calculon.fiit.stuba.sk:60080/updates/%d/links.json", updateId)
}

func AdsFilteringTxtPath(envType string) (string) {
	return fmt.Sprintf("%s/ads_blacklist.txt", appPath(envType))
}

func LinkDensityPath(pageId int, envType string) (string) {
	return fmt.Sprintf("%s/link_densities/%d.json", cachePath(envType), pageId)
}

func CurrentLinksProcessingIdPath(envType string) (string) {
	return fmt.Sprintf("%s/links_processing_id.txt", cachePath(envType))
}

func DownloadedArticlePath(pageId int, hash uint32, envType string) (string) {
	return fmt.Sprintf("%s/articles/%d/%d.html", cachePath(envType), pageId, hash)
}

func DownloadArticleFolder(envType string, pageId int) (string) {
	return fmt.Sprintf("%s/articles/%d", cachePath(envType), pageId)
}
