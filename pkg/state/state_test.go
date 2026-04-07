package state_test

import (
	"path/filepath"
	"testing"

	"github.com/ernesto2108/anvil/pkg/state"
)

func Test_Load_creates_default(t *testing.T) {
	dir := filepath.Join(t.TempDir(), ".forge")

	st, err := state.Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if st.DeployedVersion != "none" {
		t.Errorf("DeployedVersion = %q, want %q", st.DeployedVersion, "none")
	}
	if st.PreviousVersion != "none" {
		t.Errorf("PreviousVersion = %q, want %q", st.PreviousVersion, "none")
	}
	if st.DeployedAt != "never" {
		t.Errorf("DeployedAt = %q, want %q", st.DeployedAt, "never")
	}
}

func Test_Load_reads_existing(t *testing.T) {
	dir := filepath.Join(t.TempDir(), ".forge")

	st, _ := state.Load(dir)
	st.RecordDeploy("v1.0.0", "abc123", "main", "claude", []string{"claude", "opencode"})
	st.Save()

	st2, err := state.Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if st2.DeployedVersion != "v1.0.0" {
		t.Errorf("DeployedVersion = %q, want %q", st2.DeployedVersion, "v1.0.0")
	}
	if st2.DeployedSHA != "abc123" {
		t.Errorf("DeployedSHA = %q, want %q", st2.DeployedSHA, "abc123")
	}
	if st2.Provider != "claude" {
		t.Errorf("Provider = %q, want %q", st2.Provider, "claude")
	}
}

func Test_RecordDeploy_rotates_versions(t *testing.T) {
	dir := filepath.Join(t.TempDir(), ".forge")
	st, _ := state.Load(dir)

	st.RecordDeploy("v1.0.0", "aaa", "main", "claude", []string{"claude"})
	st.RecordDeploy("v2.0.0", "bbb", "main", "claude", []string{"claude"})

	if st.DeployedVersion != "v2.0.0" {
		t.Errorf("DeployedVersion = %q, want %q", st.DeployedVersion, "v2.0.0")
	}
	if st.PreviousVersion != "v1.0.0" {
		t.Errorf("PreviousVersion = %q, want %q", st.PreviousVersion, "v1.0.0")
	}
}

func Test_Pins(t *testing.T) {
	dir := filepath.Join(t.TempDir(), ".forge")
	st, _ := state.Load(dir)

	st.SetPin("skills/go-conventions", "v1.2.0")
	st.SetPin("skills/react-conventions", "v1.0.0")

	if st.PinCount("skills/") != 2 {
		t.Errorf("PinCount = %d, want 2", st.PinCount("skills/"))
	}

	st.RemovePin("skills/go-conventions")

	if st.PinCount("skills/") != 1 {
		t.Errorf("PinCount after remove = %d, want 1", st.PinCount("skills/"))
	}

	st.RemovePin("skills/react-conventions")

	if st.PinCount("skills/") != 0 {
		t.Errorf("PinCount after remove all = %d, want 0", st.PinCount("skills/"))
	}
}

func Test_SnapshotComplete(t *testing.T) {
	dir := filepath.Join(t.TempDir(), ".forge")
	st, _ := state.Load(dir)

	if st.SnapshotComplete() {
		t.Error("snapshot should not be complete initially")
	}

	st.MarkSnapshotComplete()

	if !st.SnapshotComplete() {
		t.Error("snapshot should be complete after marking")
	}
}
