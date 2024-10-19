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
            selection.Find("div").Each(func(indexDiv int, divTag *goquery.Selection) {
                divTag.Find("div").Each(func(indexH3 int, divTag2 *goquery.Selection) {
                    divTag2.Find("h3").Each(func(indexH3 int, h3Tag *goquery.Selection) {
                        h3Tag.Find("div").Each(func(indexH3 int, divTag3 *goquery.Selection) {
                            title = divTag3.Text()
                        })
                    })
                })
            })
            results = append(results, Result{
                Title: title,
                URL:   fmt.Sprintf("https://www.google.com%s", href),
            })
        }
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
