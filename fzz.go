package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
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

    items := []string{
        "apple", "apricot", "banana", "blackberry", "blueberry",
        "cherry", "date", "fig", "grape", "kiwi",
        "lemon", "lime", "mango", "melon", "orange",
        "papaya", "peach", "pear", "pineapple", "plum",
    }

    //
    fmt.Println("--init")

    //
    count := 0
    for _, item := range items {
        select {

        case <-req.Done:
            return

        default:
            if strings.Contains(item, req.Query) {
                // delay
                time.Sleep(300 * time.Millisecond)
                fmt.Println(item)
                count++
                if count >= MAX_RESULTS {
                    return
                }

            }
        }
    }

}
