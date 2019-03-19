package main

import (
    "fmt"
    "time"
    "bufio"
    "os"
)

var strUrls = []string{"https://google.com", "https://facebook.com", "https://amazon.com", "https://golang.com", "https://goodreads.com"}
var strUrls2 = []string{"https://jimhodgson.com", "https://github.com", "https://uni-giessen.de", "https://netflix.com", "https://audible.com", "https://twitter.com", "https://instagram.com"}

func input() string  {
    reader := bufio.NewReader(os.Stdin)
    in, _ := reader.ReadString('\n')
    return in[:len(in)-1]
}

func log(s string) {
    t := time.Now()
    fmt.Printf("[%d:%d:%d] %s\n", t.Hour(), t.Minute(), t.Second(), s)
}

func main() {
    keywordhandler := KeywordHandler{}
    
    /* main menu */
    log("Give me the GR-ID baseline, please.")
    base := input()
    log(fmt.Sprintf("Okay. I'm working on '%s', then. Thanks very much!", base))
    keywordhandler.Include(base)
    kws := keywordhandler.Collapse()
    for _, kw := range kws {
        fmt.Println(kw)
    }
}