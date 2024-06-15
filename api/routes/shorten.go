package routes

import (
	"os"
	"time"
	"strconv"
	"url-shortener/helpers"
	"github.com/asaskevich/govalidator"
	"github.com/google/uuid"
	"https://github.com/kshitijk4poor/golang-url-shortener/database"
	"https://github.com/kshitijk4poor/golang-url-shortener/helpers"
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

	var id string

	if body.CustomShort == "" {
		id = uuid.New().String()[:8]
	} else {
		id = body.CustomShort
	}

	r := database.CreateClient(0)
	defer r.Close() 

	val, _ = r.Get(database.Ctx, id).Result()
	if val != "" {
		return c.Status(fiber.StatusForbidden).
			JSON(fiber.Map{"error": "URL custom short already in use"})
	}
	if body.Expiry == 0 {
		body.Expiry = 24 * time.Hour	
	}

	err = r.Set(database.Ctx, id, body.URL, body)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).
			JSON(fiber.Map{"error": "Unable to connect to DB"})
	}

	resp := response{
		URL:             body.URL,
		CustomShortened: id,
		Expiry:          body.Expiry,
		XRateRemaining:  10,
		XRateLimitReset: 30
	}
	r2.Decr(database.Ctx, c.IP())

	val,_ =  r2.Get(database.Ctx, c.IP()).Result()
	Resp.rateRemaining, _ = strconv.Atoi(val)

	ttl, _ := r2.TTL(database.Ctx, C.IP()).Result()
	resp.XrateLimitReset = time.Duration(ttl) / time.Minute

	resp.CustomShort = os.Getenv("DOMAIN") + "/" + id

	return c.Status(fiber.StatusCreated).JSON(resp)
}
