package main

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// errs
var ErrAbort = errors.New("Abort")


// structs
type SearchRequest struct {
    Query string
    Done chan bool
}

// consts
const MAX_RESULTS = 25

// global
var root string

//
func main() {

    // arguments
    root = "./"
    if len(os.Args) > 1 {
        root = os.Args[1]
    }

    //
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

    // header
	fmt.Println("=======>")

    // empty query prints shallow list
    if req.Query == "" {
        shallowList, err := os.ReadDir(root)
        if err != nil {
            log.Fatalf("failed to retrieve shallow list: %v", err)
        }
        for _, entry := range shallowList {
            fullPath := filepath.Join(root, entry.Name())
            fullPath = strings.ReplaceAll(fullPath, "\\", "/")
            if entry.IsDir() {
                fullPath = fullPath + "/"
            }
            fmt.Println(fullPath)
        }
        return
    }

    // query not empty, lets roll!!
	// build pattern
	var patternBuilder strings.Builder
	patternBuilder.WriteString(".*")

	// words from query
	words := strings.Fields(strings.ReplaceAll(req.Query, ".", ""))
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

    // walk dir
	resultCount := 0
    err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {

        // err
        if err != nil {
            return err
        }

		// adjusting for folder to end with '/'
		name := strings.ReplaceAll(path, "\\", "/")
		if d.Type().IsDir() {
			name = name + "/"
		}

		// printing
		if reg.MatchString(strings.ToLower(name)) {

			fmt.Println(name)

            // result count and limit
            resultCount++
            if resultCount >= MAX_RESULTS {
                return ErrAbort
            }
            
		}

		//
		return nil
	})


	//
	if err != nil {

        //
        if errors.Is(err, ErrAbort) {
            return
        }

        //
		fmt.Printf("%q: %v", "./", err)
		return
	}

}


func breadthFirstWalk(dir string, cb func(path string, entry os.DirEntry) error) error {

	queue := []string{dir}

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
            fullPath = strings.ReplaceAll(fullPath, "\\", "/")

            // call
			err = cb(fullPath, entry)
            if err != nil {
                // abort error
                if errors.Is(err, ErrAbort) {
                    return nil
                }
                return err
            }

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


// =====
func strContains(s []string, str string) bool {
    for _, v := range s {
        if v == str {
            return true
        }
    }
    return false
}
