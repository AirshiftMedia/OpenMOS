package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"airshift/openmos/internal/model"
	"airshift/openmos/internal/repository"
)

type runningOrders struct{ db DBTX }

// roColumns are read by scanRunningOrder. ed_dur holds roEdDur; the model
// keeps it in whole seconds.
const roColumns = `ro_id, mos_id, slug, status,
	COALESCE(EXTRACT(EPOCH FROM ed_dur), 0)::bigint, ed_start,
	channel, metadata, version, created_by, created_at, updated_at`

func scanRunningOrder(row pgx.Row) (*model.RunningOrder, error) {
	var ro model.RunningOrder
	err := row.Scan(&ro.ID, &ro.MosID, &ro.Slug, &ro.Status, &ro.Duration, &ro.AirTime,
		&ro.Channel, &ro.Metadata, &ro.Version, &ro.CreatedBy, &ro.CreatedAt, &ro.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &ro, nil
}

func (s runningOrders) Create(ctx context.Context, ro *model.RunningOrder) (*model.RunningOrder, error) {
	if err := repository.CheckID("running order", ro.ID); err != nil {
		return nil, err
	}
	out, err := scanRunningOrder(s.db.QueryRow(ctx, `
		INSERT INTO openmos.running_orders
			(ro_id, mos_id, slug, status, ed_dur, ed_start, channel, metadata, version, created_by)
		VALUES ($1, $2, $3, $4, $5::bigint * interval '1 second', $6, $7, $8, 1, $9)
		RETURNING `+roColumns,
		ro.ID, ro.MosID, ro.Slug, string(ro.Status), ro.Duration, ro.AirTime, ro.Channel,
		metadata(ro.Metadata), ro.CreatedBy))
	if err != nil {
		return nil, dbError(err, "running order "+ro.ID)
	}
	ro.Version, ro.CreatedAt, ro.UpdatedAt = out.Version, out.CreatedAt, out.UpdatedAt
	return out, nil
}

func (s runningOrders) Get(ctx context.Context, roID string) (*model.RunningOrder, error) {
	ro, err := scanRunningOrder(s.db.QueryRow(ctx,
		"SELECT "+roColumns+" FROM openmos.running_orders WHERE ro_id = $1", roID))
	if err != nil {
		return nil, dbError(err, "running order "+roID)
	}
	return ro, nil
}

func (s runningOrders) Update(ctx context.Context, ro *model.RunningOrder) error {
	what := "running order " + ro.ID
	err := s.db.QueryRow(ctx, `
		UPDATE openmos.running_orders SET
			mos_id = $3, slug = $4, status = $5, ed_dur = $6::bigint * interval '1 second',
			ed_start = $7, channel = $8, metadata = $9, created_by = $10,
			version = version + 1, updated_at = now()
		WHERE ro_id = $1 AND version = $2
		RETURNING version, created_at, updated_at`,
		ro.ID, ro.Version, ro.MosID, ro.Slug, string(ro.Status), ro.Duration, ro.AirTime,
		ro.Channel, metadata(ro.Metadata), ro.CreatedBy,
	).Scan(&ro.Version, &ro.CreatedAt, &ro.UpdatedAt)
	switch {
	case err == nil:
		return nil
	case !errors.Is(err, pgx.ErrNoRows):
		return dbError(err, what)
	}
	// No row matched: either it is gone or its version moved on
	var current int
	if err := s.db.QueryRow(ctx,
		"SELECT version FROM openmos.running_orders WHERE ro_id = $1", ro.ID).Scan(&current); err != nil {
		return dbError(err, what)
	}
	return fmt.Errorf("%s at version %d, update based on %d: %w",
		what, current, ro.Version, repository.ErrVersionConflict)
}

func (s runningOrders) Delete(ctx context.Context, roID string) error {
	tag, err := s.db.Exec(ctx, "DELETE FROM openmos.running_orders WHERE ro_id = $1", roID)
	if err != nil {
		return dbError(err, "running order "+roID)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("running order %s: %w", roID, repository.ErrNotFound)
	}
	return nil
}

func (s runningOrders) List(ctx context.Context) ([]*model.RunningOrder, error) {
	rows, err := s.db.Query(ctx,
		"SELECT "+roColumns+" FROM openmos.running_orders ORDER BY created_at DESC, ro_id DESC")
	if err != nil {
		return nil, dbError(err, "running orders")
	}
	list, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (*model.RunningOrder, error) {
		return scanRunningOrder(row)
	})
	if err != nil {
		return nil, dbError(err, "running orders")
	}
	return list, nil
}
