package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/extensions"
)

type Comic struct {
	Name        string `json:"name"`
	Url         string `json:"url"`
	Image       string `json:"image"`
	LastChapter string `json:"last_chapter"`
	Status      string `json:"status"`
	Views       string `json:"views"`
	Subscribe   string `json:"subscribe"`
}

func main() {
	// remove data file (if existed) before scraping
	filename := "data.json"
	if _, err := os.Stat(filename); err == nil {
		if err := os.Remove(filename); err != nil {
			log.Fatalln("Error on removing data file:", err)
		}
		fmt.Println("Removing old data file before scraping")
	}

	url := "https://truyenqqq.vn/the-loai/action-26.html"
	domains := []string{"https://truyenqqq.vn", "truyenqqq.vn"}
	comics := []Comic{}

	c := colly.NewCollector(
		colly.AllowedDomains(domains...),
		colly.Async(true), // turn on asynchronous requests
	)

	c.SetRequestTimeout(120 * time.Second)
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*truyenqqq.vn*",
		Parallelism: 5, // limit the parallel requests to 5 request at a time
		RandomDelay: 2 * time.Second,
	})
	extensions.RandomUserAgent(c)

	// set up callbacks
	c.OnHTML("ul.list_grid > li", func(h *colly.HTMLElement) {
		comic := Comic{
			Name:        h.ChildText("div.book_info > div.book_name.qtip > h3 > a"),
			Url:         h.ChildAttr("div.book_avatar > a", "href"),
			Image:       h.ChildAttr("div.book_avatar > a > img", "src"),
			LastChapter: h.ChildText("div.book_info > div.last_chapter > a"),
			Status:      h.ChildText("div.book_info > div.more-info > p:nth-child(2)"),
			Views:       h.ChildText("div.book_info > div.more-info > p:nth-child(3)"),
			Subscribe:   h.ChildText("div.book_info > div.more-info > p:nth-child(4)"),
		}
		comics = append(comics, comic)
	})

	// scrape the next page by get its link and then call c.Visit again
	c.OnHTML("div.page_redirect a[href]", func(h *colly.HTMLElement) {
		if h.Text == "›" {
			nextPage := h.Attr("href")
			c.Visit(nextPage)
		}
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	c.OnResponse(func(r *colly.Response) {
		fmt.Println("Status code", r.StatusCode)
	})

	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("Got error from request:", r.Request.URL, "with error:", err)
	})

	// start to scrape
	start := time.Now()
	c.Visit(url)
	c.Wait()
	end := time.Since(start)

	fmt.Println("Scrape total", len(comics), "comics")
	fmt.Println("Took", end)

	// write data to file
	content, err := json.MarshalIndent(comics, "", "\t")
	if err != nil {
		log.Fatalln(err)
	}

	if err := os.WriteFile(filename, content, 0644); err != nil {
		fmt.Println("Error on writing data to file")
	}
}
