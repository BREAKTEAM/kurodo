package output

import "github.com/BREAKTEAM/kurodo/client"

type null struct{}

func (null) init()                            {}
func (null) write(r *client.Result)           {}
func (null) writeProgress(p *client.Progress) {}
func (null) close()                           {}
