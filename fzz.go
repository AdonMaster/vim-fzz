package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

// errs
var ErrAbort = errors.New("Abort")
var ErrMaxResults = errors.New("MaxResults")


// consts
const MAX_RESULTS = 25
const DEBOUNCE_INTERVAL = 750 * time.Millisecond
const SIMULATE_SLOW = 10 * time.Millisecond

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
	var wg sync.WaitGroup

    //
    reqCh := make(chan string)
    go dispatcher(reqCh, &wg)


	//
    scanner := bufio.NewScanner(os.Stdin)
    for scanner.Scan() {
        query := scanner.Text()
		reqCh <-query 
    }

	//
	wg.Wait()
}


// *************

func dispatcher(reqCh chan string, wg *sync.WaitGroup) {
    
    //
    var cancel context.CancelFunc

    //
    timer := time.NewTimer(DEBOUNCE_INTERVAL)
    timer.Stop()

    // holds the last request
	var lastRequest *string 
    for {
        select {

        // on reqCh triggers
        case req, ok := <-reqCh:
            if !ok { return }

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

			//
			if cancel != nil { cancel() }
            if lastRequest == nil { return }

			//
			ctx, newCancel := context.WithCancel(context.Background())
			cancel = newCancel

			//
            go searchWorker(ctx, *lastRequest, wg) // <-- function call
            lastRequest = nil
        }
    }
}

func searchWorker(ctx context.Context, req string, wg *sync.WaitGroup) {

	// main thread locking
	wg.Add(1)
	defer wg.Done()
	
	//
	var err error

    // header
	sendBuffer("<bof>")

    // explode qry
	words := strings.Fields(req)

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
			sendBuffer(fullPath)
        }

		// 
		sendBuffer("<eof>")

		//
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
	sendBuffer(fmt.Sprintf("<debug v='%s'>", pattern))

	// compile pattern
	reg, err := regexp.Compile(pattern)
	if err != nil {
		log.Fatalf("failed to compile regex: %v", err)
	}

    // walk dir
	resultCount := 0
    err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {

		select {
			
		// context cancelled?
		case <-ctx.Done():
			return ErrAbort
			
		// let's move on
		default:

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

				//
				sendBuffer(name)

				// result count and limit
				resultCount++
				if resultCount >= MAX_RESULTS {
					return ErrMaxResults
				}
				
			}
		}

		//
		return nil
	})


	//
	if err != nil {
		if errors.Is(err, ErrAbort) {
			sendBuffer(fmt.Sprintf("<debug v='%v'>", err))
			return
		}
		if !errors.Is(err, ErrMaxResults) {
			sendBuffer(fmt.Sprintf("<error v='%v'>", err))
			return
		}
	}

	// 
	sendBuffer("<eof>")

}


// send buffer out
func sendBuffer(path string) bool {

	//
	normalizedRoot := strings.ReplaceAll(root, "\\", "/")
	relativePath := strings.ReplaceAll(path, normalizedRoot, "")
	relativePath = strings.TrimPrefix(relativePath, "/")

	// sleep
	time.Sleep(SIMULATE_SLOW)

	//
	fmt.Println(relativePath)

	//
	return true
}
