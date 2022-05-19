package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/dgraph-io/badger"
	"golang.org/x/net/html"
)

var pageFlag = flag.String("starting_page", "https://en.wikipedia.org/wiki/Main_Page", "Wikipedia page to start from")

var numPagesFlag = flag.Int("num_pages", 5, "Maximum number of pages to crawl")

var testing = flag.String("testing", "true", "Set this false to use the result of this crawl for keyword searching")

func main() {

	flag.Parse()
	args := flag.Args()
	name := "/tmp/badger"

	db, err := badger.Open(badger.DefaultOptions(name))

	db.DropAll() // empty the badger db
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	fmt.Println(args)
	maxCount := *numPagesFlag
	count := 0
	countDivisor := 20
	start := time.Now()

	//keyword := args[1]

	queue := make(chan string)         // stores the links encountered while scrawling
	filteredQueue := make(chan string) // stores the non-visited links

	go func() { queue <- *pageFlag }()
	go filterQueue(queue, filteredQueue)

	// introduce a bool channel to synchronize execution of concurrently running crawlers
	done := make(chan bool)

	// pull from the filtered queue, add to the unfiltered queue
	go func() {
		for uri := range filteredQueue {
			count += 1

			enqueue(uri, queue, db)
			if count == maxCount {

				fmt.Printf("\nMax count of pages %d reached\n", maxCount)
				break
			}

			if count%countDivisor == 0 {
				fmt.Printf("\n%d / %d pages traversed\n", count, maxCount)

			}
		}
		done <- true
	}()

	<-done
	timeElapsed := int(time.Since(start).Seconds())
	timePerPage := float32(timeElapsed) / float32(count)
	fmt.Println("\nTotal time taken ~ ", timeElapsed, " seconds\nTime per page ~ ", timePerPage, " second(s)/page\n")

	if *testing == "false" {
		exec.Command("rm ", "/tmp/dbForSearch/*")
		cmd := exec.Command("cp", "-r", name, "/tmp/dbForSearch/")

		err := cmd.Run()

		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("\nThis db will be used for search\n")
	}
}

func findWord(db *badger.DB, keyword string, keyCh chan string) {

	count := 1
	for key := range keyCh {

		err1 := db.View(func(txn *badger.Txn) error {
			item, err2 := txn.Get([]byte(key))

			if err2 != nil {

				fmt.Println("Error2!!!")
				os.Exit(0)

			}
			err3 := item.Value(func(val []byte) error {

				fmt.Printf("\n Next page's text is: \n\n\n%s", val)

				return nil
			})

			if err3 != nil {

				fmt.Println("Error3!!!")
				os.Exit(0)

			}
			return nil
		})

		if err1 != nil {

			fmt.Println("Error1!!!")
			os.Exit(0)

		}

		fmt.Printf("\n*************Page %d(url : %s) printing done!***********\n", count, key)
		count++

	}

}

func filterQueue(in chan string, out chan string) { // Transfer links from 'queue' to 'filteredQueue' if they are non-visited
	var seen = make(map[string]bool)
	for val := range in {
		if !seen[val] {
			seen[val] = true
			out <- val
		}
	}
}

//enqueue extracts the links(index of the db) and text(value of the db) from the webpage after crawling through the HTML code of the page
func enqueue(uri string, queue chan string, db *badger.DB) {
	fmt.Println("fetching", uri)

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	client := http.Client{Transport: transport}
	resp, err := client.Get(uri) //get data from uri
	if err != nil {
		return
	}
	defer resp.Body.Close()

	links := All(resp.Body)

	resp, err = client.Get(uri)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	text := getText(resp.Body)
	err = db.Update(func(txn *badger.Txn) error { //Add the text of current page to the db
		err := txn.Set([]byte(uri), []byte(text))
		return err
	})
	var urls []string
	for ind := 0; links[ind] != "\n"; ind++ {
		link := links[ind]
		absolute := fixUrl(link, uri)
		if uri != "" && !strings.HasSuffix(strings.ToLower(link), "jpg") && !strings.HasSuffix(strings.ToLower(link), "jpeg") && !strings.HasSuffix(strings.ToLower(link), "png") && !strings.HasSuffix(strings.ToLower(link), "svg") {

			urls = append(urls, absolute)
		}
	}
	go addToQueue(queue, urls)
}

func addToQueue(q chan string, urls []string) {
	for _, url := range urls {
		q <- url
	}
}

// All takes a reader object (like the one returned from http.Client())
// It returns a slice of strings representing the "href" attributes from
// anchor links found in the provided html.
// It does not close the reader passed to it.
func All(httpBody io.Reader) [100000]string {
	links := [100000]string{}           //Store links
	col := []string{}                   //
	page := html.NewTokenizer(httpBody) // get tokens from HTML
	linksFound := make(map[string]bool)
	ind := 0

	for {
		tokenType := page.Next() // Get next token/tag
		if tokenType == html.ErrorToken {

			links[ind] = "\n"

			return links
		}
		token := page.Token()
		// a indicates link tag
		if tokenType == html.StartTagToken && token.DataAtom.String() == "a" {
			for _, attr := range token.Attr {
				if attr.Key == "href" {

					tl := trimHash(attr.Val) //trims hash from the link

					col = append(col, tl) // stores links in the present page

					resolv(&links, col, linksFound, &ind) // add those links to the total list of unique links

				}
			}
		}

	}

}

// trimHash slices a hash # from the link
func trimHash(l string) string {
	if strings.Contains(l, "#") {
		var index int
		for n, str := range l {
			if strconv.QuoteRune(str) == "'#'" {
				index = n
				break
			}
		}
		return l[:index]
	}
	return l
}

// check looks to see if a url exits in the slice.
func check(sl []string, s string) bool {
	var check bool
	for _, str := range sl {
		if str == s {
			check = true
			break
		}
	}
	return check
}

// resolv adds links to the link slice and insures that there is no repetition
// in our collection.
func resolv(sl *[100000]string, ml []string, linksFound map[string]bool, ind *int) {
	for _, str := range ml {
		if _, ok := linksFound[str]; !ok {
			(*sl)[*ind] = str
			(*ind)++
			linksFound[str] = true

		}
	}
}

//get text from the current page by taking the values inside <p> tags in the page
func getText(httpBody io.Reader) string {
	page := html.NewTokenizer(httpBody)

	res := ""

	for {
		tokenType := page.Next() // Get to the next set of open-close tags

		if tokenType == html.ErrorToken {
			return res
		}
		token := page.Token()
		//If <p> tag is found
		if tokenType == html.StartTagToken && token.DataAtom.String() == "p" {
			for { // Ignore references and take text while closing p tag isn't found
				if tokenType == html.TextToken && (string(token.String()[0]) != "[") {
					res += token.String()
				}
				if tokenType == html.EndTagToken && token.DataAtom.String() == "p" {

					break

				}
				tokenType = page.Next() //Next tag
				token = page.Token()    // Next token or data
			}
		}

	}

}

func fixUrl(href, base string) string { // Convert relative links to the absolute links
	uri, err := url.Parse(href)
	if err != nil {
		return ""
	}
	baseUrl, err := url.Parse(base)
	if err != nil {
		return ""
	}
	uri = baseUrl.ResolveReference(uri)
	return uri.String()
}
