package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Shubhouy1/todo-app/database/dbhelper"
	"github.com/Shubhouy1/todo-app/middleware"
	"github.com/Shubhouy1/todo-app/model"
	"github.com/Shubhouy1/todo-app/util"
	"github.com/go-chi/chi/v5"
)

func CreateTodo(w http.ResponseWriter, r *http.Request) {
	auth := middleware.GetAuthContext(r)
	userID := auth.UserID

	var todo model.Todo
	if err := util.ParseBody(r, &todo); err != nil {
		util.RespondError(w, http.StatusBadRequest, err, "invalid body")
		return
	}

	if err := validate.Struct(todo); err != nil {
		util.RespondError(w, http.StatusBadRequest, err, "failed to validate request body")
		return
	}

	err := dbhelper.CreateTodo(userID, todo.Title, todo.Status, todo.Description, todo.Deadline)
	if err != nil {
		util.RespondError(w, http.StatusInternalServerError, err, "failed to create todo")
		return
	}

	// status created
	util.RespondJSON(w, http.StatusCreated, "todo created successfully")
}

func UpdateTodo(w http.ResponseWriter, r *http.Request) {
	todoID := chi.URLParam(r, "id")
	auth := middleware.GetAuthContext(r)
	userID := auth.UserID

	var todo model.Todo
	if err := util.ParseBody(r, &todo); err != nil {
		util.RespondError(w, http.StatusBadRequest, err, "invalid body")
		return
	}

	if err := validate.Struct(todo); err != nil {
		util.RespondError(w, http.StatusBadRequest, err, "failed to validate request body")
		return
	}

	err := dbhelper.UpdateTodoData(
		userID,
		todoID,
		todo.Title,
		todo.Status,
		todo.Description,
		todo.Deadline,
	)
	if err != nil {
		util.RespondError(w, http.StatusInternalServerError, err, "failed to update todo")
		return
	}

	util.RespondJSON(w, http.StatusOK, "updated successfully")
}

func DeleteTodo(w http.ResponseWriter, r *http.Request) {
	todoID := chi.URLParam(r, "id")
	auth := middleware.GetAuthContext(r)
	userID := auth.UserID

	err := dbhelper.DeleteTodo(userID, todoID)
	if err != nil {
		util.RespondError(w, http.StatusNotFound, err, "todo not found")
		return
	}

	util.RespondJSON(w, http.StatusOK, "deleted successfully")
}

func UpdateTodoStatus(w http.ResponseWriter, r *http.Request) {
	todoID := chi.URLParam(r, "id")
	auth := middleware.GetAuthContext(r)
	userID := auth.UserID
	var body struct {
		Status string `json:"status" validate:"required,oneof=Completed 'Not Completed' Pending"`
	}

	if err := util.ParseBody(r, &body); err != nil {
		util.RespondError(w, http.StatusBadRequest, err, "invalid body")
		return
	}

	if err := validate.Struct(body); err != nil {
		util.RespondError(w, http.StatusBadRequest, err, "failed to validate request body")
		return
	}

	deadline, err := dbhelper.GetDeadline(todoID, userID)
	if err != nil {
		util.RespondError(w, http.StatusNotFound, err, "todo not found")
		return
	}

	if time.Now().After(deadline) && body.Status == "Completed" {
		util.RespondError(w, http.StatusForbidden, nil, "cannot mark completed after deadline")
		return
	}

	err = dbhelper.UpdateStatus(todoID, userID, body.Status)
	if err != nil {
		util.RespondError(w, http.StatusInternalServerError, err, "failed to update status")
		return
	}

	util.RespondJSON(w, http.StatusOK, "updated successfully")
}

func GetTodoByID(w http.ResponseWriter, r *http.Request) {
	todoID := chi.URLParam(r, "id")
	auth := middleware.GetAuthContext(r)
	userID := auth.UserID

	todo, err := dbhelper.GetTodoByID(todoID, userID)
	if err != nil {
		util.RespondError(w, http.StatusNotFound, err, "todo not found")
		return
	}
	if todo == nil {
		util.RespondJSON(w, http.StatusOK, map[string]interface{}{
			"todo": nil,
		})
		return
	}

	util.RespondJSON(w, http.StatusOK, todo)
}

// GetTodos remove id filter and make separate api for GetTodobyID
func GetTodos(w http.ResponseWriter, r *http.Request) {
	auth := middleware.GetAuthContext(r)
	userID := auth.UserID

	status := r.URL.Query().Get("status")
	daysStr := r.URL.Query().Get("days")

	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	page := 1
	limit := 10

	if pageStr != "" {
		p, err := strconv.Atoi(pageStr)
		if err != nil || p <= 0 {
			util.RespondError(w, http.StatusBadRequest, nil, "invalid page")
			return
		}
		page = p
	}

	if limitStr != "" {
		l, err := strconv.Atoi(limitStr)
		if err != nil || l <= 0 || l > 100 {
			util.RespondError(w, http.StatusBadRequest, nil, "invalid limit")
			return
		}
		limit = l
	}

	offset := (page - 1) * limit

	var selectedDate *time.Time
	if daysStr != "" {
		d, err := strconv.Atoi(daysStr)
		if err != nil || d <= 0 {
			util.RespondError(w, http.StatusBadRequest, nil, "invalid days")
			return
		}

		t := time.Now().AddDate(0, 0, d)
		selectedDate = &t
	}

	todos, err := dbhelper.GetTodos(
		userID,
		status,
		selectedDate,
		limit,
		offset,
	)
	if err != nil {
		util.RespondError(w, http.StatusInternalServerError, err, "failed to fetch todos")
		return
	}

	util.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"page":  page,
		"limit": limit,
		"data":  todos,
	})
}
