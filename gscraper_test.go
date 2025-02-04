package gscraper

import (
	"fmt"
	"golang.org/x/text/language"
	"net/http"
	"testing"
)

func TestNewSessionAndSearchImage(t *testing.T) {
	session, err := NewSession(&http.Client{})
	if err != nil {
		t.Fatal(err)
	}
	images, err := session.SearchImage("among us", 1, language.Japanese)
	if err != nil {
		t.Fatal(err)
	}

	for _, img := range images {
		fmt.Println(img.Title, img.Src, img.Href)
	}
}
