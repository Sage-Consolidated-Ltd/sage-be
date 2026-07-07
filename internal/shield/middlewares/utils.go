package middlewares

import (
	"strconv"
	"github.com/google/uuid"
)

func generateID() string {
	return uuid.NewString()
}
func itoa(n int) string {
	return strconv.Itoa(n)
}
