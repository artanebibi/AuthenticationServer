package repository

import (
	"AuthServer/internal/database"
	"AuthServer/internal/domain/models"
	"database/sql"
	"fmt"
	"log"
)

type IProjectRepository interface {
	FindById(id string) (*models.Project, error)
	FindByName(name string) (*models.Project, error)
	FindAll() ([]models.Project, error)
	Save(project models.Project) error
	Update(project models.Project) error
	Delete(id string) error
}

type databaseProjectRepository struct {
	db *sql.DB
}

func NewProjectRepository(s database.Service) IProjectRepository {
	return &databaseProjectRepository{
		db: s.DB(),
	}
}

func (d *databaseProjectRepository) FindById(id string) (*models.Project, error) {
	row := d.db.QueryRow(
		"SELECT id, name, description, created_at FROM project WHERE id = $1",
		id,
	)

	var project models.Project
	err := row.Scan(
		&project.ID,
		&project.Name,
		&project.Description,
		&project.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("project not found")
		}
		return nil, fmt.Errorf("failed to scan project: %v", err)
	}

	return &project, nil
}

func (d *databaseProjectRepository) FindByName(name string) (*models.Project, error) {
	row := d.db.QueryRow(
		"SELECT id, name, description, created_at FROM project WHERE name = $1",
		name,
	)

	var project models.Project
	err := row.Scan(
		&project.ID,
		&project.Name,
		&project.Description,
		&project.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("project with name=%s not found", name)
		}
		return nil, fmt.Errorf("failed to scan project by name: %v", err)
	}

	return &project, nil
}

func (d *databaseProjectRepository) FindAll() ([]models.Project, error) {
	rows, err := d.db.Query("SELECT id, name, description, created_at FROM project")
	if err != nil {
		log.Printf("failed to query projects: %v", err)
		return nil, err
	}
	defer rows.Close()

	var projects []models.Project
	for rows.Next() {
		var project models.Project
		if err := rows.Scan(&project.ID, &project.Name, &project.Description, &project.CreatedAt); err != nil {
			log.Printf("failed to scan project: %v", err)
			continue
		}
		projects = append(projects, project)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return projects, nil
}

func (d *databaseProjectRepository) Save(project models.Project) error {
	log.Println("Saving project:", project.Name)
	_, err := d.db.Exec(
		"INSERT INTO project (id, name, description, created_at) VALUES ($1, $2, $3, $4)",
		project.ID,
		project.Name,
		project.Description,
		project.CreatedAt,
	)
	return err
}

func (d *databaseProjectRepository) Update(project models.Project) error {
	_, err := d.db.Exec(
		`UPDATE project 
		SET name = $2, description = $3 
		WHERE id = $1`,
		project.ID,
		project.Name,
		project.Description,
	)

	if err != nil {
		log.Printf("failed to update project %s: %v", project.ID, err)
		return err
	}

	return nil
}

func (d *databaseProjectRepository) Delete(id string) error {
	result, err := d.db.Exec("DELETE FROM project WHERE id = $1", id)
	if err != nil {
		log.Printf("failed to delete project %s: %v", id, err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("project not found")
	}

	return nil
}
