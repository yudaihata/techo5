//go:build spot

package display

// Photo is a rectangular Show/Panel layout; keep it out of the round Spot's choices.
func photoClockChoices() []struct{ label, value string } { return nil }
