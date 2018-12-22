package output

import (
	"fmt"

	"github.com/aishee/kurodo/libs/client"
)

type txt struct{}

func (txt) init() {
	o := fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s", "Content-Length", "Words", "Lines", "Header", "Status-Code", "Result")
	fmt.Fprintln(outputFile, o)
}

func (txt) write(r *client.Result) {
	o := fmt.Sprintf("%d\t\t\t\t%d\t\t%d\t\t%d\t\t%d\t\t\t%s", r.ContentLength, r.NumWords, r.NumLines, r.HeaderSize, r.StatusCode, r.Result)
	fmt.Fprintln(outputFile, o)
}

func (txt) writeProgress(p *client.Progress) {}
func (txt) close()                           {}
