package model

import "time"

type UserRequest struct {
	Username string `json:"username" db:"username" validate:"required,min=3"`
	Password string `json:"password" db:"password" validate:"required,min=6"`
	Email    string `json:"email" db:"email" validate:"required,email"`
}

type User struct {
	Name       string     `json:"name" db:"name"`
	Password   string     `json:"password" db:"password"`
	ID         string     `json:"id" db:"id"`
	Email      string     `json:"email" db:"email"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	ArchivedAt *time.Time `json:"archived_at" db:"archived_at"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type Todo struct {
	ID          string    `json:"id" db:"id"`
	Title       string    `json:"title" db:"title"`
	Status      string    `json:"status" db:"status"`
	Description string    `json:"description" db:"description"`
	Deadline    time.Time `json:"deadline" db:"deadline"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type UserExist struct {
	ID       string `db:"id"`
	Password string `db:"password"`
}
