package environment

import "fmt"

func PageDbTemplatePath(environment_type string) (string) {
	return "/Users/matus/Programming/gocode/src/github.com/matus-tomlein/news_processing/template.db"
}

func PageDbPath(pageId int, environment_type string) (string) {
	return fmt.Sprintf("/Users/matus/Programming/gocode/bin/data/news_processing/databases/page_%d.db", pageId)
}

func UpdateJsonPath(updateId int, environment_type string) (string) {
	return fmt.Sprintf("/Users/matus/Programming/gocode/bin/data/news_processing/links/%d.json", updateId)
}

func AdsFilteringTxtPath(environment_type string) (string) {
	return "/Users/matus/Programming/gocode/src/github.com/matus-tomlein/news_processing/ads_blacklist.txt"
}

func LinkDensityPath(pageId int, environment_type string) (string) {
	return fmt.Sprintf("/Users/matus/Programming/gocode/bin/data/news_processing/link_densities/%d.json", pageId)
}