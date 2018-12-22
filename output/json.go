package output

import (
	jsons "encoding/json"
	"fmt"
	"log"

	"github.com/aishee/kurodo/libs/client"
)

type json struct{}

var jsonResults []*client.Result

func (json) write(r *client.Result) {
	jsonResults = append(jsonResults, r)
}

func (json) close() {
	json, err := jsons.Marshal(jsonResults)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintln(outputFile, string(json))
}

func (json) init()                            {}
func (json) writeProgress(p *client.Progress) {}
