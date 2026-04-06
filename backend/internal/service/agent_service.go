package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"backend/internal/model"
	"backend/internal/repository"

	"github.com/google/uuid"
)

var (
	ErrAgentDisabled = errors.New("agent is disabled")
)

type AgentService interface {
	CreateAgent(ctx context.Context, ownerID int64, name string) (*CreateAgentResult, error)
	ListAgentsByOwnerID(ctx context.Context, ownerID int64) ([]model.Agent, error)
	GetAgentByIDAndOwnerID(ctx context.Context, id string, ownerID int64) (*model.Agent, error)
	SetEnabled(ctx context.Context, id string, ownerID int64, enabled bool) error
	AuthenticateByToken(ctx context.Context, token string) (*model.Agent, error)
	Heartbeat(ctx context.Context, agentID string) error
	DeleteAgent(ctx context.Context, id string, ownerID int64) error
}

type CreateAgentResult struct {
	Agent *model.Agent `json:"agent"`
	Token string       `json:"token"`
}

type agentService struct {
	agentRepo repository.AgentRepository
}

func NewAgentService(agentRepo repository.AgentRepository) AgentService {
	return &agentService{
		agentRepo: agentRepo,
	}
}

func (s *agentService) DeleteAgent(ctx context.Context, id string, ownerID int64) error {
	return s.agentRepo.Delete(ctx, id, ownerID)
}

func (s *agentService) CreateAgent(ctx context.Context, ownerID int64, name string) (*CreateAgentResult, error) {
	now := time.Now().UTC()

	rawToken, err := generateSecureToken(32)
	if err != nil {
		return nil, err
	}

	agent := &model.Agent{
		ID:              uuid.NewString(),
		OwnerID:         ownerID,
		Name:            name,
		Enabled:         true,
		Active:          false,
		TokenHash:       hashToken(rawToken),
		LastHeartbeatAt: nil,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := s.agentRepo.Create(ctx, agent); err != nil {
		return nil, err
	}

	return &CreateAgentResult{
		Agent: agent,
		Token: rawToken,
	}, nil
}

func (s *agentService) ListAgentsByOwnerID(ctx context.Context, ownerID int64) ([]model.Agent, error) {
	return s.agentRepo.ListByOwnerID(ctx, ownerID)
}

func (s *agentService) GetAgentByIDAndOwnerID(ctx context.Context, id string, ownerID int64) (*model.Agent, error) {
	return s.agentRepo.GetByIDAndOwnerID(ctx, id, ownerID)
}

func (s *agentService) SetEnabled(ctx context.Context, id string, ownerID int64, enabled bool) error {
	return s.agentRepo.UpdateEnabled(ctx, id, ownerID, enabled, time.Now().UTC())
}

func (s *agentService) AuthenticateByToken(ctx context.Context, token string) (*model.Agent, error) {
	tokenHash := hashToken(token)

	agent, err := s.agentRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, err
	}

	if !agent.Enabled {
		return nil, ErrAgentDisabled
	}

	return agent, nil
}

func (s *agentService) Heartbeat(ctx context.Context, agentID string) error {
	now := time.Now().UTC()
	return s.agentRepo.UpdateHeartbeat(ctx, agentID, true, now, now)
}

func generateSecureToken(numBytes int) (string, error) {
	b := make([]byte, numBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func ComputeAgentStatus(agent *model.Agent, now time.Time) model.AgentStatus {
	if !agent.Enabled {
		return model.AgentStatusDisabled
	}

	if agent.LastHeartbeatAt == nil {
		return model.AgentStatusNeverConnected
	}

	if now.Sub(*agent.LastHeartbeatAt) <= 90*time.Second {
		return model.AgentStatusOnline
	}

	return model.AgentStatusOffline
}

func ToAgentView(agent *model.Agent, now time.Time) model.AgentView {
	return model.AgentView{
		ID:              agent.ID,
		OwnerID:         agent.OwnerID,
		Name:            agent.Name,
		Enabled:         agent.Enabled,
		Active:          agent.Active,
		Status:          ComputeAgentStatus(agent, now),
		LastHeartbeatAt: agent.LastHeartbeatAt,
		CreatedAt:       agent.CreatedAt,
		UpdatedAt:       agent.UpdatedAt,
	}
}

func ToAgentViews(agents []model.Agent, now time.Time) []model.AgentView {
	result := make([]model.AgentView, 0, len(agents))
	for i := range agents {
		agent := agents[i]
		result = append(result, ToAgentView(&agent, now))
	}
	return result
}
