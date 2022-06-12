package healing

type (
	// SourceKit is a custom healing source applied when a player is healed by a kit.
	SourceKit struct{}
	// SourceStew is a custom healing source applied when a player is healed by using stew.
	SourceStew struct{}
	// SourceKill is a custom healing soruce applied when a player is healed by killing another player.
	SourceKill struct{}
)

// HealingSource ...
func (SourceKit) HealingSource() {}

// HealingSource ...
func (SourceStew) HealingSource() {}

// HealingSource ...
func (SourceKill) HealingSource() {}
