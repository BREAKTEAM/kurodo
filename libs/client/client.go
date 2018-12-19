package client

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/aishee/kurodo/libs/opts"
	"github.com/aishee/kurodo/libs/utils"
)

type ResultChannels struct {
	Result   chan *Result
	Progress chan *Progress
	Finished chan bool
}

// Result contains
type Result struct {
	ContentLength int
	NumWords      int
	StatusCode    int
	NumLines      int
	HeaderSize    int
	Result        string
}

type Progress struct {
	NumDoneRequests   uint
	NumApproxRequests uint
}

type request struct {
	url     string
	data    string
	method  string
	entry   string
	ext     string
	retries uint8
	header  map[string]string
}

var httpClient http.Client
var resultChs ResultChannels

func New(o *opts.Opts) ResultChannels {
	resultChs = ResultChannels{
		Result:   make(chan *Result, o.Concurrency),
		Progress: make(chan *Progress, o.Concurrency),
		Finished: make(chan bool),
	}
	httpClient = initHTTPClient(o)
	return resultChs
}

func Start(o *opts.Opts) {
	queuedReqsCh := make(chan *request, o.Concurrency*o.Concurrency)
	producerDoneCh := make(chan bool)
	concurrencyWg := new(sync.WaitGroup)
	go produceRequests(o, queuedReqsCh, producerDoneCh)
	for i := 0; i < o.Concurrency; i++ {
		concurrencyWg.Add(1)
		go func() {
			for {
				fuzzReq, open := <-queuedReqsCh
				if !open {
					concurrencyWg.Done()
					return
				}
				consumeRequest(o, fuzzReq)
				time.Sleep(o.Sleep)
			}
		}()
	}
	go produceProgress(o)
	<-producerDoneCh
	close(queuedReqsCh)
	concurrencyWg.Wait()
	resultChs.Finished <- true
	close(resultChs.Result)
	close(resultChs.Finished)
}

func produceRequests(o *opts.Opts, queuedReqsCh chan *request, producerDoneCh chan bool) {
	fh, _ := os.Open(o.Wordlist)
	url := strings.TrimSuffix(o.URLRaw, "/")
	header := utils.SplitHeaderFields(o.CustomHeader, o.HeaderFieldSep)
	s := bufio.NewScanner(fh)
	for s.Scan() {
		for _, ext := range o.FileExtensions {
			queuedReqsCh <- &request{
				method: o.HTTPMethod,
				url:    url,
				header: header,
				data:   o.BodyData,
				ext:    ext,
				entry:  s.Text(),
			}
		}
	}
	fh.Close()
	producerDoneCh <- true
}

// produceProgress produces progress information
func produceProgress(o *opts.Opts) {
	if o.ProgressOutput {
		<-o.WordlistReadComplete
		tick := time.Tick(time.Millisecond * time.Duration(o.ProgressSendInterval))
		for {
			select {
			case <-tick:
				resultChs.Progress <- &Progress{
					NumDoneRequests:   o.NumDoneRequests,
					NumApproxRequests: o.NumApproxRequests,
				}
			}
		}
	}
}

func consumeRequest(o *opts.Opts, r *request) {
	res, err := invokeRequest(o, r)
	o.NumDoneRequests++
	if err == nil && isInHideFilter(o, res) {
		resultChs.Result <- res
	}
	if err != nil {
		if r.retries < o.MaxRequestRetries {
			r.retries++
			o.NumApproxRequests++
			consumeRequest(o, r)
		} else {
			log.Printf("Giving up a request, too many errors: %s", err)
		}
	}
}

func invokeRequest(o *opts.Opts, r *request) (*Result, error) {
	var req *http.Request
	var err error
	url := r.url
	if !o.FuzzKeywordPresent {
		r.entry = strings.TrimPrefix(r.entry, "/")
		url = r.url + "/" + r.entry + r.ext
	}
	req, err = http.NewRequest(r.method, url, strings.NewReader(r.data))

	if err != nil {
		return nil, err
	}
	if o.UserAgent != "" {
		req.Header.Set("User-Agent", o.UserAgent)
	}
	if o.Cookie != "" {
		req.Header.Set("Cookie", o.Cookie)
	}
	for h, v := range r.header {
		req.Header.Set(h, v)
	}
	if o.FuzzKeywordPresent {
		req, err = replaceFuzzKeywordOccurence(o, req, r)
		if err != nil {
			return nil, err
		}
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	result := populateResult(resp, r.entry)
	defer resp.Body.Close()
	_, err = io.Copy(ioutil.Discard, resp.Body)
	if err != nil {
		return nil, err
	}
	return result, err
}

// The FUZZ keyword can be everywhere in the HTTP request.
func replaceFuzzKeywordOccurence(o *opts.Opts, req *http.Request, r *request) (*http.Request, error) {
	reqBytes, _ := httputil.DumpRequest(req, true)
	fuzzKeywordBytes := []byte(o.FuzzKeyword)
	entryBytes := []byte(r.entry)
	replaced := bytes.Replace(reqBytes, fuzzKeywordBytes, entryBytes, -1)
	replaced = bytes.Replace(replaced, bytes.Title(bytes.ToLower(fuzzKeywordBytes)), entryBytes, -1)
	reqCopy, err := http.ReadRequest(bufio.NewReader(bytes.NewBuffer(replaced)))
	ext := strings.Replace(r.ext, o.FuzzKeyword, r.entry, -1)
	url := strings.Replace(req.URL.String()+ext, o.FuzzKeyword, r.entry, -1)
	if err != nil {
		return nil, err
	}
	body := strings.Replace(r.data, o.FuzzKeyword, r.entry, -1)
	req, err = http.NewRequest(reqCopy.Method, url, strings.NewReader(body))
	req.Header = reqCopy.Header
	if err != nil {
		return nil, err
	}
	return req, nil
}

func populateResult(resp *http.Response, entry string) *Result {
	b, _ := ioutil.ReadAll(resp.Body)
	if resp.ContentLength == -1 {
		resp.ContentLength = int64(len(b))
	}
	return &Result{
		ContentLength: int(resp.ContentLength),
		NumLines:      bytes.Count(b, []byte{'\n'}),
		NumWords:      utils.CountWords(&b),
		HeaderSize:    utils.HeaderSize(resp.Header),
		StatusCode:    resp.StatusCode,
		Result:        entry,
	}
}

// isInHideFilter determines if some values, sizes, lengths, ...
func isInHideFilter(o *opts.Opts, res *Result) bool {
	return !o.HTTPHideCodes[res.StatusCode] &&
		!o.HTTPHideBodyLength[res.ContentLength] &&
		!o.HTTPHideNumWords[res.NumWords] &&
		!o.HTTPHideBodyLines[res.NumLines] &&
		!o.HTTPHideHeaderLength[res.HeaderSize]
}

// initHTTPClient initialises the default HTTP client with fundamental
func initHTTPClient(o *opts.Opts) http.Client {
	return http.Client{
		Timeout: time.Duration(o.Timeout) * time.Millisecond,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if !o.FollowRedirects {
				return http.ErrUseLastResponse
			}
			return nil
		},
		// Ignore invalid certs by default, since we are interested in the content.
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
}
