package utils

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

func NormalizeURL(fuzzURL string) (string, *url.URL, error) {
	if !IsHTTPPrepended(fuzzURL) {
		fuzzURL = PrependHTTP(fuzzURL)
	}
	fuzzURL = strings.TrimSuffix(fuzzURL, "/")
	p, err := url.Parse(fuzzURL)
	if err != nil {
		return "", nil, fmt.Errorf("Unable to parse URL/hostname %s.%s", fuzzURL, err)
	}
	scheme := p.Scheme + "://"
	port := ""
	if p.Port() != "" {
		port = ":" + p.Port()
	}
	query := ""
	if p.RawQuery != "" {
		query = "?" + p.RawQuery
	}
	completeUrl := scheme + p.Hostname() + port + p.Path + query
	parsedUrl, _ := url.Parse(completeUrl)
	return completeUrl, parsedUrl, nil
}

func CountWordlistLines(file string) uint {
	fh, _ := os.Open(file)
	defer fh.Close()

	s := bufio.NewScanner(fh)
	var lc uint
	for s.Scan() {
		lc++
	}
	return lc
}

func CountWords(bytes *[]byte) int {
	numWords := 0
	isWord := false
	for _, c := range *bytes {
		r := rune(c)
		if unicode.IsLetter(r) {
			isWord = true
		} else if isWord && !unicode.IsLetter(r) {
			numWords++
			isWord = false
		}
	}
	return numWords
}

func HeaderSize(h http.Header) int {
	l := 0
	for field, value := range h {
		l += len(field)
		for _, v := range value {
			l += len(v)
		}
	}
	return l
}

func StrArrayToMapStrBool(arr []string) map[int]bool {
	m := map[int]bool{}
	for _, v := range arr {
		if i, err := strconv.Atoi(v); err == nil {
			m[i] = true
		}
	}
	return m
}

func MapToStrArray(m map[string]bool) []string {
	s := []string{}
	for k := range m {
		s = append(s, k)
	}
	return s
}

func IsExtFormatValid(ext string) bool {
	if string(ext[0]) != "." {
		return false
	}
	for _, letter := range ext[1:] {
		if !unicode.IsLetter(rune(letter)) && !unicode.IsDigit(rune(letter)) {
			return false
		}
	}
	return true
}

func IsHTTPPrepended(hostname string) bool {
	match, _ := regexp.MatchString("^http(s)?://", hostname)
	return match
}

func PrependHTTP(hostname string) string {
	return "http://" + hostname
}

func SplitHeaderFields(h, sep string) map[string]string {
	header := make(map[string]string)

	if len(h) == 0 {
		return header
	}
	headerLine := strings.Split(h, sep)
	for _, h := range headerLine {
		sepIndex := strings.Index(h, ":")
		if sepIndex == -1 {
			log.Fatalln("Malformed header name/value. Missing separator colon ':', like name:value")
			continue
		}
		name := strings.TrimSpace(h[:sepIndex])
		value := strings.TrimSpace(h[sepIndex+1:])
		header[name] = value
	}
	return header
}

func ConvertSeparatedCmdArg(argval, sep string) map[int]bool {
	return StrArrayToMapStrBool(strings.Split(argval, sep))
}
