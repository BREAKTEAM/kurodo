package output

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/BREAKTEAM/kurodo/client"
)

type tabCli struct{}

var tableWriter *tabwriter.Writer

func (tabCli) init() {
	fmt.Println("Kurodo Fuzzy Tools By Aishee")
	tableWriter = new(tabwriter.Writer)
	tableWriter.Init(os.Stdout, 13, 0, 0, ' ', 0)

	fmt.Fprintln(tableWriter, "")
	fmt.Fprintln(tableWriter, "Chars(-hh) \t Words(-hw) \t Lines(-hl) \t Header(-hr) \t Code(-hc) \t Result")
	fmt.Fprintln(tableWriter, "")
}

func (tabCli) write(r *client.Result) {
	o := fmt.Sprintf("%d \t %d \t %d \t %d \t %d \t %s", r.ContentLength, r.NumWords, r.NumLines, r.HeaderSize, r.StatusCode, r.Result)
	fmt.Fprintln(tableWriter, o)
	tableWriter.Flush()
}

func (tabCli) writeProgress(p *client.Progress) {
	percent := int((float64(p.NumDoneRequests) / float64(p.NumApproxRequests)) * 100)
	fmt.Printf("\r%30s\r~%d/%d (%d%%)\r", "", p.NumDoneRequests, p.NumApproxRequests, percent) // Output
}

func (tabCli) close() {
	fmt.Printf("\r%30s\r", "")
}
