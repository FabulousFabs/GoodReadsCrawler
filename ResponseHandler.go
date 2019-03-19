package main

import (
    "fmt"
    "github.com/PuerkitoBio/goquery"
    "strings"
    "strconv"
)

const key = "Zvx2qu5J51vpDU3kmyvvQ"

type RGoodReadsBook struct {
    keywordhandler *KeywordHandler
    httphandler *HttpHandler
    next []string
    chanJob chan GRBookJob
    chanResult chan GRBookResult
    results []GRBookResult
}

type GRBookJob struct {
    index int
    body string
}

type GRBookResult struct {
    books []book
}

type book struct {
    id int
    name string
    authors []string
}

func CDATA(s string) string {
    if len(s) > 13 {
        if s[:9] == "<![CDATA[" {
            s = s[9:len(s)-3]
        }
    }
    s = strings.Trim(s, "[]")
    return s
}

func (b *RGoodReadsBook) Setup(keywordhandler *KeywordHandler, httphandler *HttpHandler) {
    b.keywordhandler = keywordhandler
    b.httphandler = httphandler
}

func (b *RGoodReadsBook) Handle(books []string) bool {
    // setup target urls
    targets := []string{}
    for _, book := range books {
        targets = append(targets, fmt.Sprintf("https://goodreads.com/book/show/%s.xml?key=%s&format=xml", book, key))
    }
    
    // get results
    results := b.httphandler.Handle(targets, 5)
    
    // setup channels + make sure we close them
    b.chanJob = make(chan GRBookJob, len(results))
    b.chanResult = make(chan GRBookResult, len(results))
    defer func(){
        close(b.chanJob)
        close(b.chanResult)
    }()
    
    // create workers
    for i := 0; i < len(results); i++ {
        go GRBookWorker(i, b.chanJob, b.chanResult)
    }
    
    // feed jobs
    for _, result := range results {
        b.chanJob <- GRBookJob{result.index, result.body}
    }
    
    // wait and pull
    for {
        resp := <-b.chanResult
        b.results = append(b.results, resp)
        
        for _, book := range resp.books {
            sId := strconv.Itoa(book.id)
            b.next = append(b.next, sId)
            b.keywordhandler.Include(book.name)
            for _, author := range book.authors {
                b.keywordhandler.Include(author)
            }
        }
        
        if len(b.results) == len(results) {
            return true
        }
    }
    
    return false
}

func GRBookWorker(index int, chanJob <-chan GRBookJob, chanResult chan<- GRBookResult) {
    for job := range chanJob {
        document, _ := goquery.NewDocumentFromReader(strings.NewReader(job.body))
        books := []book{}
        document.Find("similar_books").Find("book").Each(func(index int, e *goquery.Selection) {
            id, _ := strconv.Atoi(e.Find("id").Get(0).FirstChild.Data)
            title := CDATA(e.Find("title").Get(0).FirstChild.Data)
            
            authors := []string{}
            e.Find("authors").Find("author").Each(func(inde int, a *goquery.Selection) {
                authors = append(authors, CDATA(a.Find("name").Get(0).FirstChild.Data))
            })
            books = append(books, book{id, title, authors})
        })
        chanResult <- GRBookResult{books}
    }
}