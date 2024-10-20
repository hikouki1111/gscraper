package gscraper

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	requests "github.com/RabiesDev/request-helper"
	"golang.org/x/text/language"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type Session struct {
	cookies []*http.Cookie
	Client  *http.Client
}

func NewSession(client *http.Client) (*Session, error) {
	req := requests.Get("https://www.google.com/")
	res, err := requests.Do(client, req)
	if err != nil {
		return nil, err
	}

	return &Session{
		cookies: res.Cookies(),
		Client:  client,
	}, nil
}

type Result struct {
	Title string
	URL   string
}

func (s *Session) Search(query string, page int, lang language.Tag) ([]Result, error) {
	params := url.Values{}
	params.Add("q", query)
	params.Add("start", strconv.Itoa(maxInt(page-1, 0)*10))
	params.Add("hl", lang.String())

	req := requests.Get(fmt.Sprintf("%s?%s", "https://www.google.com/search", params.Encode()))
	addCookies(req, s.cookies)

	body, res, err := requests.DoAndReadString(s.Client, req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, errors.New("invalid status code")
	}
	s.cookies = res.Cookies()

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		return nil, err
	}

	var nextURL string
	doc.Find("a").Each(func(i int, selection *goquery.Selection) {
		href, exists := selection.Attr("href")
		if exists && strings.HasPrefix(href, "/search?q=") {
			nextURL = href
		}
	})
	if nextURL == "" {
		return nil, errors.New("invalid response body or VPN detected")
	}

	req = requests.Get(fmt.Sprintf("https://www.google.com%s", nextURL))
	addCookies(req, s.cookies)
	body, res, err = requests.DoAndReadString(s.Client, req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, errors.New("invalid status code")
	}
	s.cookies = res.Cookies()

	var results []Result
	doc, err = goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	doc.Find("a").Each(func(i int, selection *goquery.Selection) {
		href, exists := selection.Attr("href")
		if exists && strings.HasPrefix(href, "/url") {
			var title string
			hasTitle := false
			selection.Find("div").Each(func(index int, div *goquery.Selection) {
				div.Find("div").Each(func(index int, div2 *goquery.Selection) {
					div2.Find("h3").Each(func(index int, h3 *goquery.Selection) {
						h3.Find("div").Each(func(index int, div3 *goquery.Selection) {
							title = div3.Text()
							hasTitle = true
						})
					})
				})
			})

			if hasTitle {
				results = append(results, Result{
					Title: title,
					URL:   fmt.Sprintf("https://www.google.com%s", href),
				})
			}
		}
	})

	return results, nil
}

type ImageResult struct {
	Title string
	Src   string
	Href  string
}

func (s *Session) SearchImage(query string, page int, lang language.Tag) ([]ImageResult, error) {
	params := url.Values{}
	params.Add("q", query)
	params.Add("udm", "2")
	params.Add("start", strconv.Itoa(maxInt(page-1, 0)*10))
	params.Add("hl", lang.String())

	req := requests.Get(fmt.Sprintf("%s?%s", "https://www.google.com/search", params.Encode()))
	addCookies(req, s.cookies)

	body, res, err := requests.DoAndReadString(s.Client, req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, errors.New("invalid status code")
	}
	s.cookies = res.Cookies()

	var results []ImageResult
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	doc.Find("tbody").Each(func(i int, tBody *goquery.Selection) {
		var (
			src   string
			title string
			href  string
		)
		tBody.Find("tr").Each(func(i int, tr *goquery.Selection) {
			tr.Find("td").Each(func(i int, td *goquery.Selection) {
				td.Find("a").Each(func(i int, a *goquery.Selection) {
					href, _ = a.Attr("href")
					a.Find("div").Each(func(i int, div *goquery.Selection) {
						div.Find("img").Each(func(i int, img *goquery.Selection) {
							src, _ = img.Attr("src")
						})
						div.Find("span").EachWithBreak(func(i int, span *goquery.Selection) bool {
							span.Find("span").EachWithBreak(func(i int, span2 *goquery.Selection) bool {
								title = span2.Text()
								return false
							})
							return false
						})
					})
				})
			})
		})
		if strings.HasPrefix(href, "/") {
			href = fmt.Sprintf("%s%s", "https://www.google.com", href)
		}

		results = append(results, ImageResult{
			Title: title,
			Src:   src,
			Href:  href,
		})
	})

	return results, nil
}

func (s *Session) GetCookies() []http.Cookie {
	var cookies []http.Cookie
	for _, c := range s.cookies {
		cookies = append(cookies, *c)
	}

	return cookies
}
