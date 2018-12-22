package main

import (
	"log"

	"github.com/BREAKTEAM/kurodo/libs/client"
	"github.com/BREAKTEAM/kurodo/libs/opts"
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
