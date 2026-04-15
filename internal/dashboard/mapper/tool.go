package mapper

import (
	"github.com/ernesto2108/anvil/internal/dashboard/dto"
	"github.com/ernesto2108/anvil/internal/dashboard/entity"
)

// ToToolUsageDTO converts a domain ToolUsage to a ToolUsageDTO.
func ToToolUsageDTO(t entity.ToolUsage) dto.ToolUsageDTO {
	return dto.ToolUsageDTO{
		ToolName: t.ToolName,
		Count:    t.Count,
	}
}

// ToToolUsageDTOs converts a slice of ToolUsage to ToolUsageDTOs.
func ToToolUsageDTOs(rows []entity.ToolUsage) []dto.ToolUsageDTO {
	out := make([]dto.ToolUsageDTO, 0, len(rows))
	for _, t := range rows {
		out = append(out, ToToolUsageDTO(t))
	}
	return out
}

// ToTurnStatsDTO converts a domain TurnStats to a TurnStatsDTO.
func ToTurnStatsDTO(t entity.TurnStats) dto.TurnStatsDTO {
	return dto.TurnStatsDTO{
		TurnNumber:    t.TurnNumber,
		Prompt:        t.Prompt,
		Timestamp:     t.Timestamp,
		EndTimestamp:  t.EndTimestamp,
		FilesCount:    t.FilesCount,
		ToolUsesCount: t.ToolUsesCount,
		AgentsCount:   t.AgentsCount,
	}
}

// ToTurnStatsDTOs converts a slice of TurnStats to TurnStatsDTOs.
func ToTurnStatsDTOs(rows []entity.TurnStats) []dto.TurnStatsDTO {
	out := make([]dto.TurnStatsDTO, 0, len(rows))
	for _, t := range rows {
		out = append(out, ToTurnStatsDTO(t))
	}
	return out
}
