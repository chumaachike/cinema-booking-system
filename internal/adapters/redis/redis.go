package redis

import (
	"context"
	"log"

	goredis "github.com/redis/go-redis/v9"
)

func NewClient(addr string) *goredis.Client {
	rdb := goredis.NewClient(&goredis.Options{Addr: addr})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("redis ping: %v", err)
	}
	log.Printf("connected to redis at %s", addr)
	return rdb
}

var ConfirmScript = goredis.NewScript(`
    local seatKey = redis.call("GET", KEYS[1])

if not seatKey then
    return redis.error_reply("SESSION_EXPIRED")
end

local data = redis.call("GET", seatKey)

if not data then
    return redis.error_reply("HOLD_EXPIRED")
end

local booking = cjson.decode(data)

if booking.ID ~= ARGV[1] then
    return redis.error_reply("INVALID_SESSION")
end

if booking.Status ~= "held" then
    return redis.error_reply("INVALID_STATUS")
end

booking.Status = "confirmed"

redis.call("SET", seatKey, cjson.encode(booking))
redis.call("DEL", KEYS[1])

return cjson.encode(booking)
`)

var ReleaseScript = goredis.NewScript(`
    local seatKey = redis.call("GET", KEYS[1])

    if not seatKey then
        return 0
    end

    local data = redis.call("GET", seatKey)

    if not data then
        redis.call("DEL", KEYS[1])
        return 0
    end

    local booking = cjson.decode(data)

    if booking.ID ~= ARGV[1] then
        return -1
    end

    if booking.Status ~= "held" then
        return -2
    end

    redis.call("DEL", seatKey)
    redis.call("DEL", KEYS[1])

    return 1
`)
