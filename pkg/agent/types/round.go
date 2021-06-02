package types

// RoundInfo identifies the current agent "round". Each round is indentified by
// a round number, which is incremented every time the agent is restarted. The
// round number is persisted on the Node in OVSDB.
type RoundInfo struct {
	RoundNum uint64
	// PrevRoundNum is nil if this is the first round or the previous round
	// number could not be retrieved.
	PrevRoundNum *uint64
}
