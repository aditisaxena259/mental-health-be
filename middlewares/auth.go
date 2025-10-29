package middlewares

import (
	"os"
	"strings"

	"github.com/aditisaxena259/mental-health-be/config"
	"github.com/aditisaxena259/mental-health-be/models"
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
		role, _ := c.Locals("role").(string)
		// Check basic allowed roles
		allowedMatch := false
		for _, allowed := range roles {
			if role == allowed {
				allowedMatch = true
				break
			}
		}
		if !allowedMatch {
			return c.Status(403).JSON(fiber.Map{"error": "Forbidden: insufficient privileges"})
		}

		// Additional hardening: if the role is a plain admin (not chief_admin), ensure the user has a Block assigned.
		if role == string(models.Admin) {
			uid, _ := c.Locals("user_id").(string)
			if uid == "" {
				return c.Status(403).JSON(fiber.Map{"error": "Forbidden: missing user id"})
			}
			var u models.User
			if err := config.DB.First(&u, "id = ?", uid).Error; err != nil {
				return c.Status(403).JSON(fiber.Map{"error": "Forbidden: cannot validate admin block"})
			}
			if strings.TrimSpace(u.Block) == "" {
				return c.Status(403).JSON(fiber.Map{"error": "Forbidden: admin not assigned to any block"})
			}
		}

		return c.Next()
	}
}
