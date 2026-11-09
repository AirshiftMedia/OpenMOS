package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"airshift/openmos/internal/model"
	"airshift/openmos/internal/repository"
)

type objects struct{ db DBTX }

// objectColumns are read by scanObject. An empty type and a zero time base
// are stored as NULL, meaning unknown.
const objectColumns = `mos_id, obj_id, COALESCE(obj_type, ''), slug, COALESCE(duration, 0),
	COALESCE(time_base, 0)::float8, status, abstract, metadata, created_at, updated_at`

func scanObject(row pgx.Row) (*model.MOSObject, error) {
	var o model.MOSObject
	err := row.Scan(&o.MosID, &o.ID, &o.ObjectType, &o.Slug, &o.Duration, &o.TimeBase,
		&o.Status, &o.Abstract, &o.Metadata, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func objectName(mosID, objID string) string { return "object " + mosID + "/" + objID }

func (s objects) Create(ctx context.Context, obj *model.MOSObject) (*model.MOSObject, error) {
	if err := repository.CheckID("MOS", obj.MosID); err != nil {
		return nil, err
	}
	if err := repository.CheckID("object", obj.ID); err != nil {
		return nil, err
	}
	out, err := scanObject(s.db.QueryRow(ctx, `
		INSERT INTO openmos.mos_objects
			(mos_id, obj_id, obj_type, slug, duration, time_base, status, abstract, metadata)
		VALUES ($1, $2, NULLIF($3, ''), $4, $5, NULLIF($6::float8, 0)::numeric, $7, $8, $9)
		RETURNING `+objectColumns,
		obj.MosID, obj.ID, obj.ObjectType, obj.Slug, obj.Duration, obj.TimeBase,
		string(obj.Status), obj.Abstract, metadata(obj.Metadata)))
	if err != nil {
		return nil, dbError(err, objectName(obj.MosID, obj.ID))
	}
	obj.CreatedAt, obj.UpdatedAt = out.CreatedAt, out.UpdatedAt
	return out, nil
}

func (s objects) Get(ctx context.Context, mosID, objID string) (*model.MOSObject, error) {
	obj, err := scanObject(s.db.QueryRow(ctx,
		"SELECT "+objectColumns+" FROM openmos.mos_objects WHERE mos_id = $1 AND obj_id = $2", mosID, objID))
	if err != nil {
		return nil, dbError(err, objectName(mosID, objID))
	}
	return obj, nil
}

func (s objects) Update(ctx context.Context, obj *model.MOSObject) error {
	err := s.db.QueryRow(ctx, `
		UPDATE openmos.mos_objects SET
			obj_type = NULLIF($3, ''), slug = $4, duration = $5, time_base = NULLIF($6::float8, 0)::numeric,
			status = $7, abstract = $8, metadata = $9, updated_at = now()
		WHERE mos_id = $1 AND obj_id = $2
		RETURNING created_at, updated_at`,
		obj.MosID, obj.ID, obj.ObjectType, obj.Slug, obj.Duration, obj.TimeBase,
		string(obj.Status), obj.Abstract, metadata(obj.Metadata),
	).Scan(&obj.CreatedAt, &obj.UpdatedAt)
	if err != nil {
		return dbError(err, objectName(obj.MosID, obj.ID))
	}
	return nil
}

func (s objects) Delete(ctx context.Context, mosID, objID string) error {
	tag, err := s.db.Exec(ctx, "DELETE FROM openmos.mos_objects WHERE mos_id = $1 AND obj_id = $2", mosID, objID)
	if err != nil {
		return dbError(err, objectName(mosID, objID))
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("%s: %w", objectName(mosID, objID), repository.ErrNotFound)
	}
	return nil
}

func (s objects) List(ctx context.Context) ([]*model.MOSObject, error) {
	rows, err := s.db.Query(ctx,
		"SELECT "+objectColumns+" FROM openmos.mos_objects ORDER BY created_at DESC, mos_id DESC, obj_id DESC")
	if err != nil {
		return nil, dbError(err, "objects")
	}
	list, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (*model.MOSObject, error) {
		return scanObject(row)
	})
	if err != nil {
		return nil, dbError(err, "objects")
	}
	return list, nil
}
