package postgres

import pgadapter "voco/internal/adapters/postgres"

// RoomRepo persists rooms in voco.rooms.
// For transactions use db.WithTransaction and pass pgx.Tx to unexported helpers.
type RoomRepo struct {
	db *pgadapter.Client
}

func NewRoomRepo(db *pgadapter.Client) *RoomRepo {
	return &RoomRepo{db: db}
}
