package repository

import (
	"AuthServer/internal/database"
	"AuthServer/internal/domain/roles"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
)

type UserRoleRepository struct {
	db *sql.DB
}

func NewUserRoleRepository(s database.Service) *UserRoleRepository {
	return &UserRoleRepository{
		db: s.DB(),
	}
}

func (r *UserRoleRepository) AssignRole(userID string, role roles.Role, resourceID *string, expiresAt *time.Time, createdBy string) error {
	log.Printf("Assigning role %s to user %s", role, userID)

	id := uuid.New().String()

	_, err := r.db.Exec(
		`INSERT INTO user_roles (id, user_id, project_id, role, expires_at, created_by, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, NOW())`,
		id,
		userID,
		resourceID,
		string(role),
		expiresAt,
		createdBy,
	)

	return err
}

func (r *UserRoleRepository) GetUserRoles(userID string) ([]roles.UserRole, error) {
	rows, err := r.db.Query(
		`SELECT user_id, role, project_id, expires_at
		 FROM user_roles
		 WHERE user_id = $1 AND (expires_at IS NULL OR expires_at > NOW())
		 ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userRoles []roles.UserRole
	for rows.Next() {
		var ur roles.UserRole
		var roleStr string
		err := rows.Scan(&ur.UserID, &roleStr, &ur.ResourceID, &ur.ExpiresAt)
		if err != nil {
			log.Printf("failed to scan user role: %v", err)
			continue
		}
		ur.Role = roles.Role(roleStr)
		userRoles = append(userRoles, ur)
	}

	return userRoles, nil
}

func (r *UserRoleRepository) RevokeRole(roleID string) error {
	result, err := r.db.Exec("DELETE FROM user_roles WHERE id = $1", roleID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("role not found")
	}

	return nil
}

func (r *UserRoleRepository) HasRole(userID string, role roles.Role, resourceID *string) (bool, error) {
	var count int
	var query string
	var args []interface{}

	if resourceID == nil {
		query = `SELECT COUNT(*) FROM user_roles 
				 WHERE user_id = $1 AND role = $2 AND project_id IS NULL
				 AND (expires_at IS NULL OR expires_at > NOW())`
		args = []interface{}{userID, string(role)}
	} else {
		query = `SELECT COUNT(*) FROM user_roles 
				 WHERE user_id = $1 AND role = $2 AND project_id = $3
				 AND (expires_at IS NULL OR expires_at > NOW())`
		args = []interface{}{userID, string(role), *resourceID}
	}

	err := r.db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *UserRoleRepository) CleanupExpiredRoles() (int, error) {
	result, err := r.db.Exec(
		"DELETE FROM user_roles WHERE expires_at IS NOT NULL AND expires_at <= NOW()",
	)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	if rowsAffected > 0 {
		log.Printf("Cleaned up %d expired roles", rowsAffected)
	}

	return int(rowsAffected), nil
}
