package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"meeting-slot-service/internal/database"
	"meeting-slot-service/internal/models"
)

type eventRepository struct {
	db *database.Database
}

// NewEventRepository creates a new event repository
func NewEventRepository(db *database.Database) EventRepository {
	return &eventRepository{db: db}
}

func (r *eventRepository) Create(ctx context.Context, event *models.Event) error {
	db, err := r.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	now := time.Now()
	query := `INSERT INTO events (id, title, description, organizer_id, duration_minutes, status, created_at, updated_at) 
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err = db.ExecContext(ctx, query, event.ID, event.Title, event.Description,
		event.OrganizerID, event.DurationMinutes, event.Status, now, now)
	if err != nil {
		return fmt.Errorf("failed to create event: %w", err)
	}

	// Set timestamps on the event object
	event.CreatedAt = now
	event.UpdatedAt = now

	// Insert proposed slots
	if len(event.ProposedSlots) > 0 {
		slotQuery := `INSERT INTO proposed_slots (event_id, start_time, end_time, timezone, created_at) 
					  VALUES (?, ?, ?, ?, ?)`
		for i := range event.ProposedSlots {
			slotNow := time.Now()
			result, err := db.ExecContext(ctx, slotQuery, event.ID, event.ProposedSlots[i].StartTime,
				event.ProposedSlots[i].EndTime, event.ProposedSlots[i].Timezone, slotNow)
			if err != nil {
				return fmt.Errorf("failed to create proposed slot: %w", err)
			}
			// Get the auto-generated slot ID and set it on the slot object
			slotID, err := result.LastInsertId()
			if err != nil {
				return fmt.Errorf("failed to get slot ID: %w", err)
			}
			event.ProposedSlots[i].ID = uint(slotID)
			event.ProposedSlots[i].EventID = event.ID
			event.ProposedSlots[i].CreatedAt = slotNow
		}
	}

	return nil
}

func (r *eventRepository) GetByID(ctx context.Context, id string) (*models.Event, error) {
	db, err := r.db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}
	var event models.Event
	query := `SELECT id, title, description, organizer_id, duration_minutes, status, created_at, updated_at 
			  FROM events WHERE id = ? AND deleted_at IS NULL`
	err = db.QueryRowContext(ctx, query, id).Scan(
		&event.ID, &event.Title, &event.Description, &event.OrganizerID,
		&event.DurationMinutes, &event.Status, &event.CreatedAt, &event.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("event not found")
		}
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	if err := r.loadRelated(ctx, db, &event); err != nil {
		return nil, err
	}

	return &event, nil
}

func (r *eventRepository) Update(ctx context.Context, event *models.Event) error {
	db, err := r.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}
	query := `UPDATE events SET title = ?, description = ?, duration_minutes = ?, status = ?, updated_at = NOW() 
			  WHERE id = ? AND deleted_at IS NULL`
	result, err := db.ExecContext(ctx, query, event.Title, event.Description,
		event.DurationMinutes, event.Status, event.ID)
	if err != nil {
		return fmt.Errorf("failed to update event: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("event not found")
	}

	// Update proposed slots if provided
	if len(event.ProposedSlots) > 0 {
		// Delete existing proposed slots
		_, err = db.ExecContext(ctx, `DELETE FROM proposed_slots WHERE event_id = ?`, event.ID)
		if err != nil {
			return fmt.Errorf("failed to delete existing proposed slots: %w", err)
		}

		// Insert new proposed slots
		slotQuery := `INSERT INTO proposed_slots (event_id, start_time, end_time, timezone, created_at) 
					  VALUES (?, ?, ?, ?, NOW())`
		for _, slot := range event.ProposedSlots {
			_, err = db.ExecContext(ctx, slotQuery, event.ID, slot.StartTime, slot.EndTime, slot.Timezone)
			if err != nil {
				return fmt.Errorf("failed to create proposed slot: %w", err)
			}
		}
	}

	return nil
}

func (r *eventRepository) Delete(ctx context.Context, id string) error {
	db, err := r.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}
	// Soft delete
	query := `UPDATE events SET deleted_at = NOW() WHERE id = ? AND deleted_at IS NULL`
	result, err := db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete event: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("event not found")
	}
	return nil
}

func (r *eventRepository) List(ctx context.Context, filter models.EventFilter) ([]*models.Event, int, error) {
	db, err := r.db.DB()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get database connection: %w", err)
	}
	// Build the query with filters
	countQuery := `SELECT COUNT(*) FROM events WHERE deleted_at IS NULL`
	query := `SELECT id, title, description, organizer_id, duration_minutes, status, created_at, updated_at 
			  FROM events WHERE deleted_at IS NULL`

	var args []interface{}

	if filter.OrganizerID != "" {
		countQuery += " AND organizer_id = ?"
		query += " AND organizer_id = ?"
		args = append(args, filter.OrganizerID)
	}
	if filter.Status != "" {
		countQuery += " AND status = ?"
		query += " AND status = ?"
		args = append(args, filter.Status)
	}

	// Get total count
	var total int
	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)
	if err := db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count events: %w", err)
	}

	// Apply pagination
	limit := filter.Limit
	if limit == 0 {
		limit = 20
	}
	offset := (filter.Page - 1) * limit
	if offset < 0 {
		offset = 0
	}

	query += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list events: %w", err)
	}
	defer rows.Close()

	events := make([]*models.Event, 0)
	for rows.Next() {
		var event models.Event
		if err := rows.Scan(&event.ID, &event.Title, &event.Description, &event.OrganizerID,
			&event.DurationMinutes, &event.Status, &event.CreatedAt, &event.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan event: %w", err)
		}
		events = append(events, &event)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating rows: %w", err)
	}

	// Enrich each event with its proposed slots and participants
	for _, event := range events {
		if err := r.loadRelated(ctx, db, event); err != nil {
			return nil, 0, err
		}
	}

	return events, total, nil
}

// loadRelated fetches proposed slots and participants (with user info) for a
// single event and attaches them to the event struct.
func (r *eventRepository) loadRelated(ctx context.Context, db interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}, event *models.Event) error {
	// Proposed slots
	slotsQuery := `SELECT id, event_id, start_time, end_time, timezone, created_at
				   FROM proposed_slots WHERE event_id = ?`
	sRows, err := db.QueryContext(ctx, slotsQuery, event.ID)
	if err != nil {
		return fmt.Errorf("failed to get proposed slots: %w", err)
	}
	defer sRows.Close()

	for sRows.Next() {
		var slot models.ProposedSlot
		if err := sRows.Scan(&slot.ID, &slot.EventID, &slot.StartTime, &slot.EndTime,
			&slot.Timezone, &slot.CreatedAt); err != nil {
			return fmt.Errorf("failed to scan proposed slot: %w", err)
		}
		event.ProposedSlots = append(event.ProposedSlots, slot)
	}

	// Participants with user info
	participantsQuery := `SELECT ep.id, ep.event_id, ep.user_id, ep.status,
						  ep.created_at, ep.updated_at,
						  u.id, u.name, u.email, u.created_at, u.updated_at
						  FROM event_participants ep
						  LEFT JOIN users u ON ep.user_id = u.id
						  WHERE ep.event_id = ?`
	pRows, err := db.QueryContext(ctx, participantsQuery, event.ID)
	if err != nil {
		return fmt.Errorf("failed to get participants: %w", err)
	}
	defer pRows.Close()

	for pRows.Next() {
		var p models.EventParticipant
		var user models.User
		if err := pRows.Scan(&p.ID, &p.EventID, &p.UserID, &p.Status,
			&p.CreatedAt, &p.UpdatedAt,
			&user.ID, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return fmt.Errorf("failed to scan participant: %w", err)
		}
		p.User = &user
		event.Participants = append(event.Participants, p)
	}

	return nil
}
