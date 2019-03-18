package main

import "net/http"
import "sort"

type job struct {
    url string
    index int
}

type response struct {
    url string
    index int
    code int
}

type HttpHandler struct {
    chanJob chan job
    chanResponse chan response
    results []response
}

func (httphandler HttpHandler) Handle(targets []string, workers int) []response {
    // setup channels + make sure they are defer-closed
    httphandler.chanJob = make(chan job, len(targets))
    httphandler.chanResponse = make(chan response)
    defer func(){
        close(httphandler.chanJob)
        close(httphandler.chanResponse)
        httphandler.results = httphandler.results[:0]
    }()
    
    // create worker pool
    for i := 0; i < workers; i++ {
        go worker(i, httphandler.chanJob, httphandler.chanResponse)
    }
    
    // feed job channel
    for index, url := range targets {
        httphandler.chanJob <- job{url, index}
    }
    
    // wait and pull results
    for {
        resp := <-httphandler.chanResponse
        httphandler.results = append(httphandler.results, resp)
                
        if (len(httphandler.results) == len(targets)) {
            break
        }
    }
    
    // sort slice before returning
    sort.Slice(httphandler.results, func(i, j int) bool {
        return httphandler.results[i].index < httphandler.results[j].index
    })
    
    return httphandler.results
}

func worker(index int, chanJob <-chan job, chanResponse chan<- response){
    for job := range chanJob {
        resp, _ := http.Get(job.url)
        chanResponse <- response{job.url, job.index, resp.StatusCode}
        resp.Body.Close()
    }
}