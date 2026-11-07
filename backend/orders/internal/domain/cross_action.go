package domain

type CrossingAction uint8

const (
	CrossingAllow CrossingAction = iota // allow internal crosses
	CrossingCancelTaker
	CrossingCancelMaker
)

type InternalCrossingPolicy interface {
	Decide(taker, maker *Order) CrossingAction
}

type (
	SameTraderCrossingPolicy struct{ Action CrossingAction }
	AllowCrossingPolicy      struct{}
)

func (p SameTraderCrossingPolicy) Decide(taker, maker *Order) CrossingAction {
	if taker.traderID == maker.traderID {
		return p.Action
	}
	return CrossingAllow
}

func (p AllowCrossingPolicy) Decide(taker, maker *Order) CrossingAction {
	return CrossingAllow
}
