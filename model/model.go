package model

import "time"

type Stats struct {
	RegisteredUsers uint64 `json:"registered_users"`
	ProofsGenerated uint64 `json:"proofs_generated"`
	ProversDeployed uint64 `json:"programs"`
	ProofsVerified  uint64 `json:"proofs_verified"`
}

type Event struct {
	State     string    `json:"state"`
	TxID      string    `db:"tx_id" json:"tx_id"`
	ProverID  string    `db:"prover_id" json:"prover_id"`
	Tag       string    `json:"tag"`
	Timestamp time.Time `json:"timestamp"`
}
