local key = KEYS[1]

local limit = tonumber(ARGV[1])
local window_seconds = tonumber(ARGV[2])

if not limit or not window_seconds then
    return redis.error_reply("invalid sliding window arguments")
end

if limit <= 0 or window_seconds <= 0 then
    return redis.error_reply(
        "sliding window arguments must be greater than zero"
    )
end

local redis_time = redis.call("TIME")

local now_seconds =
    tonumber(redis_time[1]) +
    tonumber(redis_time[2]) / 1000000

local window_start =
    now_seconds - window_seconds

-- Remove requests outside the sliding window.
redis.call(
    "ZREMRANGEBYSCORE",
    key,
    0,
    window_start
)

local current_count =
    redis.call("ZCARD", key)

local allowed = 0
local remaining = 0
local retry_after_ms = 0

if current_count < limit then

    local request_id =
        redis_time[1] ..
        "-" ..
        redis_time[2] ..
        "-" ..
        math.random(1000000)

    redis.call(
        "ZADD",
        key,
        now_seconds,
        request_id
    )

    current_count = current_count + 1

    allowed = 1
    remaining = limit - current_count

else

    remaining = 0

    local oldest = redis.call(
        "ZRANGE",
        key,
        0,
        0,
        "WITHSCORES"
    )

    if oldest and #oldest >= 2 then
        local oldest_timestamp =
            tonumber(oldest[2])

        retry_after_ms = math.ceil(
            (oldest_timestamp + window_seconds - now_seconds)
            * 1000
        )

        if retry_after_ms < 0 then
            retry_after_ms = 0
        end
    end
end

-- Keep Redis memory bounded.
redis.call(
    "EXPIRE",
    key,
    window_seconds + 1
)

return {
    allowed,
    remaining,
    limit,
    retry_after_ms
}