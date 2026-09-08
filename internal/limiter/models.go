package limiter

type CheckInput struct {
	Service  string
	Resource string
	Key      string
}

type DecisionType string

const (
	DecisionAllow DecisionType = "ALLOW"
	DecisionDeny  DecisionType = "DENY"
)

type Decision struct {
	Decision     DecisionType `json:"decision"`
	Limit        int          `json:"limit"`
	Remaining    int          `json:"remaining"`
	RetryAfterMs int64        `json:"retry_after_ms"`
}
