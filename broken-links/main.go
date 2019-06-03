package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const (
	base   = "https://directory.archivists.org.au"
	prefix = base + "/index.php/repository/browse?page="
	suffix = "&view=table&sort=alphabetic"
)

func getID(url string) (int, error) {
	doc, err := goquery.NewDocument(base + url)
	if err != nil {
		return 0, err
	}
	sel := doc.Find("#identifyArea > div > div")
	txt := strings.TrimSpace(sel.First().Text())
	if !strings.HasPrefix(txt, "daa/") {
		if strings.HasPrefix(txt, "archives/") {
			return strconv.Atoi(strings.TrimPrefix(txt, "archives/"))
		} else {
			return 0, fmt.Errorf("not a valid ID: %s", txt)
		}
	}
	return strconv.Atoi(strings.TrimPrefix(txt, "daa/"))
}

func getURLs(page int) ([]string, error) {
	doc, err := goquery.NewDocument(fmt.Sprintf("%s%d%s", prefix, page, suffix))
	if err != nil {
		return nil, err
	}
	sel := doc.Find("td > a")
	ret := make([]string, sel.Length())
	sel.Each(func(i int, node *goquery.Selection) {
		ret[i] = node.AttrOr("href", "")
	})
	return ret, nil
}

func main() {
	urls := make([]string, 0, 550)
	for i := 1; i < 56; i++ {
		u, err := getURLs(i)
		if err != nil {
			fmt.Println(err)
			return
		}
		urls = append(urls, u...)
	}
	for _, u := range urls {
		id, err := getID(u)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("/archives/%d %s;\n/archives/%d/ %s;\n", id, u, id, u)
	}
}
