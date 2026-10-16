package main

import (
	simple "github.com/oandrew/ipod/lingo-simpleremote"
	"github.com/oandrew/ipod/media"
	"testing"
)

type buttonMedia struct{ calls []string }

func (f *buttonMedia) Snapshot() media.Snapshot { return media.Snapshot{} }
func (f *buttonMedia) Control(a string) error   { f.calls = append(f.calls, a); return nil }
func TestButtonHoldDoesNotSkipRepeatedly(t *testing.T) {
	old := media.Current
	defer func() { media.Current = old; simpleButtons = 0 }()
	f := &buttonMedia{}
	media.Current = f
	simpleButtons = 0
	press := simple.ContextButtonMask(simple.ContextButtonNextTrack)
	handleSimpleButtons(press)
	handleSimpleButtons(press)
	handleSimpleButtons(0)
	handleSimpleButtons(press)
	if len(f.calls) != 2 || f.calls[0] != "Next" {
		t.Fatalf("bad press handling %v", f.calls)
	}
}
