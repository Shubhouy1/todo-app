package middleware

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/Shubhouy1/todo-app/database/dbhelper"
	"github.com/Shubhouy1/todo-app/util"
)

type AuthContext struct {
	UserID    string
	SessionID int64
}

type contextKey string

const authKey contextKey = "authContext"

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionHeader := strings.TrimSpace(r.Header.Get("X-Session-ID"))
		if sessionHeader == "" {
			util.RespondError(w, http.StatusUnauthorized, nil, "session header is required")
			return
		}

		sessionID, err := strconv.ParseInt(sessionHeader, 10, 64)
		if err != nil {
			util.RespondError(w, http.StatusUnauthorized, nil, "invalid session header")
			return
		}

		userID, err := dbhelper.GetUserIDBySession(sessionID)
		if err != nil {
			util.RespondError(w, http.StatusUnauthorized, nil, "invalid session")
			return
		}

		authCtx := AuthContext{
			UserID:    userID,
			SessionID: sessionID,
		}

		ctx := context.WithValue(r.Context(), authKey, authCtx)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetAuthContext(r *http.Request) AuthContext {
	return r.Context().Value(authKey).(AuthContext)
}
