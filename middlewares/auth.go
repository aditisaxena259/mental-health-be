package middlewares


import (
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func ProtectRoute(c *fiber.Ctx) error {
	auth := c.Get("Authorization")
	if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
		return c.Status(401).JSON(fiber.Map{"error": "Missing token"})
	}
	tokenStr := strings.TrimPrefix(auth, "Bearer ")

	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil || !token.Valid {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid or expired token"})
	}

	claims := token.Claims.(jwt.MapClaims)
	c.Locals("user_id", claims["user_id"])
	c.Locals("role", claims["role"])
	return c.Next()
}
func RequireRole(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		role := c.Locals("role")
		for _, allowed := range roles {
			if role == allowed {
				return c.Next()
			}
		}
		return c.Status(403).JSON(fiber.Map{"error": "Forbidden: insufficient privileges"})
	}
}
