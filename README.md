# A web crawler written in Go

# How to run

### Install go from the official website

### Download the repo and go into the root folder

### Enable dependency tracking by typing "go mod init example/gocrawler"

### Install requried modules using "go mod tidy"

### You might have to install badger db. Use this link to do that : "https://pkg.go.dev/github.com/dgraph-io/badger#readme-installing"

### Create badger db(database storing the crawled pages) :

##### Run the go module by typing "go run crawl.go". By default, the crawler will start crawling from the page "https://en.wikipedia.org/wiki/Main_Page" and will crawl upto 5 pages in testing mode i.e the created db won't be used for searching the keyword. You can change these flags as per your needs :
###### Use the flag '--starting_page' to provide your own wikipedia page
###### Use the flag '--num_pages' to give the maximum number of pages crawled. 
 Use the flag '--testing' to specify whether to use the resultant db for searching(true or false)
###### For example : "go run crawl.go --starting_page https://en.wikipedia.org/wiki/Video_game --num_pages 3 --testing false"

##### The created badger db will be stored in /tmp/badger. If testing is set to false, a copy will be created in /tmp/dbForSearch/badger which will be used for searching

### Searching the keyword :
##### search for a keyword by typing "go run search.go". By default, the keyword "game" will be search. Provide your own keyword by using "--keyword" flag. Ex : "go run search.go --keyword video"
##### The program will output the urls which contains the keyword as well as the statements preceeding and succeeding the keyword if found

### Demonstration video can be found here : https://drive.google.com/drive/folders/1wcoD_OPPffLA-9X5UEHcRo4gJAvUOjLK?usp=sharing

##### Coming soon : 

###### Time the crawler takes to crawl 1,10,100,1000,10k,100k,1M pages