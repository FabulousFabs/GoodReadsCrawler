package main

import (
    "regexp"
    "strings"
)

type KeywordHandler struct {
    keywords []string
}

func (keywordhandler *KeywordHandler) IncludeSlice(keywords []string) {
    for _, keyword := range keywords {
        keywordhandler.Include(keyword)
    }
}

func (keywordhandler *KeywordHandler) Slice(keyword string) []string {
    re := regexp.MustCompile(`([^\-:]+)[\-:]+(.+)`)
    matches := re.FindAllSubmatch([]byte(keyword), -1)
    
    var temp []string
    
    if len(matches) > 0 {
        for index, bytes := range matches[0] {
            if index == 0 {
                continue
            }
            
            temp = append(temp, string(bytes[:len(bytes)]))
        }
    }
    
    return temp
}

func (keywordhandler *KeywordHandler) Sanitize(keyword string) string {
    re := regexp.MustCompile(`[^a-zA-Z0-9\' ]+`)
    keyword = re.ReplaceAllString(keyword, "")
    re = regexp.MustCompile(`[ ]+`)
    keyword = re.ReplaceAllString(keyword, " ")
    keyword = strings.TrimSpace(keyword)
    return keyword
}

func (keywordhandler *KeywordHandler) Add(keyword string) {
    keyword = keywordhandler.Sanitize(keyword)
    keywordhandler.keywords = append(keywordhandler.keywords, keyword)
}

func (keywordhandler *KeywordHandler) Include(keyword string) {
    kws := keywordhandler.Slice(keyword)
    
    keywordhandler.Add(keyword)
    if len(kws) >= 1 {
        keywordhandler.IncludeSlice(kws)
    }
}

func (keywordhandler *KeywordHandler) Includes(keyword string) bool {
    for _, kw := range keywordhandler.keywords {
        if kw == keyword {
            return true
        }
    }
    return false
}

func (keywordhandler *KeywordHandler) Collapse() []string {
    enc := map[string]bool{}
    kws := []string{}
    
    for _, keyword := range keywordhandler.keywords {
        if enc[keyword] == false {
            enc[keyword] = true
            kws = append(kws, keyword)
        }
    }
    
    keywordhandler.keywords = kws
    return kws
}