package engine

import (
	"time"

	"github.com/gocolly/colly"
)

type ConcurrentEngine struct {
	ItemChan    chan Item
	colly       *colly.Collector
	requestChan chan Request
}

var (
	userAgents = []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.36 Edge/16.16299",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/72.0.3626.109 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_3) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/12.0.3 Safari/605.1.15",
		"Mozilla/5.0 (Macintosh; Intel …) Gecko/20100101 Firefox/64.0",
	}
	count         = 0
	currentEngine *ConcurrentEngine
)

func SetCurrentEngine(engine *ConcurrentEngine) {
	currentEngine = engine
}

func (e *ConcurrentEngine) work() {
	e.colly = colly.NewCollector()
	e.colly.OnResponse(func(response *colly.Response) {
		url := response.Request.URL.String()
		request, ok := parserRequest(url)
		if ok == false {
			return
		}
		result := request.Parser.Parse(response.Body, url)
		for _, r := range result.Requests {
			e.requestChan <- r
		}
		for _, item := range result.Items {
			e.ItemChan <- item
		}
	})
	e.colly.OnError(func(response *colly.Response, e error) {
		key := response.Request.URL.String()
		RemoveKey(key)
	})
	e.colly.OnRequest(func(r *colly.Request) {
		//url := r.URL.String()
		//fmt.Println("OnRequest:", url)
		r.Headers.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8")
		r.Headers.Set("Accept-Encoding", "gzip, deflate")
		r.Headers.Set("Accept-Language", "zh-CN,zh;q=0.9")
		r.Headers.Set("Cache-Control", "no-cache")
		r.Headers.Set("Connection", "keep-alive")
		r.Headers.Set("Content-Encoding", "gzip")
		r.Headers.Set("Content-Type", "application/x-www-form-urlencoded")
		uCount := count % 4
		r.Headers.Set("User-Agent", userAgents[uCount])
		count++
	})

	e.colly.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		RandomDelay: 300 * time.Millisecond,
	})

}

func (e *ConcurrentEngine) Run(seeds ...Request) {
	e.work()
	e.requestChan = make(chan Request)
	go func() {
		for _, r := range seeds {
			SetKeyValue(r.Url, r)
			e.colly.Visit(r.Url)
		}
	}()
	for {
		select {
		case r := <-e.requestChan:
			go e.fetch(r)
		}
	}
}

func (e *ConcurrentEngine) fetch(request Request) {
	SetKeyValue(request.Url, request)
	e.colly.Visit(request.Url)
}

func parserRequest(url string) (Request, bool) {
	var request = Request{}
	v, ok := Value(url)
	if ok == true {
		request, ok = v.(Request)
	}
	return request, ok
}
