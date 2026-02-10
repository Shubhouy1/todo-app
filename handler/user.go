package handler

import (
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

	var sessionID int64

	txErr := database.Tx(func(tx *sqlx.Tx) error {
		userID, err := dbhelper.CreateUser(tx, body.Username, body.Email, string(hashPassword))
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

	util.RespondJSON(w, http.StatusCreated, map[string]int64{
		"sessionId": sessionID,
	})
}

func Login(w http.ResponseWriter, r *http.Request) {
	var body model.LoginRequest

	if err := util.ParseBody(r, &body); err != nil {
		util.RespondError(w, http.StatusBadRequest, err, "invalid request body")
		return
	}

	if err := validate.Struct(body); err != nil {
		util.RespondError(w, http.StatusBadRequest, err, "failed to validate request body")
		return
	}

	userID, err := dbhelper.GetUserByEmail(body.Email, body.Password)
	if err != nil {
		util.RespondError(w, http.StatusUnauthorized, nil, "invalid credentials")
		return
	}
	sessionID := util.GenerateSessionID()
	if err := dbhelper.CreateUserSessionOnLogin(userID, sessionID); err != nil {
		util.RespondError(w, http.StatusInternalServerError, nil, "failed to create session")
		return
	}

	util.RespondJSON(w, http.StatusOK, map[string]int64{
		"sessionId": sessionID,
	})
}

func Logout(w http.ResponseWriter, r *http.Request) {
	auth := middleware.GetAuthContext(r)
	sessionID := auth.SessionID

	if err := dbhelper.DeleteUserSession(sessionID); err != nil {
		util.RespondError(w, http.StatusInternalServerError, err, "logout failed")
		return
	}

	util.RespondJSON(w, http.StatusOK, "logout succeeded")
}
func GetUserDetail(w http.ResponseWriter, r *http.Request) {
	auth := middleware.GetAuthContext(r)
	userID := auth.UserID

	user, err := dbhelper.GetDetailByID(userID)
	if err != nil {
		util.RespondError(w, http.StatusNotFound, err, "user not found")
		return
	}

	util.RespondJSON(w, http.StatusOK, user)
}
func DeleteUser(w http.ResponseWriter, r *http.Request) {
	auth := middleware.GetAuthContext(r)

	userID := auth.UserID
	if userID == "" {
		util.RespondError(w, http.StatusUnauthorized, nil, "unauthorized")
		return
	}

	txErr := database.Tx(func(tx *sqlx.Tx) error {
		if err := dbhelper.DeleteUserSessionsByUser(tx, userID); err != nil {
			return err
		}
		if err := dbhelper.DeleteUser(tx, userID); err != nil {
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
