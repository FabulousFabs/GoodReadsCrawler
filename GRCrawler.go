package main

import "fmt"
import "time"

var strUrls = []string{"https://google.com", "https://facebook.com", "https://amazon.com", "https://golang.com", "https://goodreads.com"}
var strUrls2 = []string{"https://jimhodgson.com", "https://github.com", "https://uni-giessen.de", "https://netflix.com", "https://audible.com", "https://twitter.com", "https://instagram.com"}

func main() {
    tStart := time.Now()
    defer func() {
        tEnd := time.Now()
        tElapsed := tEnd.Sub(tStart)
        fmt.Println(tElapsed)
    }()
    
    httphandler := HttpHandler{}
    
    r := httphandler.Handle(strUrls2, 5)
    fmt.Println(r)
}