package service

import (
	"AuthServer/internal/domain/roles"
	"AuthServer/internal/repository"
	"time"
)

type RBACService struct {
	userRoleRepo *repository.UserRoleRepository
}

func NewRBACService(repo *repository.UserRoleRepository) *RBACService {
	return &RBACService{
		userRoleRepo: repo,
	}
}

func (s *RBACService) AssignRole(userID string, role roles.Role, resourceID *string, expiresAt *time.Time, assignedBy string) error {
	return s.userRoleRepo.AssignRole(userID, role, resourceID, expiresAt, assignedBy)
}

func (s *RBACService) GetUserRoles(userID string) ([]roles.UserRole, error) {
	return s.userRoleRepo.GetUserRoles(userID)
}

func (s *RBACService) RevokeRole(roleID string) error {
	return s.userRoleRepo.RevokeRole(roleID)
}

func (s *RBACService) HasPermission(userID string, requiredRole roles.Role, resourceID *string) (bool, error) {
	userRoles, err := s.userRoleRepo.GetUserRoles(userID)
	if err != nil {
		return false, err
	}

	for _, ur := range userRoles {
		// Skip expired roles
		if ur.ExpiresAt != nil && ur.ExpiresAt.Before(time.Now()) {
			continue
		}

		// Check resource-specific role
		if resourceID != nil && ur.ResourceID != nil {
			if *ur.ResourceID == *resourceID {
				if ur.Role == requiredRole || s.CheckHierarchy(ur.Role, requiredRole) {
					return true, nil
				}
			}
		}

		// Check global role
		if ur.ResourceID == nil {
			if ur.Role == requiredRole || s.CheckHierarchy(ur.Role, requiredRole) {
				return true, nil
			}
		}
	}

	return false, nil
}

func (s *RBACService) CheckHierarchy(userRole roles.Role, requiredRole roles.Role) bool {
	userLevel, userExists := roles.RoleHierarchy[userRole]
	requiredLevel, requiredExists := roles.RoleHierarchy[requiredRole]

	if !userExists || !requiredExists {
		return false
	}

	return userLevel >= requiredLevel
}
