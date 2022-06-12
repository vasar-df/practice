package command

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/vasar-network/practice/vasar/user"
	"github.com/vasar-network/vails/lang"
	"os"
	"syscall"
	"time"
)

// Restart is a command used to restart the server.
type Restart struct{}

// Run ...
func (Restart) Run(s cmd.Source, _ *cmd.Output) {
	l := locale(s)

	o := &cmd.Output{}
	o.Print(lang.Translatef(l, "command.restart.begins"))
	s.SendCommandOutput(o)

	user.Broadcast("command.restart.countdown", 15)

	expected := time.Now().Add(time.Second * 15)
	time.AfterFunc(time.Second*4, func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for now := range ticker.C {
			remaining := expected.Sub(now).Round(time.Second)
			user.Broadcast("command.restart.countdown", remaining.Seconds())
			if now.After(expected) {
				if p, err := os.FindProcess(os.Getpid()); err != nil {
					panic(err)
				} else if err = p.Signal(syscall.SIGINT); err != nil {
					panic(err)
				}
				return
			}
		}
	})
}

// Allow ...
func (Restart) Allow(s cmd.Source) bool {
	return allow(s, true)
}
