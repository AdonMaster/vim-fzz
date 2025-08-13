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
	"sync/atomic"
	"time"
)

// errs
var ErrAbort = errors.New("Abort")


// structs
type SearchRequest struct {
	Id int32
    Query string
}

// consts
const MAX_RESULTS = 25
const DEBOUNCE_INTERVAL = 900

// global
var root string
var requestId atomic.Int32

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

	//
    scanner := bufio.NewScanner(os.Stdin)
    for scanner.Scan() {
        query := scanner.Text()
		requestId.Add(1)
		reqCh <- SearchRequest{Id: requestId.Load(), Query: query}
    }
}


// *************

func dispatcher(reqCh chan SearchRequest) {
    
    //
    timer := time.NewTimer(DEBOUNCE_INTERVAL)
    timer.Stop()

    // holds the last request
    var lastRequest *SearchRequest
    for {
        select {

        // on reqCh triggers
        case req, ok := <-reqCh:
            if !ok { return }

            // 
            lastRequest = &req

            // clears the channel
            if !timer.Stop() {
                select {
                case <- timer.C:
                default:
                }
            }
			// restart timer
            timer.Reset(DEBOUNCE_INTERVAL)


        // on timer triggers
        case <-timer.C:
            if lastRequest == nil { return }
            go searchWorker(*lastRequest) // <-- function call
            lastRequest = nil
        }
    }
}

func searchWorker(req SearchRequest) {

	var err error

    // header
	if !sendBuffer(req.Id, "<bof>") {
		return
	}

    // explode qry
	words := strings.Fields(req.Query)

    // empty query prints shallow list
    if len(words) <= 0 {

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

			// 
			if !sendBuffer(req.Id, fullPath) {
				return
			}
        }

		// 
		if !sendBuffer(req.Id, "<eof>") {
			return
		}

        return
    }

    // query not empty, lets roll!!
	// build pattern
	var patternBuilder strings.Builder
	patternBuilder.WriteString("(?i).*")

	// words from query
    for i, word := range words {
        for _, char := range word {
            charString := string(char)
            escapedChar := regexp.QuoteMeta(charString)
            patternBuilder.WriteString(escapedChar)
            if charString != "/" {
                patternBuilder.WriteString("[^/]*")
            }
        }
        if i < len(words)-1 {
            patternBuilder.WriteString("\\/.*")
        }
    }
    patternBuilder.WriteString(".*")
	pattern := patternBuilder.String()
	sendBuffer(req.Id, fmt.Sprintf("<debug v='%s'>", pattern))

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

			if !sendBuffer(req.Id, name) {
				return nil
			}

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
	if err != nil && !errors.Is(err, ErrAbort) {
		fmt.Printf("%q: %v", "./", err)
		return
	}

	// end!? Only checking this condition in case of additional code
	if !sendBuffer(req.Id, "<eof>") {
		return
	}

}


// send buffer out
func sendBuffer(rid int32, path string) bool {

	// is request the current
	if requestId.Load() != rid { return false }
	//
	normalizedRoot := strings.ReplaceAll(root, "\\", "/")
	relativePath := strings.ReplaceAll(path, normalizedRoot, "")
	relativePath = strings.TrimPrefix(relativePath, "/")

	// sleep
	time.Sleep(100 * time.Millisecond)

	//
	fmt.Println(relativePath)

	//
	return true
}
