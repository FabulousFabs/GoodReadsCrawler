package main

import (
    "net/http"
    "net/url"
    "io/ioutil"
    "github.com/PuerkitoBio/goquery"
    "sort"
    "fmt"
    "strings"
    "reflect"
    "strconv"
)

type job struct {
    url string
    index int
}

type response struct {
    url string
    index int
    code int
    body string
}

type HttpHandler struct {
    chanJob chan job
    chanProxy chan proxy
    chanResponse chan response
    results []response
    proxiesLoaded bool
    proxies []proxy
}

// supply []string{target urls}, int maximumworkers[, bool useProxy, bool fakeAgent]
func (httphandler HttpHandler) Handle(targets []string, workers int, options ... bool) []response {
    useProxy := false
    maxProxy := 0
    // check for proxy argument
    if len(options) > 0 {
        // use proxies?
        if options[0] {
            useProxy = true
            maxProxy = workers * 4
            httphandler.chanProxy = make(chan proxy, maxProxy)
            
            // make sure proxies are loaded
            if !httphandler.proxiesLoaded {
                loadProxies(&httphandler, maxProxy)
            }
            
            // make sure there can't be more workers than proxies
            if workers > len(httphandler.proxies) {
                workers = len(httphandler.proxies)
            }
        }
    }
    fakeAgent := false
    // check for fake agent argument
    if len(options) > 1 {
        // use fake agents?
        fakeAgent = options[1]
    }
    
    // setup channels + make sure they are defer-closed
    httphandler.chanJob = make(chan job, len(targets))
    httphandler.chanResponse = make(chan response, len(targets))
    defer func(){
        close(httphandler.chanJob)
        close(httphandler.chanResponse)
        httphandler.results = httphandler.results[:0]
        if useProxy {
            close(httphandler.chanProxy)
        }
    }()
    
    // create worker pool
    for i := 0; i < workers; i++ {
        if !useProxy {
            go worker(i, httphandler.chanJob, httphandler.chanResponse, fakeAgent)
        } else {
            go workerProxy(i, httphandler.chanProxy, httphandler.chanJob, httphandler.chanResponse, fakeAgent)
        }
    }
    
    // feed job channel
    for index, url := range targets {
        httphandler.chanJob <- job{url, index}
    }
    
    // feed proxy channel
    if useProxy {
        for _, p := range httphandler.proxies {
            httphandler.chanProxy <- p
        }
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

func worker(index int, chanJob <-chan job, chanResponse chan<- response, agent bool){
    for job := range chanJob {
        client := &http.Client{}
        req, err := http.NewRequest("GET", job.url, nil)
        if agent {
            req.Header.Add("User-Agent", `Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/39.0.2171.27 Safari/537.36`)
        }
        if err != nil {
            fmt.Println(err)
        }
        resp, err2 := client.Do(req)
        if err2 != nil {
            fmt.Println(err2)
        }
        contents, _ := ioutil.ReadAll(resp.Body)
        chanResponse <- response{job.url, job.index, resp.StatusCode, string(contents)}
        resp.Body.Close()
    }
}

type proxy struct {
    ip string
    port int
    address *url.URL
}

func loadProxies(httphandler *HttpHandler, max int) {
    // load web content + initialize goquery
    resp, _ := http.Get("https://free-proxy-list.net")
    contents, _ := ioutil.ReadAll(resp.Body)
    resp.Body.Close()
    document, _ := goquery.NewDocumentFromReader(strings.NewReader(string(contents)))
    
    var p []proxy
    
    // find TRs in document
    document.Find("tr").Each(func(index int, tr *goquery.Selection) {
        // make sure it's a proxy element
        tds := tr.Find("td")
        if tds.Length() != 8 {
            return
        }
        
        // make sure it's a-ok
        if reflect.TypeOf(tds.Get(0).FirstChild.Data).Kind() != reflect.String ||
           reflect.TypeOf(tds.Get(1).FirstChild.Data).Kind() != reflect.String ||
           reflect.TypeOf(tds.Get(4).FirstChild.Data).Kind() != reflect.String ||
           reflect.TypeOf(tds.Get(6).FirstChild.Data).Kind() != reflect.String {
            return
        }
        
        // does this proxy support https?
        if string(tds.Get(6).FirstChild.Data) != "yes" {
            return
        }
        
        // is it transparent?
        if string(tds.Get(4).FirstChild.Data) == "transparent" {
            return
        }
        
        ip := tds.Get(0).FirstChild.Data
        port, _ := strconv.Atoi(tds.Get(1).FirstChild.Data)
        url, _ := url.Parse(fmt.Sprintf("http://%s:%d", ip, port))
        
        p = append(p, proxy{ip, port, url})
    })
    
    // weed out bad proxies (tests)
    for _, pr := range p {
        client := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(pr.address)}}
        resp, err := client.Get("https://jsonplaceholder.typicode.com/todos/1")
        resp.Body.Close()
        
        if err == nil {
            httphandler.proxies = append(httphandler.proxies, pr)
        }
        
        if len(httphandler.proxies) >= max {
            break
        }
    }
}

func workerProxy(index int, chanProxy chan proxy, chanJob chan job, chanResponse chan<- response, agent bool){
    for job := range chanJob {
        proxy := <-chanProxy
        client := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxy.address)}}
        req, _ := http.NewRequest("GET", job.url, nil)
        if agent {
            req.Header.Add("User-Agent", `Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/39.0.2171.27 Safari/537.36`)
        }
        resp, errGet := client.Do(req)
        contents, errIO := ioutil.ReadAll(resp.Body)
        
        if errGet != nil || errIO != nil {
            chanJob <- job
        } else {
            chanResponse <- response{job.url, job.index, resp.StatusCode, string(contents)}
            chanProxy <- proxy
        }
        
        resp.Body.Close()
    }
}
