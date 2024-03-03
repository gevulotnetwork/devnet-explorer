package model

type Stats struct {
	RegisteredUsers int64 `json:"registered_users"`
	ProofsGenerated int64 `json:"proofs_generated"`
	Programs        int64 `json:"programs"`
	ProofsVerified  int64 `json:"proofs_verified"`
}
