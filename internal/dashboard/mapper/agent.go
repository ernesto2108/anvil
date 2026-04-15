package mapper

import (
	"time"

	"github.com/ernesto2108/anvil/internal/dashboard/dto"
	"github.com/ernesto2108/anvil/internal/dashboard/entity"
)

// ToAgentDTO converts a domain AgentRow to an AgentDTO.
func ToAgentDTO(r entity.AgentRow) dto.AgentDTO {
	a := dto.AgentDTO{
		ID:         r.AgentID,
		Name:       r.AgentRole,
		Status:     r.Status,
		DurationMs: r.DurationMs,
	}
	if r.StartedAt != nil {
		a.StartedAt = r.StartedAt.Format(time.RFC3339Nano)
	}
	if r.EndedAt != nil {
		a.EndedAt = r.EndedAt.Format(time.RFC3339Nano)
	}
	return a
}

// ToAgentDetailDTO converts a domain AgentDetail and its files to an AgentDetailDTO.
func ToAgentDetailDTO(d entity.AgentDetail, files []entity.FileRecord) dto.AgentDetailDTO {
	agent := dto.AgentDTO{
		ID:         d.AgentID,
		Name:       d.AgentRole,
		Status:     d.Status,
		DurationMs: d.DurationMs,
		ErrorMsg:   d.ErrorMsg,
	}
	if d.StartedAt != nil {
		agent.StartedAt = d.StartedAt.Format(time.RFC3339Nano)
	}
	if d.EndedAt != nil {
		agent.EndedAt = d.EndedAt.Format(time.RFC3339Nano)
	}

	fileDTOs := make([]dto.FileDTO, 0, len(files))
	for _, f := range files {
		fileDTOs = append(fileDTOs, ToFileDTO(f))
	}

	return dto.AgentDetailDTO{
		Agent:  agent,
		Files:  fileDTOs,
		Output: d.Output,
	}
}

// ToSessionAgentDTO converts a domain AgentRow to a SessionAgentDTO.
// output must be provided separately (fetched from AgentDetail).
func ToSessionAgentDTO(r entity.AgentRow, output string) dto.SessionAgentDTO {
	sa := dto.SessionAgentDTO{
		ID:         r.AgentID,
		Role:       r.AgentRole,
		Status:     r.Status,
		DurationMs: r.DurationMs,
		Output:     output,
	}
	if r.StartedAt != nil {
		sa.StartedAt = r.StartedAt.Format(time.RFC3339Nano)
	}
	if r.EndedAt != nil {
		sa.EndedAt = r.EndedAt.Format(time.RFC3339Nano)
	}
	return sa
}
