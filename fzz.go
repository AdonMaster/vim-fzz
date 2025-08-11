package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
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

	//
	err := filepath.Walk("./", func(path string, info os.FileInfo, err error) error {
		
		//
		if err != nil {
			return err
		}
	
		//
		println(path)

		//
		return nil

	})

	if err != nil {
		fmt.Printf("%q: %v", "./", err)
		return
	}

}
