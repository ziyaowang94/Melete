package tendermint

const (
	RoundStepNone = iota
	RoundStepNewHeight
	RoundStepNewRound
	RoundStepPropose
	RoundStepPrevote
	RoundStepPrecommit
	RoundStepCrossShardAccept
	RoundStepCrossShardCommit
)

const (
	RoundStepCrossShardPropose = RoundStepCrossShardAccept
)

func RoundStepString(s int8) string {
	switch s {
	case RoundStepNewHeight:
		return "New-Height"
	case RoundStepNewRound:
		return "NewRound"
	case RoundStepPropose:
		return "Propose"
	case RoundStepPrevote:
		return "Prevote"
	case RoundStepPrecommit:
		return "Precommit"
	case RoundStepCrossShardAccept:
		return "CrossShardAccept"
	case RoundStepCrossShardCommit:
		return "CrossShardCommit"
	default:
		return "Unknown"
	}
}
