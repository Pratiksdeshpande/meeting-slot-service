package repository

import (
	"context"
	"fmt"

	"meeting-slot-service/internal/database"
	"meeting-slot-service/internal/models"
)

type availabilityRepository struct {
	db *database.Database
}

// NewAvailabilityRepository creates a new availability repository
func NewAvailabilityRepository(db *database.Database) AvailabilityRepository {
	return &availabilityRepository{db: db}
}

func (r *availabilityRepository) CreateSlots(ctx context.Context, slots []models.AvailabilitySlot) error {
	if len(slots) == 0 {
		return nil
	}

	db, err := r.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	query := `INSERT INTO availability_slots (event_id, user_id, start_time, end_time, timezone, created_at, updated_at) 
			  VALUES (?, ?, ?, ?, ?, NOW(), NOW())`

	for _, slot := range slots {
		_, err := db.ExecContext(ctx, query, slot.EventID, slot.UserID, slot.StartTime, slot.EndTime, slot.Timezone)
		if err != nil {
			return fmt.Errorf("failed to create availability slot: %w", err)
		}
	}
	return nil
}

func (r *availabilityRepository) GetByEventAndUser(ctx context.Context, eventID, userID string) ([]models.AvailabilitySlot, error) {
	db, err := r.db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	query := `SELECT id, event_id, user_id, start_time, end_time, timezone, created_at, updated_at 
			  FROM availability_slots WHERE event_id = ? AND user_id = ? ORDER BY start_time ASC`

	rows, err := db.QueryContext(ctx, query, eventID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get availability slots: %w", err)
	}
	defer rows.Close()

	var slots []models.AvailabilitySlot
	for rows.Next() {
		var slot models.AvailabilitySlot
		if err := rows.Scan(&slot.ID, &slot.EventID, &slot.UserID, &slot.StartTime,
			&slot.EndTime, &slot.Timezone, &slot.CreatedAt, &slot.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan availability slot: %w", err)
		}
		slots = append(slots, slot)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return slots, nil
}

func (r *availabilityRepository) GetByEvent(ctx context.Context, eventID string) ([]models.AvailabilitySlot, error) {
	db, err := r.db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	query := `SELECT id, event_id, user_id, start_time, end_time, timezone, created_at, updated_at 
			  FROM availability_slots WHERE event_id = ? ORDER BY user_id ASC, start_time ASC`

	rows, err := db.QueryContext(ctx, query, eventID)
	if err != nil {
		return nil, fmt.Errorf("failed to get availability slots: %w", err)
	}
	defer rows.Close()

	var slots []models.AvailabilitySlot
	for rows.Next() {
		var slot models.AvailabilitySlot
		if err := rows.Scan(&slot.ID, &slot.EventID, &slot.UserID, &slot.StartTime,
			&slot.EndTime, &slot.Timezone, &slot.CreatedAt, &slot.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan availability slot: %w", err)
		}
		slots = append(slots, slot)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return slots, nil
}

func (r *availabilityRepository) UpdateUserSlots(ctx context.Context, eventID, userID string, slots []models.AvailabilitySlot) error {
	db, err := r.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	// Use transaction to delete old slots and create new ones
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete existing slots
	deleteQuery := `DELETE FROM availability_slots WHERE event_id = ? AND user_id = ?`
	if _, err := tx.ExecContext(ctx, deleteQuery, eventID, userID); err != nil {
		return fmt.Errorf("failed to delete old slots: %w", err)
	}

	// Create new slots
	if len(slots) > 0 {
		insertQuery := `INSERT INTO availability_slots (event_id, user_id, start_time, end_time, timezone, created_at, updated_at) 
					    VALUES (?, ?, ?, ?, ?, NOW(), NOW())`
		for _, slot := range slots {
			if _, err := tx.ExecContext(ctx, insertQuery, slot.EventID, slot.UserID,
				slot.StartTime, slot.EndTime, slot.Timezone); err != nil {
				return fmt.Errorf("failed to create new slot: %w", err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *availabilityRepository) DeleteUserSlots(ctx context.Context, eventID, userID string) error {
	db, err := r.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	query := `DELETE FROM availability_slots WHERE event_id = ? AND user_id = ?`
	_, err = db.ExecContext(ctx, query, eventID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete availability slots: %w", err)
	}
	return nil
}
