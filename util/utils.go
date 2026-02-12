package util

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/form3tech-oss/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

type Error struct {
	Error      string `json:"error,omitempty"`
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
}

func RespondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		return
	}
}
func RespondError(w http.ResponseWriter, statusCode int, err error, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	var errStr string
	if err != nil {
		errStr = err.Error()
	}

	resp := Error{
		Error:      errStr,
		StatusCode: statusCode,
		Message:    message,
	}

	if encodeErr := json.NewEncoder(w).Encode(resp); encodeErr != nil {
		fmt.Printf("failed to encode error response: %v\n", encodeErr)
	}
}

func ParseBody(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}

func GenerateSessionID() int64 {
	var b [8]byte
	_, err := rand.Read(b[:])
	if err != nil {
		panic(err)
	}
	return int64(binary.BigEndian.Uint64(b[:]))
}
func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func GenerateJWT(userID, sessionID string) (string, error) {
	claims := jwt.MapClaims{
		"userId":    userID,
		"sessionId": sessionID,
		"exp":       time.Now().Add(time.Minute * 10).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET_KEY")))
}
