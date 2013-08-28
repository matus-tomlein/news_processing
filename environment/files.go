package environment

import "fmt"

func cachePath(envType string) (string) {
	if envType == "production" {
		return "/home/tomlein/webcache"
	}
	return "/Users/matus/Programming/gocode/bin/data/news_processing"
}

func appPath(envType string) (string) {
	if envType == "production" {
		return "/home/tomlein/gocode/src/github.com/matus-tomlein/news_processing"
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
	return fmt.Sprintf("%s/links/%d.json", cachePath(envType), updateId)
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