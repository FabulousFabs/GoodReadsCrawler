package main

import "fmt"
import "time"

var strUrls = []string{"https://google.com", "https://facebook.com", "https://amazon.com", "https://golang.com", "https://goodreads.com"}

func main() {
    tStart := time.Now()
    defer func() {
        tEnd := time.Now()
        tElapsed := tEnd.Sub(tStart)
        fmt.Println(tElapsed)
    }()
    
    httphandler := HttpHandler{}
    responses := httphandler.Handle(strUrls, 5)
    
    fmt.Println(responses)
}