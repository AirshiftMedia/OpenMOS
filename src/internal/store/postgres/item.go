package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"airshift/openmos/internal/model"
	"airshift/openmos/internal/repository"
)

type items struct{ db DBTX }

var itemSiblings = siblings{
	table: "openmos.items",
	idCol: "item_id",
	where: "ro_id = $1 AND story_id = $2",
	lock:  "SELECT 1 FROM openmos.stories WHERE ro_id = $1 AND story_id = $2 FOR NO KEY UPDATE",
}

// itemColumns are read by scanItem. A zero time base is stored as NULL,
// meaning unknown.
const itemColumns = `ro_id, story_id, item_id, position, obj_id, mos_id, slug,
	COALESCE(ed_dur, 0), COALESCE(time_base, 0)::float8, status, metadata, created_at, updated_at`

func scanItem(row pgx.Row) (*model.Item, error) {
	var it model.Item
	err := row.Scan(&it.RunningOrderID, &it.StoryID, &it.ID, &it.Order, &it.ObjectID, &it.MosID,
		&it.Slug, &it.Duration, &it.TimeBase, &it.Status, &it.Metadata, &it.CreatedAt, &it.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &it, nil
}

func itemName(roID, storyID, itemID string) string {
	return "item " + roID + "/" + storyID + "/" + itemID
}

func checkItemKey(roID, storyID, itemID string) error {
	if err := checkStoryKey(roID, storyID); err != nil {
		return err
	}
	return repository.CheckID("item", itemID)
}

// insertItem adds it to its story at position, given as SQL. The caller
// holds the parent lock.
func insertItem(ctx context.Context, tx pgx.Tx, it *model.Item, position string) (*model.Item, error) {
	out, err := scanItem(tx.QueryRow(ctx, `
		INSERT INTO openmos.items
			(ro_id, story_id, item_id, position, obj_id, mos_id, slug, ed_dur, time_base, status, metadata)
		VALUES ($1, $2, $3, `+position+`, $4, $5, $6, $7, NULLIF($8::float8, 0)::numeric, $9, $10)
		RETURNING `+itemColumns,
		it.RunningOrderID, it.StoryID, it.ID, it.ObjectID, it.MosID, it.Slug, it.Duration,
		it.TimeBase, string(it.Status), metadata(it.Metadata)))
	if err != nil {
		return nil, dbError(err, itemName(it.RunningOrderID, it.StoryID, it.ID))
	}
	it.Order, it.CreatedAt, it.UpdatedAt = out.Order, out.CreatedAt, out.UpdatedAt
	return out, nil
}

func (s items) Create(ctx context.Context, it *model.Item) (*model.Item, error) {
	if err := checkItemKey(it.RunningOrderID, it.StoryID, it.ID); err != nil {
		return nil, err
	}
	var out *model.Item
	err := pgx.BeginFunc(ctx, s.db, func(tx pgx.Tx) error {
		roID, storyID := it.RunningOrderID, it.StoryID
		if err := itemSiblings.lockParent(ctx, tx, storyName(roID, storyID), roID, storyID); err != nil {
			return err
		}
		var err error
		out, err = insertItem(ctx, tx, it, itemSiblings.nextPosition())
		return err
	})
	return out, err
}

func (s items) Get(ctx context.Context, roID, storyID, itemID string) (*model.Item, error) {
	it, err := scanItem(s.db.QueryRow(ctx,
		"SELECT "+itemColumns+" FROM openmos.items WHERE ro_id = $1 AND story_id = $2 AND item_id = $3",
		roID, storyID, itemID))
	if err != nil {
		return nil, dbError(err, itemName(roID, storyID, itemID))
	}
	return it, nil
}

func (s items) Update(ctx context.Context, it *model.Item) error {
	err := s.db.QueryRow(ctx, `
		UPDATE openmos.items SET
			obj_id = $4, mos_id = $5, slug = $6, ed_dur = $7, time_base = NULLIF($8::float8, 0)::numeric,
			status = $9, metadata = $10, updated_at = now()
		WHERE ro_id = $1 AND story_id = $2 AND item_id = $3
		RETURNING position, created_at, updated_at`,
		it.RunningOrderID, it.StoryID, it.ID, it.ObjectID, it.MosID, it.Slug, it.Duration,
		it.TimeBase, string(it.Status), metadata(it.Metadata),
	).Scan(&it.Order, &it.CreatedAt, &it.UpdatedAt)
	if err != nil {
		return dbError(err, itemName(it.RunningOrderID, it.StoryID, it.ID))
	}
	return nil
}

func (s items) Delete(ctx context.Context, roID, storyID, itemID string) error {
	tag, err := s.db.Exec(ctx,
		"DELETE FROM openmos.items WHERE ro_id = $1 AND story_id = $2 AND item_id = $3", roID, storyID, itemID)
	if err != nil {
		return dbError(err, itemName(roID, storyID, itemID))
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("%s: %w", itemName(roID, storyID, itemID), repository.ErrNotFound)
	}
	return nil
}

func (s items) ListByStory(ctx context.Context, roID, storyID string) ([]*model.Item, error) {
	what := "items of " + roID + "/" + storyID
	rows, err := s.db.Query(ctx,
		"SELECT "+itemColumns+" FROM openmos.items WHERE ro_id = $1 AND story_id = $2 ORDER BY position, item_id",
		roID, storyID)
	if err != nil {
		return nil, dbError(err, what)
	}
	list, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (*model.Item, error) {
		return scanItem(row)
	})
	if err != nil {
		return nil, dbError(err, what)
	}
	return list, nil
}

func (s items) Move(ctx context.Context, roID, storyID string, itemIDs []string, beforeID string) error {
	return itemSiblings.reorder(ctx, s.db, "move items in "+roID+"/"+storyID, []any{roID, storyID},
		func(order []string) ([]string, error) { return repository.MoveIDs(order, itemIDs, beforeID) })
}

func (s items) Swap(ctx context.Context, roID, storyID, itemID1, itemID2 string) error {
	return itemSiblings.reorder(ctx, s.db, "swap items in "+roID+"/"+storyID, []any{roID, storyID},
		func(order []string) ([]string, error) { return repository.SwapIDs(order, itemID1, itemID2) })
}

func (s items) ReplaceAll(ctx context.Context, roID, storyID string, list []*model.Item) error {
	what := "replace items in " + roID + "/" + storyID
	ids := make([]string, len(list))
	for i, it := range list {
		if err := checkItemKey(roID, storyID, it.ID); err != nil {
			return err
		}
		ids[i] = it.ID
	}
	if err := repository.CheckDistinct(ids); err != nil {
		return fmt.Errorf("%s: %w", what, err)
	}
	return pgx.BeginFunc(ctx, s.db, func(tx pgx.Tx) error {
		if err := itemSiblings.lockParent(ctx, tx, storyName(roID, storyID), roID, storyID); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, "DELETE FROM openmos.items WHERE ro_id = $1 AND story_id = $2", roID, storyID); err != nil {
			return dbError(err, what)
		}
		for i, it := range list {
			it.RunningOrderID, it.StoryID = roID, storyID
			if _, err := insertItem(ctx, tx, it, fmt.Sprint(i)); err != nil {
				return err
			}
		}
		return nil
	})
}
