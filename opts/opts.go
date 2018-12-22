package opts

import (
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/aishee/kurodo/libs/utils"
)

type Opts struct {
	URLRaw                  string
	HTTPHideBodyLinesRaw    string
	HTTPHideBodyLengthRaw   string
	HTTPHideNumWordsRaw     string
	HTTPHideHeaderLengthRaw string
	HTTPHideCodesRaw        string
	FileExtensionsRaw       string
	CustomHeader            string
	UserAgent               string
	Cookie                  string
	HTTPMethod              string
	Wordlist                string
	BodyData                string
	OutputFile              string
	OutputFormat            string
	SleepRaw                int
	Timeout                 int
	Concurrency             int
	FollowRedirects         bool
	ProgressOutput          bool
	Show404                 bool
	FileExtensions          []string
	HTTPHideBodyLines       map[int]bool
	HTTPHideBodyLength      map[int]bool
	HTTPHideNumWords        map[int]bool
	HTTPHideHeaderLength    map[int]bool
	HTTPHideCodes           map[int]bool
	URL                     *url.URL
	Sleep                   time.Duration

	// Meta
	FuzzKeyword            string
	HeaderFieldSep         string
	CmdLineValueSep        string
	MaxRequestRetries      uint8
	NumApproxRequests      uint
	NumDoneRequests        uint
	WordlistLineCount      uint
	ProgressSendInterval   int
	FuzzKeywordPresent     bool
	WordlistReadComplete   chan bool
	SupportedOutputFormats map[string]bool
}

func New() *Opts {
	return &Opts{}
}

func (o *Opts) Parse(outputFormats map[string]bool) error {
	o.SupportedOutputFormats = outputFormats
	fs := flag.NewFlagSet("kurodo", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Print("USAGE: kurodo -u domain.com -w wordlist.txt [options]")
		fmt.Println("\n Example: ")
		fmt.Println(" Find hidden files or directories: ")
		fmt.Println(" kurodo -u domain.com -w wordlist.txt")
		fmt.Println("\n Brute force a header field: ")
		fmt.Println(" kurodo -u domain.com -w wordlist.txt -H 'User-Agent: Kurodo'")
		fmt.Println("\n Brute force a file extension: ")
		fmt.Println(" kurodo -u domain.com/file.FUZZ -w ext.txt")
		fmt.Println("\n Brute force a password send via a form: ")
		fmt.Println(" kurodo -u domain.com/login.php -w wordlist.txt -m POST -d 'user=admin&passwd=kurodo&submit-s' -H 'Content-Type: application/x-www-form-urlencoded'")
		fmt.Println("\nOPTIONS: ")
		fs.PrintDefaults()
	}
	fs.StringVar(&o.URLRaw, "u", "", "URL/Hostname.")
	fs.StringVar(&o.Wordlist, "w", "", "Wordlist file.")
	fs.StringVar(&o.HTTPMethod, "m", http.MethodGet, "HTTP method. GET, POST, <CUSTOM>, ...")
	fs.StringVar(&o.HTTPHideCodesRaw, "hc", "", "Hide results with specific HTTP codes, separated by comma. Example: -hc 404,500")
	fs.StringVar(&o.HTTPHideBodyLinesRaw, "hl", "", "Hide results with specific number of lines, separated by comma. Example: -hl 48,1024")
	fs.StringVar(&o.HTTPHideBodyLengthRaw, "hh", "", "Hide results with specific number of chars, separated by comma. Example: -hh 48,1024")
	fs.StringVar(&o.HTTPHideNumWordsRaw, "hw", "", "Hide results with specific number of words, separated by comma. Example: -hw 48,1024")
	fs.StringVar(&o.HTTPHideHeaderLengthRaw, "hr", "", "Hide result with specific header length, separated by comma. Example: -hr 48,1024")
	fs.StringVar(&o.FileExtensionsRaw, "x", "", "Extension to append to the path, separated by comma. Example: -x .php, .html, .jpg")
	fs.StringVar(&o.CustomHeader, "H", "", "Custom header fields, separated by comma. Example: -H 'User-Agent: Chrome,Cookie:Session=abcd'")
	fs.StringVar(&o.BodyData, "d", "", "Post data.")
	fs.StringVar(&o.UserAgent, "a", "", "User-Agent.")
	fs.StringVar(&o.Cookie, "c", "", "Cookie.")
	fs.StringVar(&o.OutputFile, "o", "", "Output file for the results.")
	fs.StringVar(&o.OutputFormat, "of", "", "Format of output file. Currently supported: "+strings.Join(utils.MapToStrArray(outputFormats), ", ")+". Example: -of txt")
	fs.IntVar(&o.Concurrency, "t", 8, "Concurrency level.")
	fs.IntVar(&o.Timeout, "to", 10000, "HTTP timeout in milliseconds.")
	fs.IntVar(&o.SleepRaw, "s", 0, "Sleep time in milliseconds between requests per Go routine.")
	fs.BoolVar(&o.FollowRedirects, "f", false, "Follow 30x redirects.")
	fs.BoolVar(&o.ProgressOutput, "p", true, "Progress output.")
	fs.BoolVar(&o.Show404, "404", false, "Show 404 status code responses.")
	if len(os.Args) <= 1 {
		fs.Usage()
		os.Exit(0)
	}
	fs.Parse(os.Args[1:])
	if err := o.validate(); err != nil {
		return err
	}
	o.initialize()
	return nil
}

