package model

import (
	"fmt"
	"strings"
	"time"
)

type Stats struct {
	RegisteredUsers      uint64  `json:"registered_users"`
	ProofsGenerated      uint64  `json:"proofs_generated"`
	ProversDeployed      uint64  `json:"programs"`
	ProofsVerified       uint64  `json:"proofs_verified"`
	RegisteredUsersDelta float64 `json:"registered_users_delta"`
	ProofsGeneratedDelta float64 `json:"proofs_generated_delta"`
	ProversDeployedDelta float64 `json:"programs_delta"`
	ProofsVerifiedDelta  float64 `json:"proofs_verified_delta"`
}

type Event struct {
	State     string    `json:"state"`
	TxID      string    `db:"tx_id" json:"tx_id"`
	ProverID  string    `db:"prover_id" json:"prover_id"`
	Tag       string    `json:"tag"`
	Timestamp time.Time `json:"timestamp"`
}

type TxInfo struct {
	State    string        `json:"state"`
	Duration time.Duration `json:"duration"`
	TxID     string        `json:"tx_id"`
	UserID   string        `json:"user_id"`
	ProverID string        `json:"prover_id"`
	Log      []TxLogEvent  `json:"log"`
}

type TxLogEvent struct {
	State     string    `json:"state"`
	IDType    string    `json:"id_type"`
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
}

type StatsRange interface {
	String() string
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
