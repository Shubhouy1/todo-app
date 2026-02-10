package dbhelper

import (
	"database/sql"
	"time"

	"github.com/Shubhouy1/todo-app/database"
	"github.com/Shubhouy1/todo-app/model"
	"github.com/jmoiron/sqlx"
)

func CreateTodo(userId, title, status, description string, deadline time.Time) error {
	query := `
		INSERT INTO todos (user_id, title, status,description,deadline)
		VALUES ($1, $2, $3,$4, $5)
	`
	_, err := database.Todo.Exec(query, userId, title, status, description, deadline)
	if err != nil {
		return err
	}
	return nil
}

func UpdateTodoData(userID, todoID, Title, Status, Description string, Deadline time.Time) error {
	query := `
		UPDATE todos
		SET title = $1, status = $2, description = $3, deadline = $4
		WHERE id = $5 AND user_id = $6
	`

	_, err := database.Todo.Exec(query, Title, Status, Description, Deadline, todoID, userID)

	if err != nil {
		return err
	}
	return nil
}

// GetTodoByID
func GetTodoByID(todoID, userID string) (*model.Todo, error) {
	var todo model.Todo

	query := `
		SELECT id, title, status, description, deadline
		FROM todos
		WHERE id = $1
		  AND user_id = $2
		  AND archived_at IS NULL
	`

	err := database.Todo.Get(&todo, query, todoID, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &todo, nil
}

func GetDeadline(todoID, userID string) (time.Time, error) {
	var deadline time.Time

	queryDeadline := `
		SELECT deadline
		FROM todos
		WHERE id = $1 AND user_id = $2 AND archived_at IS NULL`

	err := database.Todo.Get(&deadline, queryDeadline, todoID, userID)
	if err != nil {
		return time.Time{}, err
	}
	return deadline, nil

}
func UpdateStatus(todoID, userID, status string) error {
	queryUpdate := `
		UPDATE todos
		SET status = $1
		WHERE id = $2 AND user_id = $3
	`

	_, err := database.Todo.Exec(queryUpdate, status, todoID, userID)
	return err
}
func DeleteTodo(userID, todoID string) error {
	query := `
		UPDATE todos
		SET archived_at = NOW()
		WHERE id = $1
		  AND user_id = $2
		  AND archived_at IS NULL
	`

	_, err := database.Todo.Exec(query, todoID, userID)
	if err != nil {
		return err
	}
	return nil
}

func GetTodos(
	userID string,
	status string,
	selectedDate *time.Time,
	limit int,
	offset int,
) ([]model.Todo, error) {

	query := `
		SELECT id, title, description, status, deadline, created_at
		FROM todos
		WHERE user_id = $1
		  AND archived_at IS NULL
		  AND (
			  $2 = '' OR status = $2::status
		  )
		  AND (
			  $3::timestamp IS NULL OR deadline <= $3::timestamp
		  )
		ORDER BY created_at DESC
		LIMIT $4 OFFSET $5
	`

	todos := []model.Todo{}

	err := database.Todo.Select(
		&todos,
		query,
		userID,
		status,
		selectedDate,
		limit,
		offset,
	)

	return todos, err
}

func DeleteAllTodos(tx *sqlx.Tx, userID string) error {
	query := `UPDATE todos 
            SET archived_at = NOW()
            WHERE user_id = $1
            AND archived_at IS NULL`

	_, err := tx.Exec(query, userID)
	if err != nil {
		return err
	}
	return nil
}
