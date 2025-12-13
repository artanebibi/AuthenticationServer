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

type JITRequestRepository struct {
	db *sql.DB
}

func NewJITRequestRepository(s database.Service) *JITRequestRepository {
	return &JITRequestRepository{
		db: s.DB(),
	}
}

func (r *JITRequestRepository) Create(userID string, role roles.Role, resourceID *string, durationMinutes int, reason string) (*roles.JITRequestDB, error) {
	log.Printf("Creating JIT request for user %s, role %s", userID, role)

	id := uuid.New().String()
	now := time.Now()

	_, err := r.db.Exec(
		`INSERT INTO jit_requests (id, user_id, role, resource_id, duration_minutes, reason, status, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, 'pending', $7, $8)`,
		id,
		userID,
		string(role),
		resourceID,
		durationMinutes,
		reason,
		now,
		now,
	)

	if err != nil {
		return nil, err
	}

	return &roles.JITRequestDB{
		ID:              id,
		UserID:          userID,
		Role:            string(role),
		ResourceID:      resourceID,
		DurationMinutes: durationMinutes,
		Reason:          &reason,
		Status:          "pending",
		CreatedAt:       now,
		UpdatedAt:       now,
	}, nil
}

func (r *JITRequestRepository) GetByID(id string) (*roles.JITRequestDB, error) {
	row := r.db.QueryRow(
		`SELECT id, user_id, role, resource_id, duration_minutes, reason, status, approved_by, created_at, updated_at
		 FROM jit_requests
		 WHERE id = $1`,
		id,
	)

	var req roles.JITRequestDB
	err := row.Scan(
		&req.ID,
		&req.UserID,
		&req.Role,
		&req.ResourceID,
		&req.DurationMinutes,
		&req.Reason,
		&req.Status,
		&req.ApprovedBy,
		&req.CreatedAt,
		&req.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("JIT request not found")
		}
		return nil, err
	}

	return &req, nil
}

func (r *JITRequestRepository) GetPendingRequests() ([]roles.JITRequestDB, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, role, resource_id, duration_minutes, reason, status, approved_by, created_at, updated_at
		 FROM jit_requests
		 WHERE status = 'pending'
		 ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanRequests(rows)
}

func (r *JITRequestRepository) GetUserRequests(userID string) ([]roles.JITRequestDB, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, role, resource_id, duration_minutes, reason, status, approved_by, created_at, updated_at
		 FROM jit_requests
		 WHERE user_id = $1
		 ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanRequests(rows)
}

func (r *JITRequestRepository) UpdateStatus(id, status string, approvedBy *string) error {
	_, err := r.db.Exec(
		`UPDATE jit_requests
		 SET status = $2, approved_by = $3, updated_at = NOW()
		 WHERE id = $1`,
		id,
		status,
		approvedBy,
	)

	if err != nil {
		log.Printf("failed to update JIT request status: %v", err)
	}

	return err
}

func (r *JITRequestRepository) scanRequests(rows *sql.Rows) ([]roles.JITRequestDB, error) {
	var requests []roles.JITRequestDB
	for rows.Next() {
		var req roles.JITRequestDB
		err := rows.Scan(
			&req.ID,
			&req.UserID,
			&req.Role,
			&req.ResourceID,
			&req.DurationMinutes,
			&req.Reason,
			&req.Status,
			&req.ApprovedBy,
			&req.CreatedAt,
			&req.UpdatedAt,
		)
		if err != nil {
			log.Printf("failed to scan JIT request: %v", err)
			continue
		}
		requests = append(requests, req)
	}

	return requests, nil
}
