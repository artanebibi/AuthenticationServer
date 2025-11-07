package repository

import (
	"AuthServer/internal/database"
	"AuthServer/internal/domain/models"
	"database/sql"
	"fmt"
	"log"
)

type IUserRepository interface {
	//FindAll() []models.User
	FindById(id string) (*models.User, error)
	FindByEmail(email string) *models.User
	FindByEmailOrUsername(emailOrUsername string) (*models.User, error)
	Save(user models.User) error
	Update(user models.User) error
	Delete(id string) error
}

type databaseUserRepository struct {
	db *sql.DB
}

func NewUserRepository(s database.Service) IUserRepository {
	return &databaseUserRepository{
		db: s.DB(),
	}
}

//func (d *databaseUserRepository) FindAll() []models.User {
//	rows, err := d.db.Query("SELECT id, first_name, last_name, google_email FROM users")
//	if err != nil {
//		log.Println(err)
//		return nil
//	}
//	defer rows.Close()
//
//	var users []models.User
//	for rows.Next() {
//		var u models.User
//		if err := rows.Scan(&u.ID, &u.FirstName, &u.LastName, &u.GoogleEmail); err != nil {
//			log.Println(err)
//			continue
//		}
//		users = append(users, u)
//	}
//	return users
//}

func (d *databaseUserRepository) FindById(id string) (*models.User, error) {
	row := d.db.QueryRow(
		"SELECT * FROM users WHERE id = $1",
		id,
	)

	var user models.User
	err := row.Scan(
		&user.ID,
		&user.FullName,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to scan user: %v", err)
	}

	return &user, nil
}

func (d *databaseUserRepository) FindByEmail(email string) *models.User {
	row := d.db.QueryRow(
		"SELECT * FROM users WHERE email = $1",
		email,
	)

	var user models.User
	err := row.Scan(
		&user.ID,
		&user.FullName,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("user with email=%s not found", email)
			return nil
		}
		log.Printf("failed to scan user by email: %v", err)
		return nil
	}

	return &user
}

func (r *databaseUserRepository) FindByEmailOrUsername(identifier string) (*models.User, error) {
	var user models.User
	err := r.db.QueryRow(
		"SELECT id, full_name, username, email, password, created_at FROM users WHERE email = $1 OR username = $1",
		identifier,
	).Scan(&user.ID, &user.FullName, &user.Username, &user.Email, &user.Password, &user.CreatedAt)

	return &user, err
}

func (d *databaseUserRepository) Save(user models.User) error {
	log.Println("Saving user:", user.FullName)
	_, err := d.db.Exec(
		"INSERT INTO users (id, full_name, username, email, password, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
		user.ID, user.FullName, user.Username, user.Email, user.Password, user.CreatedAt,
	)
	return err
}

func (d *databaseUserRepository) Update(user models.User) error {
	_, err := d.db.Exec(
		`
		UPDATE users
		
		 SET full_name = $2,
		     username = $3,
		     email = $4,
		     password = $5 
	    WHERE id = $1`,
		user.ID,
		user.FullName,
		user.Username,
		user.Email,
		user.Password,
	)
	if err != nil {
		log.Printf("failed to update user %s: %v", user.ID, err)
		return err
	}
	return nil
}

func (d *databaseUserRepository) Delete(id string) error {
	d.db.Exec("DELETE FROM users WHERE id = $1", id)
	return nil
}
