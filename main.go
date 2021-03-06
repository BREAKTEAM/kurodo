package main

import (
	"log"

	"github.com/BREAKTEAM/kurodo/client"
	"github.com/BREAKTEAM/kurodo/opts"
	"github.com/BREAKTEAM/kurodo/output"
)

func main() {
	o := opts.New()
	if err := o.Parse(output.Formats()); err != nil {
		log.Fatal(err)
	}
	output.SetOutput(o.OutputFile, o.OutputFormat)
	chans := client.New(o)
	go client.Start(o)
	for {
		select {
		case r := <-chans.Result:
			output.Write(r)
		case p := <-chans.Progress:
			go output.WriteProgress(p)
		case <-chans.Finished:
			output.Close()
			return
		}
	}
}
