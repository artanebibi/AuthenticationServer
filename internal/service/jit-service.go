package service

import (
	"AuthServer/internal/domain/roles"
	"AuthServer/internal/repository"
	"fmt"
	"time"
)

type JITService struct {
	jitRepo      *repository.JITRequestRepository
	userRoleRepo *repository.UserRoleRepository
}

func NewJITService(jitRepo *repository.JITRequestRepository, userRoleRepo *repository.UserRoleRepository) *JITService {
	return &JITService{
		jitRepo:      jitRepo,
		userRoleRepo: userRoleRepo,
	}
}

func (s *JITService) CreateRequest(userID string, role roles.Role, resourceID *string, durationMinutes int, reason string) (*roles.JITRequestDB, error) {
	return s.jitRepo.Create(userID, role, resourceID, durationMinutes, reason)
}

func (s *JITService) GetPendingRequests() ([]roles.JITRequestDB, error) {
	return s.jitRepo.GetPendingRequests()
}

func (s *JITService) GetUserRequests(userID string) ([]roles.JITRequestDB, error) {
	return s.jitRepo.GetUserRequests(userID)
}

func (s *JITService) ApproveRequest(requestID, approverID string) error {
	request, err := s.jitRepo.GetByID(requestID)
	if err != nil {
		return err
	}

	if request.Status != "pending" {
		return fmt.Errorf("request is not pending")
	}

	// Update request status
	err = s.jitRepo.UpdateStatus(requestID, "approved", &approverID)
	if err != nil {
		return err
	}

	// Assign the role with expiration
	expiresAt := time.Now().Add(time.Duration(request.DurationMinutes) * time.Minute)

	return s.userRoleRepo.AssignRole(
		request.UserID,
		roles.Role(request.Role),
		request.ResourceID,
		&expiresAt,
		approverID,
	)
}

func (s *JITService) RejectRequest(requestID, approverID string) error {
	request, err := s.jitRepo.GetByID(requestID)
	if err != nil {
		return err
	}

	if request.Status != "pending" {
		return fmt.Errorf("request is not pending")
	}

	return s.jitRepo.UpdateStatus(requestID, "rejected", &approverID)
}
