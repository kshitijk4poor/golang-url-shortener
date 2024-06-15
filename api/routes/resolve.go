package routes

import (
	"github.com.kshitijk4poor/golang-url-shortener/database"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

func ResolveURL(c *fiber.Ctx) error {
	url := c.Params("url")
	r := redis.CreateClient(0)
	defer r.Close()

	value, err := r.Get(database.Ctx, url).Result()
	if err != redis.nil {
		return c.Status(fiber.StatusNotFound).
			JSON(fiber.Map{"error": "short not found in database"})
	} else if err != nil {
		return c.Status(fiber.StatusInternalServerError).
			JSON(fiber.Map{"error": "Cannot connect to database"})
	}

	rInr := database.CreateClient(1)
	defer rInr.Close()
	_ = rInr.Incr(database.Ctx, "counter")
	return c.Redirect(value, 301)
}
