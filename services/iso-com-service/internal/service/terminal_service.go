package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/iso-com-service/internal/repository"
)

type TerminalService struct {
	repo repository.TerminalCredentialRepository
}

func NewTerminalService(repo repository.TerminalCredentialRepository) *TerminalService {
	return &TerminalService{repo: repo}
}

func (s *TerminalService) CreateTerminalCredentials(
	ctx context.Context,
	userID string,
	hostType string,
	tid string,
	mid string,
	active bool,
) error {
	u, err := uuid.Parse(strings.TrimSpace(userID))
	if err != nil {
		return fmt.Errorf("invalid userId")
	}

	hostType = strings.TrimSpace(hostType)
	if hostType != "HOST1_ISO8583" && hostType != "HOST2_REST" {
		return fmt.Errorf("invalid hostType")
	}

	tid = strings.TrimSpace(tid)
	mid = strings.TrimSpace(mid)

	if len(tid) != 8 {
		return fmt.Errorf("tid must be exactly 8 chars")
	}
	if len(mid) != 15 {
		return fmt.Errorf("mid must be exactly 15 chars")
	}

	return s.repo.Create(ctx, u, hostType, tid, mid, active)
}
