package llm

type LimitState string
type Tokens = int64

const (
	LimitStateNormal   LimitState = "normal"
	LimitStateSoftStop LimitState = "soft_stop"
	LimitStateMaxStop  LimitState = "max_stop"
)

type Context struct {
	Prompt     Tokens
	Completion Tokens
}

func (c Context) Tokens() Tokens {
	total := c.Prompt + c.Completion
	if total < 0 {
		return 0
	}

	return total
}

type Snapshot struct {
	UsedTokens Tokens
	State      LimitState
}
