package link_search

import (
	"testing"
)

func TestGetSearchResults(t *testing.T) {
	results, err := link_search.Search("Obama")
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	if len(results) == 0 {
		t.Log("No results found")
		t.FailNow()
	}

	for _, result := range results {
		if !strings.Contains(result.Title, "Obama") {
			t.Log("The result does not contain the search query.")
			t.FailNow()
		}
	}
}
