package output

import (
	"fmt"

	"github.com/aishee/kurodo/libs/client"
)

type csv struct{}

func (csv) init() {
	o := fmt.Sprintf("%s;%s;%s;%s;%s;%s", "Content-Length", "Words", "Lines", "Header", "Status-Code", "Result")
	fmt.Fprintln(outputFile, o)
}

func (csv) write(r *client.Result) {
	o := fmt.Sprintf("%d;%d;%d;%d;%d;%s", r.ContentLength, r.NumWords, r.NumLines, r.HeaderSize, r.StatusCode, r.Result)
	fmt.Fprintln(outputFile, o)
}

func (csv) writeProgress(p *client.Progress) {}
func (csv) close()                           {}
