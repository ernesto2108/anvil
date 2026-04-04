package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type TargetState struct {
	Status string            `json:"status,omitempty"`
	Pins   map[string]string `json:"pins,omitempty"`
}

type State struct {
	DeployedVersion string                 `json:"deployed_version"`
	DeployedSHA     string                 `json:"deployed_sha,omitempty"`
	DeployedBranch  string                 `json:"deployed_branch,omitempty"`
	PreviousVersion string                 `json:"previous_version"`
	DeployedAt      string                 `json:"deployed_at"`
	Provider        string                 `json:"provider,omitempty"`
	Targets         map[string]TargetState `json:"targets"`

	dir  string
	path string
}

func Load(stateDir string) (*State, error) {
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		return nil, fmt.Errorf("mkdir state dir: %w", err)
	}

	path := filepath.Join(stateDir, "state.json")
	s := &State{
		dir:  stateDir,
		path: path,
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		s.DeployedVersion = "none"
		s.PreviousVersion = "none"
		s.DeployedAt = "never"
		s.Targets = make(map[string]TargetState)
		return s, s.Save()
	}
	if err != nil {
		return nil, fmt.Errorf("read state: %w", err)
	}

	if err := json.Unmarshal(data, s); err != nil {
		return nil, fmt.Errorf("parse state: %w", err)
	}
	s.dir = stateDir
	s.path = path

	if s.Targets == nil {
		s.Targets = make(map[string]TargetState)
	}
	return s, nil
}

func (s *State) Save() error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal state: %w", err)
	}
	data = append(data, '\n')
	return os.WriteFile(s.path, data, 0o644)
}

func (s *State) Dir() string { return s.dir }

func (s *State) SnapshotDir() string {
	return filepath.Join(s.dir, "pre-install")
}

func (s *State) SnapshotComplete() bool {
	_, err := os.Stat(filepath.Join(s.SnapshotDir(), ".snapshot-complete"))
	return err == nil
}

func (s *State) MarkSnapshotComplete() error {
	marker := filepath.Join(s.SnapshotDir(), ".snapshot-complete")
	if err := os.MkdirAll(marker, 0o755); err != nil {
		return fmt.Errorf("mark snapshot complete: %w", err)
	}
	return nil
}

func (s *State) RecordDeploy(version, sha, branch, provider string, targets []string) {
	s.PreviousVersion = s.DeployedVersion
	s.DeployedVersion = version
	s.DeployedSHA = sha
	s.DeployedBranch = branch
	s.DeployedAt = time.Now().UTC().Format("2006-01-02T15:04:05Z")
	s.Provider = provider

	s.Targets = make(map[string]TargetState)
	for _, t := range targets {
		s.Targets[t] = TargetState{Status: "deployed"}
	}
}

func (s *State) SetPin(component, version string) {
	ts, ok := s.Targets["claude"]
	if !ok {
		ts = TargetState{}
	}
	if ts.Pins == nil {
		ts.Pins = make(map[string]string)
	}
	ts.Pins[component] = version
	s.Targets["claude"] = ts
}

func (s *State) RemovePin(component string) {
	ts, ok := s.Targets["claude"]
	if !ok {
		return
	}
	delete(ts.Pins, component)
	s.Targets["claude"] = ts
}

func (s *State) PinCount(prefix string) int {
	ts, ok := s.Targets["claude"]
	if !ok {
		return 0
	}
	count := 0
	for k := range ts.Pins {
		if len(prefix) == 0 || (len(k) > len(prefix) && k[:len(prefix)] == prefix) {
			count++
		}
	}
	return count
}

func (s *State) Remove() error {
	return os.RemoveAll(s.dir)
}
