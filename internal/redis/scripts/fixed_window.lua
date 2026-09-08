local key = KEYS[1]

local limit = tonumber(ARGV[1])
local window_seconds = tonumber(ARGV[2])

if not limit or not window_seconds then
    return redis.error_reply("invalid fixed window arguments")
end

if limit <= 0 or window_seconds <= 0 then
    return redis.error_reply(
        "fixed window arguments must be greater than zero"
    )
end

local current_count = tonumber(redis.call("GET", key))

if not current_count then
    current_count = 0
end

local allowed = 0
local remaining = 0

if current_count < limit then
    current_count = current_count + 1
    allowed = 1
    remaining = limit - current_count
else
    remaining = 0
end

if allowed == 1 then
    redis.call(
        "SET",
        key,
        current_count,
        "EX",
        window_seconds
    )
end

local ttl_ms = redis.call("PTTL", key)

local retry_after_ms = 0

if allowed == 0 then
    retry_after_ms = ttl_ms
end

return {
    allowed,
    remaining,
    limit,
    retry_after_ms
}