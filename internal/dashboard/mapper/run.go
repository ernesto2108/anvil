package mapper

import (
	"time"

	"github.com/ernesto2108/anvil/internal/dashboard/dto"
	"github.com/ernesto2108/anvil/internal/dashboard/entity"
)

// ToRunDTO converts a domain RunSummary to the frontend DTO.
func ToRunDTO(r entity.RunSummary) dto.RunDTO {
	endedAt := ""
	if r.EndedAt != nil {
		endedAt = r.EndedAt.Format(time.RFC3339Nano)
	}

	var durationMs int64
	if r.DurationMs != nil {
		durationMs = *r.DurationMs
	}

	return dto.RunDTO{
		ID:            r.ID,
		TaskID:        r.TaskID,
		TaskDesc:      r.TaskDesc,
		Status:        r.Status,
		Complexity:    r.Complexity,
		Provider:      r.Provider,
		Project:       r.Project,
		StartedAt:     r.StartedAt.Format(time.RFC3339Nano),
		EndedAt:       endedAt,
		DurationMs:    durationMs,
		FilesCount:    r.FilesCount,
		AgentsCount:   r.AgentsCount,
		ParentRunID:   r.ParentRunID,
		ChildrenCount: r.ChildrenCount,
		Branch:        r.Branch,
		ErrorReason:   r.ErrorReason,
	}
}

// ToRunDTOs converts a slice of RunSummary to RunDTOs.
func ToRunDTOs(rows []entity.RunSummary) []dto.RunDTO {
	out := make([]dto.RunDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, ToRunDTO(r))
	}
	return out
}
