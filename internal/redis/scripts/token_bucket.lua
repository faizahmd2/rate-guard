local key = KEYS[1]

local capacity = tonumber(ARGV[1])
local refill_rate = tonumber(ARGV[2])
local cost = tonumber(ARGV[3])

if not capacity or not refill_rate or not cost then
    return redis.error_reply("invalid token bucket arguments")
end

if capacity <= 0 or refill_rate <= 0 or cost <= 0 then
    return redis.error_reply("token bucket arguments must be greater than zero")
end

local redis_time = redis.call("TIME")
local now = tonumber(redis_time[1]) +
            tonumber(redis_time[2]) / 1000000

local tokens = tonumber(redis.call("HGET", key, "tokens"))
local last_refill = tonumber(redis.call("HGET", key, "last_refill"))

-- First request for this bucket.
if tokens == nil or last_refill == nil then
    tokens = capacity
    last_refill = now
end

-- Calculate elapsed time.
local elapsed = now - last_refill

-- Refill tokens.
tokens = math.min(
    capacity,
    tokens + (elapsed * refill_rate)
)

-- Update refill timestamp.
last_refill = now

local allowed = 0
local retry_after_ms = 0

if tokens >= cost then
    tokens = tokens - cost
    allowed = 1
else
    local missing_tokens = cost - tokens

    retry_after_ms = math.ceil(
        (missing_tokens / refill_rate) * 1000
    )
end

redis.call("HSET", key,
    "tokens", tokens,
    "last_refill", last_refill
)

-- Keep the bucket state from living forever.
local ttl = math.ceil(capacity / refill_rate)

if ttl < 1 then
    ttl = 1
end

redis.call("EXPIRE", key, ttl)

local remaining = math.floor(tokens)

return {
    allowed,
    remaining,
    math.floor(capacity),
    retry_after_ms
}