package mapper

import (
	"github.com/ernesto2108/anvil/internal/dashboard/dto"
	"github.com/ernesto2108/anvil/internal/dashboard/entity"
)

// ToFileDTO converts a domain FileRecord to a FileDTO.
func ToFileDTO(f entity.FileRecord) dto.FileDTO {
	return dto.FileDTO{
		Path:    f.Path,
		Action:  f.Operation,
		Diff:    f.Diff,
		AgentID: f.AgentID,
	}
}

// ToFileDTOs converts a slice of FileRecord to FileDTOs.
func ToFileDTOs(rows []entity.FileRecord) []dto.FileDTO {
	out := make([]dto.FileDTO, 0, len(rows))
	for _, f := range rows {
		out = append(out, ToFileDTO(f))
	}
	return out
}
