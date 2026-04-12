package orchestrator

import (
	"context"
	"fmt"
	"time"

	"github.com/ernesto2108/anvil/internal/instrumentation"
)

// ValidRoles defines the set of recognised agent roles. A pipeline node
// referencing a role outside this set will be rejected at load time.
var ValidRoles = map[string]bool{
	"pm":        true,
	"architect": true,
	"designer":  true,
	"developer": true,
	"tester":    true,
	"qa":        true,
	"devops":    true,
	"dba":       true,
	"security":  true,
}

// ValidateRole returns an error if role is not in ValidRoles.
func ValidateRole(role string) error {
	if !ValidRoles[role] {
		return fmt.Errorf("orchestrator: unknown role %q", role)
	}
	return nil
}

// Node representa un agente dentro del DAG.
type Node struct {
	ID        string
	Role      string
	DependsOn []string
	Gate      bool
	OnFail    FailPolicy
	Timeout   time.Duration // per-node timeout; 0 means no limit
}

type FailPolicy string

const (
	FailPolicyRetry FailPolicy = "retry"
	FailPolicySkip  FailPolicy = "skip"
	FailPolicyAbort FailPolicy = "abort"
	FailPolicyAsk   FailPolicy = "ask"
)

type GateDecision string

const (
	GateApproved GateDecision = "approved"
	GateRejected GateDecision = "rejected"
	GateSkipped  GateDecision = "skipped"
)

type GateState struct {
	NodeID    string
	Decision  GateDecision
	Reason    string
	DecidedAt time.Time
}

type NodeStatus string

const (
	NodePending  NodeStatus = "pending"
	NodeWaiting  NodeStatus = "waiting"
	NodeGated    NodeStatus = "gated"
	NodeRunning  NodeStatus = "running"
	NodeSuccess  NodeStatus = "success"
	NodeFailed   NodeStatus = "failed"
	NodeSkipped  NodeStatus = "skipped"
	NodeRetrying NodeStatus = "retrying"
)

type DAG struct {
	Nodes    map[string]Node
	Edges    map[string][]string // adjacency: node -> dependents
	InDegree map[string]int
	Order    []string
}

type AgentResult struct {
	NodeID       string
	Status       NodeStatus
	DurationMs   int64
	Output       string
	Error        error
	FilesTouched []string
}

type RunState struct {
	RunID       string
	Statuses    map[string]NodeStatus
	Gates       map[string]GateState
	Results     map[string]AgentResult
	StartedAt   time.Time
	FinishedAt  time.Time
	FinalStatus string
}

// AgentRunner executes a single agent node and returns the result.
// upstream contains the AgentResult of every dependency that already completed
// successfully, keyed by node ID.
type AgentRunner interface {
	RunAgent(ctx context.Context, node Node, upstream map[string]AgentResult) (AgentResult, error)
}

// EventSink receives instrumentation events from the orchestrator.
type EventSink interface {
	Emit(ev instrumentation.Event)
}
