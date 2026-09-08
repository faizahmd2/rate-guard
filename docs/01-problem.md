# RateGuard — Problem

## Problem

Microservices and backend services often need to call external APIs that have
their own rate limits.

If each service handles rate limiting independently, the same logic may be
implemented multiple times and can become difficult to maintain, especially
when multiple instances of a service are running.

RateGuard will provide a centralized service that can be called before making
a request to a protected API.

The caller asks RateGuard:

> "Can I make this request now?"

RateGuard responds with a decision such as:

- ALLOW
- DENY
- UNAVAILABLE

Rate limits will be configurable based on the resource being protected.

Examples:

- Xoxoday purchase API → 120 requests/minute
- Another external API → 100 requests/second
- Another internal resource → 500 requests/minute

## Who Uses RateGuard?

RateGuard is intended to be used by:

- Backend services
- Microservices
- Queue workers
- Cron jobs
- Other services that need to control request rates

## What Does RateGuard Solve?

RateGuard provides:

- Centralized rate-limit decisions
- Configurable rate limits
- Shared rate-limit state across multiple RateGuard instances
- Consistent rate limiting for distributed services
- Information about when a denied request can be retried

## What RateGuard Does NOT Do?

RateGuard does not make the actual external API request.

The caller is responsible for deciding what to do after receiving the
RateGuard decision.

For example:

ALLOW
→ Caller makes the external API request.

DENY
→ Caller can retry later, delay the request, or put it back into a queue.

UNAVAILABLE
→ Caller decides whether to fail, retry, queue, or use another strategy.