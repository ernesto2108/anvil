package orchestrator

// ReplanAction is the decision returned by EvaluateFailure.
type ReplanAction string

const (
	ReplanRetry ReplanAction = "retry"
	ReplanSkip  ReplanAction = "skip"
	ReplanAbort ReplanAction = "abort"
	ReplanAsk   ReplanAction = "ask"
)

// EvaluateFailure decides what to do when a node fails based on its OnFail
// policy and the attempt number (1-based).
//
// Decision table:
//   - OnFail=retry:          attempt 1 → retry; attempt 2+ → ask
//   - OnFail=skip:           always skip
//   - OnFail=abort:          always abort
//   - OnFail=ask (or empty): always ask
func EvaluateFailure(node Node, attempt int) ReplanAction {
	switch node.OnFail {
	case FailPolicyRetry:
		if attempt <= 1 {
			return ReplanRetry
		}
		return ReplanAsk
	case FailPolicySkip:
		return ReplanSkip
	case FailPolicyAbort:
		return ReplanAbort
	default:
		// Covers FailPolicyAsk and any unset/unknown value.
		return ReplanAsk
	}
}
