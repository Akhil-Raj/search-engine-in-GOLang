package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/dgraph-io/badger"
)

var keyword = flag.String("keyword", "game", "keyword to search for")

func main() {

	name := "/tmp/dbForSearch/badger"
	db, err := badger.Open(badger.DefaultOptions(name)) // open the db
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	flag.Parse()
	args := flag.Args()
	fmt.Println(args)
	start := time.Now()

	findWord(db, *keyword)

	fmt.Println("\nTime taken for search : ", time.Since(start).Seconds(), " seconds\n")

}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func findWord(db *badger.DB, keyword string) {

	err := db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() { // iterate through the keys of the db
			item := it.Item()
			k := item.Key()
			err := item.Value(func(v []byte) error {
				s := string(v)
				l := len(s)
				puncts := []string{",", ".", "!", " ", "?", "\t"} //possible punctuations before and after the keyword
				ind := -1
				for _, st := range puncts {
					if ind != -1 {
						break
					}
					for _, en := range puncts {
						word := st + keyword + en
						if ind = strings.Index(s, word); ind != -1 {
							fmt.Printf("\nKeyword found!\nKeyword : %s\nURI : %s\nStatement : %s\n", keyword, k, string(v)[max(0, ind-200):min(ind+200, l)])
							break
						}

					}

				}

				return nil
			})

			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {

		fmt.Println("\nError in iterating!\n")
		os.Exit(0)

	}

}
