package routes

import (
	"time"
	"os"
	"github.com/gofiber/fiber"
)

type request struct {
	URL         string        `json:"url"`
	CustomShort string        `json:"short"`
	Expiry      time.Duration `json:"expiry"`
}
type response struct {
	URL             string        `json:"url"`
	CustomShortened string        `json:"short"`
	Expiry          time.Duration `json:"expiry"`
	XRateRemaining  int           `json:"short"`
	XRateLimitReset time.Duration `json:"rate_limit_reset"`
}

func ShortenURL(c *fiber.Ctx) error {
	body := new(request)

	if err := c.BodyParser(body); err != nil {
		return c.Status(fiber.StatusBadRequest).
			JSON(fiber.Map{"error": err.Error()})
	}

	// rate limiting

	// check if url already exists
	if !govalidator.IsURL(body.URL) {
		return c.Status(fiber.StatusBadRequest).
			JSON(fiber.Map{"error": "Invalid URL"})
	}

	// check domain error
	if !helpers.RemoveDomainError(body.URL) {
		return c.Status(fiber.StatusBadRequest).
			JSON(fiber.Map{"error": "Invalid URL"})
	}

	// enforce https

	body.URL = helpers.EnforceHTTP(body.URL)
}
