package routes

import (
	"os"
	"time"
	"url-shortener/helpers"

	"github.com.kshitijk4poor/url-shortener/database"
	"github.com/go-redis/redis"
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
	r2 := database.CreateClient(1)
	defer r2.Close()
	val, err := r2.Get(database.Ctx, c.IP()).Result()
	if err == redis.Nil {
		// set expiry to 12 hours
		_ = r2.Set(database.Ctx, c.IP(), os.Getenv("API_QUOTA"), time.Minute*30).Err()
	} else{
			val, _ = r2.Get(database.Ctx, c.IP()).Result()
			valInt, _ := strconv.Atoi(val)
			if valInt <= 0 {
				return c.Status(fiber.StatusTooManyRequests).
					JSON(fiber.Map{"error": "Rate limit reached"})
			}
		}
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).
				JSON(fiber.Map{"error": err.Error()})
		}
	}

	


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

	r2.Decr(database.Ctx, c.IP())

}
