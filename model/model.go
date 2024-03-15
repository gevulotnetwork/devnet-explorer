package model

import "time"

type Stats struct {
	RegisteredUsers int64 `json:"registered_users"`
	ProofsGenerated int64 `json:"proofs_generated"`
	ProversDeployed int64 `json:"programs"`
	ProofsVerified  int64 `json:"proofs_verified"`
}

type Event struct {
	State     string    `json:"state"`
	TxID      string    `json:"tx_id"`
	ProverID  string    `json:"prover_id"`
	Timestamp time.Time `json:"timestamp"`
}