func (o *Opts) validate() error {
	if o.URLRaw == "" {
		return fmt.Errorf("No URL/hostname provided. Use flag: -u domain.com")
	}
	if _, _, err := utils.NormalizeURL(o.URLRaw); err != nil {
		return err
	}
	if o.Wordlist == "" {
		return fmt.Errorf("No wordlist provided. Use flag: -w wordlist.txt")
	}
	if _, err := os.Stat(o.Wordlist); os.IsNotExist(err) {
		return fmt.Errorf("Wordlist not found at '%s'", o.Wordlist)
	}
	if o.FileExtensionsRaw != "" {
		for _, ext := range strings.Split(o.FileExtensionsRaw, ",") {
			if !utils.IsExtFormatValid(ext) {
				return fmt.Errorf("Invalid extension %s. Extensions must contain a period followed by alphanummeric letters. Example .php,.html", ext)
			}
		}
	}
	if o.Concurrency < 1 || o.Concurrency > 100 {
		return fmt.Errorf("The concurrency level is invalid. Must be >=1 and <=100")
	}
	if o.OutputFile != "" {
		if o.OutputFormat == "" {
			return fmt.Errorf("Provide an output format with -of")
		}
		if _, err := os.Create(o.OutputFile); err != nil {
			return err
		}
	}
	if o.OutputFormat != "" {
		if o.OutputFile == "" {
			return fmt.Errorf("Provide an output filename with -o")
		}
		o.OutputFormat = strings.ToLower(o.OutputFormat)
		if !o.SupportedOutputFormats[o.OutputFormat] {
			return fmt.Errorf("Only the following output formats are supported: %s", strings.Join(utils.MapToStrArray(o.SupportedOutputFormats), ", "))
		}
	}
	return nil
}

func (o *Opts) initialize() {
	o.WordlistReadComplete = make(chan bool)
	go func() {
		o.WordlistLineCount = utils.CountWordlistLines(o.Wordlist)
		o.NumApproxRequests = o.WordlistLineCount * uint(len(o.FileExtensions))
		o.WordlistReadComplete <- true
	}()
	o.FuzzKeyword = "FUZZ"
	o.CmdLineValueSep, o.HeaderFieldSep = ",", ","
	o.MaxRequestRetries = 3
	o.ProgressSendInterval = 75
	o.URLRaw, o.URL, _ = utils.NormalizeURL(o.URLRaw)
	o.Sleep = time.Duration(o.SleepRaw) * time.Millisecond
	o.HTTPMethod = strings.ToUpper(o.HTTPMethod)
	o.HTTPHideCodes = utils.ConvertSeparatedCmdArg(o.HTTPHideCodesRaw, o.CmdLineValueSep)
	o.HTTPHideBodyLength = utils.ConvertSeparatedCmdArg(o.HTTPHideBodyLengthRaw, o.CmdLineValueSep)
	o.HTTPHideNumWords = utils.ConvertSeparatedCmdArg(o.HTTPHideNumWordsRaw, o.CmdLineValueSep)
	o.HTTPHideBodyLines = utils.ConvertSeparatedCmdArg(o.HTTPHideBodyLinesRaw, o.CmdLineValueSep)
	o.HTTPHideHeaderLength = utils.ConvertSeparatedCmdArg(o.HTTPHideHeaderLengthRaw, o.CmdLineValueSep)
	if !o.Show404 {
		o.HTTPHideCodes[http.StatusNotFound] = true
	}
	if o.FileExtensionsRaw != "" {
		for _, ext := range strings.Split(o.FileExtensionsRaw, o.CmdLineValueSep) {
			o.FileExtensions = append(o.FileExtensions, ext)
		}
	} else {
		// With an empty extension we can enter the main loop once.
		o.FileExtensions = append(o.FileExtensions, "")
	}
	o.FuzzKeywordPresent = func(o *Opts) bool {
		return strings.Contains(o.URL.Path, o.FuzzKeyword) ||
			strings.Contains(o.URL.RawQuery, o.FuzzKeyword) ||
			strings.Contains(o.CustomHeader, o.FuzzKeyword) ||
			strings.Contains(o.BodyData, o.FuzzKeyword) ||
			strings.Contains(o.HTTPMethod, o.FuzzKeyword) ||
			strings.Contains(o.FileExtensionsRaw, o.FuzzKeyword) ||
			strings.Contains(o.UserAgent, o.FuzzKeyword) ||
			strings.Contains(o.Cookie, o.FuzzKeyword)
	}(o)
}
