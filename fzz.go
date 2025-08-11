package main

import (
	"bufio"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// structs
type SearchRequest struct {
    Query string
    Done chan bool
}

// consts
const MAX_RESULTS = 5

//
func main() {

    reqCh := make(chan SearchRequest)
    go dispatcher(reqCh)

    scanner := bufio.NewScanner(os.Stdin)
    for scanner.Scan() {
        query := scanner.Text()
        reqCh <- SearchRequest{Query: query, Done: make(chan bool)}
    }
}


// *************

func dispatcher(reqCh chan SearchRequest) {
    
    var current SearchRequest
    for req := range reqCh {

        // cancel previous
        if current.Done != nil {
            close(current.Done)
        }

        current = req
        go searchWorker(req)

    }

}

func searchWorker(req SearchRequest) {

	var err error

	// build pattern
	var patternBuilder strings.Builder
	patternBuilder.WriteString(".*")

	// words from query
	words := strings.Fields(req.Query)
	if len(words) > 0 {
		for i, word := range words {
			escapedWord := regexp.QuoteMeta(word)
			patternBuilder.WriteString(".*")
			for _, char := range escapedWord {
				patternBuilder.WriteRune(char)
				//patternBuilder.WriteString(".*")
			}
			patternBuilder.WriteString(".*")
			if i < len(words)-1 {
				patternBuilder.WriteString("/")
			}
		}
	}
	pattern := patternBuilder.String()

	// compile pattern
	reg, err := regexp.Compile(pattern)
	if err != nil {
		log.Fatalf("failed to compile regex: %v", err)
	}

	//
	resultCount := 0
	fmt.Println("=======>", pattern)
	err = breadthFirstWalk("./", func(path string, d fs.DirEntry) error {

		// adjusting for folder to end with '/'
		name := path
		if d.Type().IsDir() {
			name = path + "/"
		}


		// printing
		if reg.MatchString(strings.ToLower(path)) {
			fmt.Println(name)
		}

		//
		return nil
	})


	//
	if err != nil {
		fmt.Printf("%q: %v", "./", err)
		return
	}

}


func breadthFirstWalk(root string, cb func(path string, entry os.DirEntry) error) error {

	queue := []string{root}

	for len(queue) > 0 {

		//
		currentDir := queue[0]
		queue = queue[1:]

		//
		entries, err := os.ReadDir(currentDir)
		if err != nil {
			return err
		}

		//
		var subdirs []string
		for _, entry := range entries {
			fullPath := filepath.Join(currentDir, entry.Name())
			cb(fullPath, entry) // <-- cb call
			if entry.IsDir() {
				subdirs = append(subdirs, fullPath)
			}
		}

		//
		queue = append(queue, subdirs...)
	}

	//
	return nil
}
