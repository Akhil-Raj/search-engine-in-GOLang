# A web crawler written in Go

# HOW TO RUN

## Enable dependency tracking by typing "go mod init example/gocrawler"

## Install requried modules using "go mod tidy"

## You might have to install badger db. Use this link to do that : "https://pkg.go.dev/github.com/dgraph-io/badger#readme-installing"

## Run the go module by typing "go run crawl.go". By default, the crawler will start crawling from the page "https://en.wikipedia.org/wiki/Main_Page" and will crawl upto 5 pages. Use the flag '--starting_page' to provide your own wikipedia page and the flag '--num_pages' to give the maximum number of pages crawled. For example : "go run crawl.go --starting_page https://en.wikipedia.org/wiki/Video_game --num_pages 3"

## The created badger db will be stored in /tmp/badger
