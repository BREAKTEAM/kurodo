package output

import "github.com/aishee/kurodo/libs/client"

type null struct{}

func (null) init()                            {}
func (null) write(r *client.Result)           {}
func (null) writeProgress(p *client.Progress) {}
func (null) close()                           {}
