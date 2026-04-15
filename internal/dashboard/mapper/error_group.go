package mapper

import (
	"time"

	"github.com/ernesto2108/anvil/internal/dashboard/dto"
	"github.com/ernesto2108/anvil/internal/dashboard/entity"
)

// ToErrorGroupDTO converts a domain ErrorGroup to an ErrorGroupDTO.
func ToErrorGroupDTO(r entity.ErrorGroup) dto.ErrorGroupDTO {
	return dto.ErrorGroupDTO{
		ID:               r.ID,
		Fingerprint:      r.Fingerprint,
		Title:            r.Title,
		NormalizedMsg:    r.NormalizedMsg,
		ResolutionStatus: r.ResolutionStatus,
		FirstSeenAt:      r.FirstSeenAt.Format(time.RFC3339Nano),
		LastSeenAt:       r.LastSeenAt.Format(time.RFC3339Nano),
		OccurrenceCount:  r.OccurrenceCount,
		Notes:            r.Notes,
		CommitLink:       r.CommitLink,
	}
}

// ToErrorGroupDetailDTO converts domain error group data to the full detail DTO.
func ToErrorGroupDetailDTO(
	group entity.ErrorGroup,
	runs []entity.ErrorGroupRun,
	history []entity.ErrorGroupHistory,
	trend []entity.TrendPoint,
) dto.ErrorGroupDetailDTO {
	runDTOs := make([]dto.ErrorGroupRunDTO, 0, len(runs))
	for _, r := range runs {
		runDTOs = append(runDTOs, dto.ErrorGroupRunDTO{
			RunID:      r.RunID,
			AgentName:  r.AgentName,
			ErrorMsg:   r.ErrorMsg,
			ExitCode:   r.ExitCode,
			OccurredAt: r.OccurredAt,
		})
	}

	historyDTOs := make([]dto.ErrorGroupHistoryDTO, 0, len(history))
	for _, h := range history {
		historyDTOs = append(historyDTOs, dto.ErrorGroupHistoryDTO{
			OldStatus:  h.OldStatus,
			NewStatus:  h.NewStatus,
			Note:       h.Note,
			CommitLink: h.CommitLink,
			CreatedAt:  h.CreatedAt,
		})
	}

	trendDTOs := make([]dto.TrendPointDTO, 0, len(trend))
	for _, t := range trend {
		trendDTOs = append(trendDTOs, dto.TrendPointDTO{Date: t.Date, Count: t.Count})
	}

	return dto.ErrorGroupDetailDTO{
		Group:   ToErrorGroupDTO(group),
		Runs:    runDTOs,
		History: historyDTOs,
		Trend:   trendDTOs,
	}
}
