package main

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
	"github.com/gocolly/colly"
)

var num_of_workers int = 12
var wait_group sync.WaitGroup

/*
scrapeUrl will visit the URLs that it recieves. It will decrease the wait group count once all the tasks are completed

Parameters:
1. url: URL of the site to visit
2. c : instance of Colly collector which will be used to visit the URLs

Output:
No return value,but a request, to the URL passed as a parameter is made
*/
func scrapeUrl(url string, c *colly.Collector) {
	defer wait_group.Done()
	c.Visit(url)
}

/*
It will search the entire page and look for all the emails present and store it in an email list.
The regular expression for email is already specified using regexp

Parameters:
1. c : instance of Colly collector which will be used to visit the URLs

Output:
No reuturn value, prints the emails found on the specific page
*/
func emailFinder(c *colly.Collector) {
	c.OnHTML("body", func(h *colly.HTMLElement) {
		page_text := h.Text
		var emailRegExp = regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)

		email_list := emailRegExp.FindAllString(page_text, -1)
		if len(email_list) > 0 {
			fmt.Println("Emails Found: ")
			for _, email := range email_list {
				fmt.Println(email)
			}
		}
	})

}

/*
Worker function is used to find all the links present on a page. It will find all the anchor tags and get the links from
the href attribute. It will also store the links in a map inorder to ensure that the same links are not visited more
than once

Parameters:
1. c : instance of Colly collector which will be used to visit the URLs
2. visited_links: a map to store the visited URLs
3. UrlChannel: The URLs are sent in to the UrlChannel

Output:
No return values, finds all the URLs and sends them to the UrlChannel
*/
func WorkerFunction(c *colly.Collector, visited_links map[string]bool, UrlChannel chan string) {
	var absolute string
	c.OnHTML("a", func(h *colly.HTMLElement) {
		link := h.Attr("href")
		absolute = h.Request.AbsoluteURL(link)

		if len(absolute) != 0 && strings.HasPrefix(absolute, "https://www.w3.org/staff/") {
			if !visited_links[absolute] {
				visited_links[absolute] = true
				UrlChannel <- absolute
			}
		}else{
			fmt.Println("Some error occured")
		}
	})
}

/*
createWorkers function is used to create worker goroutines as specied by the programmer

Parameters:
1. UrlChannel: The URLs are sent in to the UrlChannel
2. c : instance of Colly collector which will be used to visit the URLs

Output:
No return value,it will create specified number of goroutines
*/
func createWorkers(UrlChannel chan string, c *colly.Collector) {
	for i := 0; i < num_of_workers; i++ {
		go func() {
			for url := range UrlChannel {
				wait_group.Add(1)
				fmt.Println(url)
				scrapeUrl(url, c)
			}
		}()
	}
}

/*
Execute is used to perform all the functions like workerfunction, email finding, creation of workers and visiting the
base URL

Parameters:
1. c : instance of Colly collector which will be used to visit the URLs
2. visited_links: a map to store the visited URLs
3. UrlChannel: The URLs are sent in to the UrlChannel
4. base_url: It stores the base URL

Output:
No return value, it will start the email scraping process by invoking other functions
*/
func execute(c *colly.Collector, base_url string) {
	visited_links := make(map[string]bool)
	UrlChannel := make(chan string, 100)


	WorkerFunction(c, visited_links, UrlChannel)
	emailFinder(c)
	createWorkers(UrlChannel, c)
	c.Visit(base_url)

	go func() {
		wait_group.Wait()
		close(UrlChannel)
	}()
}

func main() {
	base_url := "https://www.w3.org/staff/"
	c := colly.NewCollector()

	execute(c, base_url)

	wait_group.Wait()
	fmt.Println("Completed all tasks")
}
