package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {

    scanner := bufio.NewScanner(os.Stdin)
    for scanner.Scan() {
        line := scanner.Text()
        result := strings.ToUpper(line)
        fmt.Println(result)
        os.Stdout.Sync()
    }

}
