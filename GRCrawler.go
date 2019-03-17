package main

import "fmt"
import "net/http"

const iThreadCount = 3;
var strUrls = [...]string{"https://google.com", "https://facebook.com", "https://amazon.com", "https://golang.com", "https://goodreads.com"}

type response struct {
    url string
    index int
    code int
}

func request(url string, index int, chanResponse chan<- response) {
    resp, _ := http.Get(url)
    defer resp.Body.Close()
    chanResponse <- response{url, index, resp.StatusCode};
}

func main() {
    // setup channel + make sure it gets closed
    chanResponse := make(chan response, iThreadCount)
    defer close(chanResponse)
    
    // setup async requests
    for index, url := range strUrls {
        go request(url, index, chanResponse)
    }
    
    // wait for answers
    var responses []response
    for {
        resp := <-chanResponse
        responses = append(responses, resp)
        
        if (len(responses) == len(strUrls)) {
            break
        }
    }
    
    // did it work?
    fmt.Println(len(responses))
    fmt.Println(responses)
}