package middleware

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/Shubhouy1/todo-app/database/dbhelper"
	"github.com/Shubhouy1/todo-app/util"
	"github.com/form3tech-oss/jwt-go"
)

type AuthContext struct {
	UserID    string
	SessionID int64
}

type contextKey string

const authKey contextKey = "authContext"

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
		if authHeader == "" {
			util.RespondError(w, http.StatusUnauthorized, nil, "authorization header required")
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			util.RespondError(w, http.StatusUnauthorized, nil, "invalid authorization header")
			return
		}

		tokenStr := parts[1]

		secret := os.Getenv("JWT_SECRET_KEY")
		if secret == "" {
			util.RespondError(w, http.StatusInternalServerError, nil, "server configuration error")
			return
		}

		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("invalid signing method")
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			util.RespondError(w, http.StatusUnauthorized, nil, "invalid or expired token")
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			util.RespondError(w, http.StatusUnauthorized, nil, "invalid token claims")
			return
		}

		userID, ok := claims["userId"].(string)
		if !ok {
			util.RespondError(w, http.StatusUnauthorized, nil, "invalid token data")
			return
		}

		sessionStr, ok := claims["sessionId"].(string)
		if !ok {
			util.RespondError(w, http.StatusUnauthorized, nil, "invalid token data")
			return
		}

		sessionID, err := strconv.ParseInt(sessionStr, 10, 64)
		if err != nil {
			util.RespondError(w, http.StatusUnauthorized, nil, "invalid session")
			return
		}

		dbUserID, err := dbhelper.GetUserIDBySession(sessionID)
		if err != nil || dbUserID != userID {
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

func GetAuthContext(r *http.Request) (AuthContext, bool) {
	auth, ok := r.Context().Value(authKey).(AuthContext)
	return auth, ok
}
