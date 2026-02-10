package dbhelper

import (
	"github.com/Shubhouy1/todo-app/database"
	"github.com/Shubhouy1/todo-app/model"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

func IsUserExist(email string) (bool, error) {
	query := `
		SELECT COUNT(*) > 0
		FROM users
		WHERE TRIM(LOWER(email)) = TRIM(LOWER($1))
		  AND archived_at IS NULL
	`

	var exist bool
	err := database.Todo.Get(&exist, query, email)
	return exist, err
}

func CreateUser(tx *sqlx.Tx, name, email, password string) (string, error) {
	query := `
		INSERT INTO users (name, email, password)
		VALUES ($1, TRIM(LOWER($2)), $3)
		RETURNING id
	`

	var userID string
	err := tx.Get(&userID, query, name, email, password)
	if err != nil {
		return "", err
	}

	return userID, nil
}

func CreateUserSession(tx *sqlx.Tx, userID string, sessionID int64) error {
	query := `
		INSERT INTO user_sessions (session_id, user_id, expires_at)
		VALUES ($1, $2, NOW() + INTERVAL '1 day')
	`
	_, err := tx.Exec(query, sessionID, userID)
	return err
}
func CreateUserSessionOnLogin(userId string, sessionID int64) error {
	query := `
	INSERT INTO user_sessions (session_id, user_id, expires_at)
	VALUES ($1, $2, NOW() + INTERVAL '1 day')
	`
	_, err := database.Todo.Exec(query, sessionID, userId)
	if err != nil {
		return err
	}
	return nil
}

func GetUserByEmail(tx *sqlx.Tx, email, password string) (string, error) {
	query := `
		SELECT id, password
		FROM users
		WHERE TRIM(LOWER(email)) = TRIM(LOWER($1))
		AND archived_at IS NULL
	`
	var result model.UserExist

	err := tx.Get(&result, query, email)
	if err != nil {
		return "", err
	}
	if err := bcrypt.CompareHashAndPassword(
		[]byte(result.Password),
		[]byte(password),
	); err != nil {
		return "", err
	}

	return result.ID, nil

}

func GetUserIDBySession(sessionID int64) (string, error) {
	query := `
		SELECT user_id
		FROM user_sessions
		WHERE session_id = $1
		AND archived_at IS NULL
	`
	var userID string
	err := database.Todo.Get(&userID, query, sessionID)
	if err != nil {
		return "", err
	}
	return userID, nil
}
func GetDetailByID(userID string) (model.User, error) {
	var user model.User

	query := `
		SELECT id, name, email, created_at
		FROM users
		WHERE id = $1
		AND archived_at IS NULL
    
`

	err := database.Todo.Get(&user, query, userID)
	return user, err
}
func DeleteUser(tx *sqlx.Tx, userID string) error {
	query := `
		UPDATE users
		SET archived_at = NOW()
		WHERE id = $1 AND archived_at IS NULL
	`
	_, err := tx.Exec(query, userID)
	return err
}

func DeleteUserSessionsByUser(tx *sqlx.Tx, userID string) error {
	query := `
		UPDATE user_sessions
		SET archived_at = NOW()
		WHERE user_id = $1 AND archived_at IS NULL
	`
	_, err := tx.Exec(query, userID)
	return err
}
func DeleteUserSession(sessionID int64) error {
	query := `UPDATE user_sessions SET archived_at = NOW() WHERE session_id = $1`
	_, err := database.Todo.Exec(query, sessionID)
	return err
}
