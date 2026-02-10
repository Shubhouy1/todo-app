package router

import (
	"github.com/Shubhouy1/todo-app/handler"
	"github.com/Shubhouy1/todo-app/middleware"
	"github.com/go-chi/chi/v5"
)

func SetupRouter() chi.Router {
	r := chi.NewRouter()
	r.Post("/register", handler.RegisterUser)
	r.Post("/login", handler.Login)
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware)
		r.Post("/logout", handler.Logout)
		r.Post("/todo", handler.CreateTodo)
		r.Get("/get-details", handler.GetUserDetail)
		r.Get("/todos", handler.GetTodos)
		r.Get("/todos/{id}", handler.GetTodoByID)
		r.Put("/todos/{id}", handler.UpdateTodo)
		r.Patch("/todos/{id}", handler.UpdateTodoStatus)
		r.Delete("/todos/{id}", handler.DeleteTodo)
		r.Delete("/delete-user", handler.DeleteUser)
	})
	return r
}
