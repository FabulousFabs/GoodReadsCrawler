package main

import (
    "fmt"
    "time"
    "bufio"
    "os"
)

func input(arg ... bool) string  {
    if len(arg) == 0 {
        fmt.Printf("> ")
    }
    reader := bufio.NewReader(os.Stdin)
    in, _ := reader.ReadString('\n')
    return in[:len(in)-1]
}

func log(s string) {
    t := time.Now()
    fmt.Printf("[%d:%d:%d] %s\n", t.Hour(), t.Minute(), t.Second(), s)
}

func main() {
    httphandler := HttpHandler{}
    keywordhandler := KeywordHandler{}
    
    // prompt baseline
    log("Give me the GR-ID baseline, please.")
    base := input()
    log(fmt.Sprintf("Okay. I'm working on '%s', then. Thanks very much!", base))
    
    // load baseline from GR API
    b := RGoodReadsBook{}
    b.Setup(&keywordhandler, &httphandler)
    b.Handle([]string{base})
    log("Loaded.")
    
    // start crawling in second thread
    log("Starting to crawl.")
    poisonpill := false
    go crawl(&httphandler, &keywordhandler, &poisonpill, b.next)
    
    log("I'm crawling. Press any key to stop.")
    input(false)
    poisonpill = true
    
    keywordhandler.Collapse()
    log(fmt.Sprintf("Keywords total (unique): %d", len(keywordhandler.keywords)))
    /*for _, kw := range keywordhandler.keywords {
        log(kw)
    }*/
}

func crawl(httphandler *HttpHandler, keywordhandler *KeywordHandler, poisonpill *bool, next []string) {
    for {
        if *poisonpill {
            log("Alright, I stopped crawling.")
            break
        }
        
        b := RGoodReadsBook{}
        b.Setup(&*keywordhandler, &*httphandler)
        b.Handle(next)
        
        log(fmt.Sprintf("Keywords total (generic): %d", len(keywordhandler.keywords)))
        
        next = b.next
    }
}