package events

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/nova/backend/pkg/errors"
)

type EventService struct {
	repo EventRepository
}

func NewEventService(repo EventRepository) *EventService {
	return &EventService{repo: repo}
}

func (s *EventService) FindByID(ctx context.Context, id string) (*Event, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *EventService) FindByCode(ctx context.Context, code string) (*Event, error) {
	return s.repo.FindByCode(ctx, code)
}

func (s *EventService) FindAll(ctx context.Context, tenantID string, org string, limit, offset int) ([]*Event, int, error) {
	return s.repo.FindAll(ctx, tenantID, org, limit, offset)
}

func (s *EventService) FindByOrg(ctx context.Context, org string) ([]*Event, error) {
	return s.repo.FindByOrg(ctx, org)
}

func (s *EventService) FindByObject(ctx context.Context, objectCode, objectOrg string) ([]*Event, error) {
	return s.repo.FindByObject(ctx, objectCode, objectOrg)
}

func (s *EventService) Create(ctx context.Context, tenantID string, req *CreateEventRequest) (*Event, error) {
	event := &Event{
		ID:        uuid.New().String(),
		Code:      req.Code,
		Org:       req.Org,
		Desc:      req.Desc,
		Type:      req.Type,
		RType:     req.RType,
		Status:    req.Status,
		RStatus:   req.RStatus,
		Object:    req.Object,
		ObjectOrg: req.ObjectOrg,
		TenantID:  tenantID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.Create(ctx, event); err != nil {
		return nil, err
	}

	return event, nil
}

func (s *EventService) Update(ctx context.Context, id string, req *UpdateEventRequest) (*Event, error) {
	event, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, errors.ErrNotFound
	}

	if req.Org != "" {
		event.Org = req.Org
	}
	if req.Desc != "" {
		event.Desc = req.Desc
	}
	if req.Type != "" {
		event.Type = req.Type
	}
	if req.RType != "" {
		event.RType = req.RType
	}
	if req.Status != "" {
		event.Status = req.Status
	}
	if req.RStatus != "" {
		event.RStatus = req.RStatus
	}
	if req.Object != "" {
		event.Object = req.Object
	}
	if req.ObjectOrg != "" {
		event.ObjectOrg = req.ObjectOrg
	}
	event.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, event); err != nil {
		return nil, err
	}

	return event, nil
}

func (s *EventService) UpdateStatus(ctx context.Context, id string, status string) (*Event, error) {
	event, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, errors.ErrNotFound
	}

	event.Status = status
	event.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, event); err != nil {
		return nil, err
	}

	return event, nil
}

func (s *EventService) Delete(ctx context.Context, id string) error {
	event, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if event == nil {
		return errors.ErrNotFound
	}

	return s.repo.Delete(ctx, id)
}