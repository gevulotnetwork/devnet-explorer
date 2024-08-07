package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrNotFound = errors.New("not found")

type Stats struct {
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	RegisteredUsers uint64    `json:"registered_users" db:"registered_users"`
	ProofsGenerated uint64    `json:"proofs_generated" db:"proofs_generated"`
	ProversDeployed uint64    `json:"programs" db:"programs"`
	ProofsVerified  uint64    `json:"proofs_verified" db:"proofs_verified"`
}

type DeltaStats struct {
	RegisteredUsers float64 `json:"registered_users_delta" db:"registered_users_delta"`
	ProofsGenerated float64 `json:"proofs_generated_delta" db:"proofs_generated_delta"`
	ProversDeployed float64 `json:"programs_delta" db:"programs_delta"`
	ProofsVerified  float64 `json:"proofs_verified_delta" db:"proofs_verified_delta"`
}

type CombinedStats struct {
	Stats
	DeltaStats
}

type Event struct {
	State     State     `json:"state"`
	TxID      string    `db:"tx_id" json:"tx_id"`
	ProverID  string    `db:"prover_id" json:"prover_id"`
	Tag       string    `json:"tag"`
	Timestamp time.Time `json:"timestamp"`
}

type TxInfo struct {
	State    State         `json:"state"`
	Duration time.Duration `json:"duration"`
	TxID     string        `json:"tx_id"`
	UserID   string        `json:"user_id"`
	ProverID string        `json:"prover_id"`
	Log      []TxLogEvent  `json:"log"`
}

type TxLogEvent struct {
	State     State     `json:"state"`
	IDType    string    `json:"id_type"`
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
}

type StatsRange interface {
	String() string
	Since() time.Time
	sr()
}

type sr uint8

func (s sr) String() string {
	switch s {
	case RangeWeek:
		return "1w"
	case RangeMonth:
		return "1m"
	case RangeHalfYear:
		return "6m"
	case RangeYear:
		return "1y"
	default:
		return ""
	}
}

func (s sr) Since() time.Time {
	switch s {
	case RangeWeek:
		return time.Now().AddDate(0, 0, -7)
	case RangeMonth:
		return time.Now().AddDate(0, -1, 0)
	case RangeHalfYear:
		return time.Now().AddDate(0, -6, 0)
	case RangeYear:
		return time.Now().AddDate(-1, 0, 0)
	default:
		return time.Time{}
	}
}

func (s sr) sr() {}

const (
	RangeWeek     sr = 0
	RangeMonth    sr = 1
	RangeHalfYear sr = 2
	RangeYear     sr = 3
)

func ParseStatsRange(r string) (StatsRange, error) {
	switch strings.ToUpper(r) {
	case "1W":
		return RangeWeek, nil
	case "1M":
		return RangeMonth, nil
	case "6M":
		return RangeHalfYear, nil
	case "1Y":
		return RangeYear, nil
	default:
		return nil, fmt.Errorf("invalid StatsRange string")
	}
}

func SupportedStatsRanges() []StatsRange {
	return []StatsRange{RangeWeek, RangeMonth, RangeHalfYear, RangeYear}
}

type State uint8

const (
	StateUnknown   State = 0
	StateSubmitted State = 1
	StateProving   State = 2
	StateVerifying State = 3
	StateComplete  State = 4
)

func (s *State) String() string {
	switch *s {
	case StateSubmitted:
		return "submitted"
	case StateProving:
		return "proving"
	case StateVerifying:
		return "verifying"
	case StateComplete:
		return "complete"
	default:
		return "unknown"
	}
}

func (s *State) Scan(value interface{}) error {
	stateStr, ok := value.(string)
	if !ok {
		return fmt.Errorf("incompatible type for State: %T", value)
	}
	newState, err := ParseState(stateStr)
	if err != nil {
		return err
	}
	*s = newState
	return nil
}

func (s *State) Value() (driver.Value, error) {
	return s.String(), nil
}

func (s *State) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

func (s *State) UnmarshalJSON(data []byte) error {
	var stateStr string
	if err := json.Unmarshal(data, &stateStr); err != nil {
		return err
	}
	newState, err := ParseState(stateStr)
	if err != nil {
		return err
	}
	*s = newState
	return nil
}

func ParseState(r string) (State, error) {
	switch strings.ToLower(r) {
	case "submitted":
		return StateSubmitted, nil
	case "proving":
		return StateProving, nil
	case "verifying":
		return StateVerifying, nil
	case "complete":
		return StateComplete, nil
	default:
		return StateUnknown, fmt.Errorf("invalid State string: %s", r)
	}
}
