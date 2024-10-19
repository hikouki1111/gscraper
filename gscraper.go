package gscraper

import "net/http"

func addCookies(req *http.Request, cookies []*http.Cookie) {
    for _, c := range cookies {
        req.AddCookie(c)
    }
}

func maxInt(a, b int) int {
    if a > b {
        return a
    }

    return b
}
