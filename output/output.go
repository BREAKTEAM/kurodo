package output

import (
	"io"
	"os"

	"github.com/aishee/kurodo/libs/client"
)

var cli Writer
var outputWriter Writer
var outputFile io.Writer

type Writer interface {
	init()
	write(*client.Result)
	writeProgress(*client.Progress)
	close()
}

func Formats() map[string]bool {
	return map[string]bool{"csv": true, "txt": true, "json": true}
}

func SetOutput(filename, outputFormat string) {
	outputFile, _ = os.Create(filename)
	var ow Writer
	switch outputFormat {
	case "csv":
		ow = csv{}
	case "txt":
		ow = txt{}
	case "json":
		ow = json{}
	default:
		ow = null{}
	}
	ow.init()
	outputWriter = ow
	cli = tabCli{}
	cli.init()
}

func Write(r *client.Result) {
	outputWriter.write(r)
	cli.write(r)
}

func WriteProgress(pr *client.Progress) {
	outputWriter.writeProgress(pr)
	cli.writeProgress(pr)
}

func Close() {
	outputWriter.close()
}
