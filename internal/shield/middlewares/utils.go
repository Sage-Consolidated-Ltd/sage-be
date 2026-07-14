package middlewares

import (
	"github.com/google/uuid"
	"strconv"
)

func generateID() string {
	return uuid.NewString()
}
func itoa(n int) string {
	return strconv.Itoa(n)
}
