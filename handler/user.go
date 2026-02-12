package handler

import (
	"fmt"
	"net/http"

	"github.com/Shubhouy1/todo-app/database"
	"github.com/Shubhouy1/todo-app/database/dbhelper"
	"github.com/Shubhouy1/todo-app/middleware"
	"github.com/Shubhouy1/todo-app/model"
	"github.com/Shubhouy1/todo-app/util"
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
)

var validate = validator.New()

func RegisterUser(w http.ResponseWriter, r *http.Request) {
	var body model.UserRequest
	var sessionID int64
	var userID string

	if err := util.ParseBody(r, &body); err != nil {
		util.RespondError(w, http.StatusBadRequest, err, "failed to parse request body")
		return
	}

	if err := validate.Struct(body); err != nil {
		util.RespondError(w, http.StatusBadRequest, err, "failed to validate request body")
		return
	}

	exist, err := dbhelper.IsUserExist(body.Email)
	if err != nil {
		util.RespondError(w, http.StatusInternalServerError, err, "database error")
		return
	}

	if exist {
		util.RespondError(w, http.StatusConflict, nil, "user already exists")
		return
	}

	hashPassword, err := util.HashPassword(body.Password)
	if err != nil {
		util.RespondError(w, http.StatusInternalServerError, err, "password hashing failed")
		return
	}

	txErr := database.Tx(func(tx *sqlx.Tx) error {
		var err error

		userID, err = dbhelper.CreateUser(tx, body.Username, body.Email, string(hashPassword))
		if err != nil {
			return err
		}

		sessionID = util.GenerateSessionID()

		if err := dbhelper.CreateUserSession(tx, userID, sessionID); err != nil {
			return err
		}

		return nil
	})

	if txErr != nil {
		util.RespondError(w, http.StatusInternalServerError, txErr, "failed to register user")
		return
	}
	token, err := util.GenerateJWT(userID, fmt.Sprintf("%d", sessionID))
	if err != nil {
		util.RespondError(w, http.StatusInternalServerError, err, "failed to generate token")
		return
	}

	util.RespondJSON(w, http.StatusCreated, map[string]interface{}{
		"sessionId":   sessionID,
		"accessToken": token,
	})
}

func Login(w http.ResponseWriter, r *http.Request) {
	var body model.LoginRequest
	var sessionID int64
	var userID string
	if err := util.ParseBody(r, &body); err != nil {
		util.RespondError(w, http.StatusBadRequest, err, "invalid request body")
		return
	}

	if err := validate.Struct(body); err != nil {
		util.RespondError(w, http.StatusBadRequest, err, "failed to validate request body")
		return
	}
	txErr := database.Tx(func(tx *sqlx.Tx) error {
		var err error
		userID, err = dbhelper.GetUserByEmail(tx, body.Email, body.Password)
		if err != nil {
			return err
		}
		sessionID = util.GenerateSessionID()
		if err := dbhelper.CreateUserSession(tx, userID, sessionID); err != nil {
			return err
		}
		return nil
	})
	if txErr != nil {
		util.RespondError(w, http.StatusUnauthorized, txErr, "invalid credentials")
		return
	}
	token, err := util.GenerateJWT(userID, fmt.Sprintf("%d", sessionID))
	if err != nil {
		util.RespondError(w, http.StatusInternalServerError, err, "failed to generate token")
		return
	}

	util.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"sessionId":   sessionID,
		"accessToken": token,
	})
}
func Logout(w http.ResponseWriter, r *http.Request) {
	auth, ok := middleware.GetAuthContext(r)
	if !ok {
		util.RespondError(w, http.StatusUnauthorized, nil, "unauthorized")
		return
	}

	sessionID := auth.SessionID

	if err := dbhelper.DeleteUserSession(sessionID); err != nil {
		util.RespondError(w, http.StatusInternalServerError, err, "logout failed")
		return
	}

	util.RespondJSON(w, http.StatusOK, "logout succeeded")
}
func GetUserDetail(w http.ResponseWriter, r *http.Request) {
	auth, ok := middleware.GetAuthContext(r)
	if !ok {
		util.RespondError(w, http.StatusUnauthorized, nil, "unauthorized")
		return
	}

	userID := auth.UserID

	user, err := dbhelper.GetDetailByID(userID)
	if err != nil {
		util.RespondError(w, http.StatusNotFound, err, "user not found")
		return
	}

	util.RespondJSON(w, http.StatusOK, user)
}
func DeleteUser(w http.ResponseWriter, r *http.Request) {
	auth, ok := middleware.GetAuthContext(r)
	if !ok {
		util.RespondError(w, http.StatusUnauthorized, nil, "unauthorized")
		return
	}

	userID := auth.UserID
	if userID == "" {
		util.RespondError(w, http.StatusUnauthorized, nil, "unauthorized")
		return
	}

	txErr := database.Tx(func(tx *sqlx.Tx) error {
		if err := dbhelper.DeleteUserSessionsByUserID(tx, userID); err != nil {
			return err
		}
		if err := dbhelper.DeleteUser(tx, userID); err != nil {
			return err
		}
		if err := dbhelper.DeleteAllTodos(tx, userID); err != nil {
			return err
		}

		return nil
	})

	if txErr != nil {
		util.RespondError(w, http.StatusInternalServerError, txErr, "failed to delete user")
		return
	}

	util.RespondJSON(w, http.StatusOK, map[string]string{
		"message": "user deleted successfully",
	})
}
