package orchestrator

import "testing"

func Test_EvaluateFailure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		policy  FailPolicy
		attempt int
		want    ReplanAction
	}{
		// retry policy
		{name: "retry policy attempt 1 → ReplanRetry", policy: FailPolicyRetry, attempt: 1, want: ReplanRetry},
		{name: "retry policy attempt 2 → ReplanAsk", policy: FailPolicyRetry, attempt: 2, want: ReplanAsk},
		{name: "retry policy attempt 3 → ReplanAsk", policy: FailPolicyRetry, attempt: 3, want: ReplanAsk},

		// skip policy
		{name: "skip policy attempt 1 → ReplanSkip", policy: FailPolicySkip, attempt: 1, want: ReplanSkip},
		{name: "skip policy attempt 2 → ReplanSkip", policy: FailPolicySkip, attempt: 2, want: ReplanSkip},

		// abort policy
		{name: "abort policy attempt 1 → ReplanAbort", policy: FailPolicyAbort, attempt: 1, want: ReplanAbort},
		{name: "abort policy attempt 2 → ReplanAbort", policy: FailPolicyAbort, attempt: 2, want: ReplanAbort},

		// ask policy
		{name: "ask policy attempt 1 → ReplanAsk", policy: FailPolicyAsk, attempt: 1, want: ReplanAsk},
		{name: "ask policy attempt 2 → ReplanAsk", policy: FailPolicyAsk, attempt: 2, want: ReplanAsk},

		// empty/default policy
		{name: "empty policy attempt 1 → ReplanAsk", policy: FailPolicy(""), attempt: 1, want: ReplanAsk},
		{name: "unknown policy → ReplanAsk", policy: FailPolicy("unknown"), attempt: 1, want: ReplanAsk},
	}

	for _, tc := range tests {

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			node := Node{ID: "test-node", OnFail: tc.policy}
			got := EvaluateFailure(node, tc.attempt)
			if got != tc.want {
				t.Errorf("EvaluateFailure(policy=%q, attempt=%d) = %q, want %q",
					tc.policy, tc.attempt, got, tc.want)
			}
		})
	}
}
