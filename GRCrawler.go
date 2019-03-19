/* PSA:
 * Currently throttled. Using one worker only, because the API soft bans for excessive use.
 * Change in ResponseHandler.go:59 if you want to.
 * @to-do: Support for multiple API keys?
 */

package main

import (
    "fmt"
    "time"
    "bufio"
    "os"
    "math"
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
    
    export(base, &keywordhandler)
}

func crawl(httphandler *HttpHandler, keywordhandler *KeywordHandler, poisonpill *bool, next []string) {
    defer log("Alright, I stopped crawling.")
    
    for {
        if *poisonpill {
            break
        }
        
        b := RGoodReadsBook{}
        b.Setup(&*keywordhandler, &*httphandler)
        b.Handle(next)
        
        log(fmt.Sprintf("Keywords total (generic): %d", len(keywordhandler.keywords)))
        
        next = b.next
    }
}

func export(base string, keywordhandler *KeywordHandler) {
    log("Exporting your keywords.")
    
    files := int(math.Ceil(float64(len(keywordhandler.keywords)) / float64(999)))
    
    fmt.Println(files)
    
    for i := 0; i < files; i++ {
        name := fmt.Sprintf("/users/fabianschneider/desktop/programming/go/GRCrawler/export/%s_%d.txt", base, i)
        f, err := os.Create(name)
        if err != nil {
            fmt.Println(err)
        }
        bytes := []byte{}
        for index, kw := range keywordhandler.keywords {
            if index > i * 998 + 998 {
                break
            }
            if index > i * 998 && index < i * 998 + 998 {
                str := fmt.Sprintf("%s\n", kw)
                bb := []byte(str)
                
                for _, bbb := range bb {
                    bytes = append(bytes, bbb)
                }
            }
        }
        b, err2 := f.Write(bytes)
        if err2 != nil {
            fmt.Println(err2)
        }
        f.Close()
        log(fmt.Sprintf("File%d of %d: %d bytes written.", (i+1), files, b))
    }
    
    log("Done! Hope it works well. Bye now!")
}