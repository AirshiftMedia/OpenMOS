package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"airshift/openmos/internal/model"
	"airshift/openmos/internal/repository"
)

type stories struct{ db DBTX }

var storySiblings = siblings{
	table: "openmos.stories",
	idCol: "story_id",
	where: "ro_id = $1",
	lock:  "SELECT 1 FROM openmos.running_orders WHERE ro_id = $1 FOR NO KEY UPDATE",
}

const storyColumns = "ro_id, story_id, position, slug, number, presenter, status, metadata, created_at, updated_at"

func scanStory(row pgx.Row) (*model.Story, error) {
	var s model.Story
	err := row.Scan(&s.RunningOrderID, &s.ID, &s.Order, &s.Slug, &s.Number, &s.Presenter,
		&s.Status, &s.Metadata, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func storyName(roID, storyID string) string { return "story " + roID + "/" + storyID }

func checkStoryKey(roID, storyID string) error {
	if err := repository.CheckID("running order", roID); err != nil {
		return err
	}
	return repository.CheckID("story", storyID)
}

// insertStory adds story to its running order at position, given as SQL.
// The caller holds the parent lock.
func insertStory(ctx context.Context, tx pgx.Tx, story *model.Story, position string) (*model.Story, error) {
	out, err := scanStory(tx.QueryRow(ctx, `
		INSERT INTO openmos.stories (ro_id, story_id, position, slug, number, presenter, status, metadata)
		VALUES ($1, $2, `+position+`, $3, $4, $5, $6, $7)
		RETURNING `+storyColumns,
		story.RunningOrderID, story.ID, story.Slug, story.Number, story.Presenter,
		string(story.Status), metadata(story.Metadata)))
	if err != nil {
		return nil, dbError(err, storyName(story.RunningOrderID, story.ID))
	}
	story.Order, story.CreatedAt, story.UpdatedAt = out.Order, out.CreatedAt, out.UpdatedAt
	return out, nil
}

func (s stories) Create(ctx context.Context, story *model.Story) (*model.Story, error) {
	if err := checkStoryKey(story.RunningOrderID, story.ID); err != nil {
		return nil, err
	}
	var out *model.Story
	err := pgx.BeginFunc(ctx, s.db, func(tx pgx.Tx) error {
		roID := story.RunningOrderID
		if err := storySiblings.lockParent(ctx, tx, "running order "+roID, roID); err != nil {
			return err
		}
		var err error
		out, err = insertStory(ctx, tx, story, storySiblings.nextPosition())
		return err
	})
	return out, err
}

func (s stories) Get(ctx context.Context, roID, storyID string) (*model.Story, error) {
	story, err := scanStory(s.db.QueryRow(ctx,
		"SELECT "+storyColumns+" FROM openmos.stories WHERE ro_id = $1 AND story_id = $2", roID, storyID))
	if err != nil {
		return nil, dbError(err, storyName(roID, storyID))
	}
	return story, nil
}

func (s stories) Update(ctx context.Context, story *model.Story) error {
	err := s.db.QueryRow(ctx, `
		UPDATE openmos.stories SET
			slug = $3, number = $4, presenter = $5, status = $6, metadata = $7, updated_at = now()
		WHERE ro_id = $1 AND story_id = $2
		RETURNING position, created_at, updated_at`,
		story.RunningOrderID, story.ID, story.Slug, story.Number, story.Presenter,
		string(story.Status), metadata(story.Metadata),
	).Scan(&story.Order, &story.CreatedAt, &story.UpdatedAt)
	if err != nil {
		return dbError(err, storyName(story.RunningOrderID, story.ID))
	}
	return nil
}

func (s stories) Delete(ctx context.Context, roID, storyID string) error {
	tag, err := s.db.Exec(ctx, "DELETE FROM openmos.stories WHERE ro_id = $1 AND story_id = $2", roID, storyID)
	if err != nil {
		return dbError(err, storyName(roID, storyID))
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("%s: %w", storyName(roID, storyID), repository.ErrNotFound)
	}
	return nil
}

func (s stories) ListByRunningOrder(ctx context.Context, roID string) ([]*model.Story, error) {
	rows, err := s.db.Query(ctx,
		"SELECT "+storyColumns+" FROM openmos.stories WHERE ro_id = $1 ORDER BY position, story_id", roID)
	if err != nil {
		return nil, dbError(err, "stories of "+roID)
	}
	list, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (*model.Story, error) {
		return scanStory(row)
	})
	if err != nil {
		return nil, dbError(err, "stories of "+roID)
	}
	return list, nil
}

func (s stories) Move(ctx context.Context, roID string, storyIDs []string, beforeID string) error {
	return storySiblings.reorder(ctx, s.db, "move stories in "+roID, []any{roID},
		func(order []string) ([]string, error) { return repository.MoveIDs(order, storyIDs, beforeID) })
}

func (s stories) Swap(ctx context.Context, roID, storyID1, storyID2 string) error {
	return storySiblings.reorder(ctx, s.db, "swap stories in "+roID, []any{roID},
		func(order []string) ([]string, error) { return repository.SwapIDs(order, storyID1, storyID2) })
}

func (s stories) ReplaceAll(ctx context.Context, roID string, list []*model.Story) error {
	ids := make([]string, len(list))
	for i, story := range list {
		if err := checkStoryKey(roID, story.ID); err != nil {
			return err
		}
		ids[i] = story.ID
	}
	if err := repository.CheckDistinct(ids); err != nil {
		return fmt.Errorf("replace stories in %s: %w", roID, err)
	}
	return pgx.BeginFunc(ctx, s.db, func(tx pgx.Tx) error {
		if err := storySiblings.lockParent(ctx, tx, "running order "+roID, roID); err != nil {
			return err
		}
		// Deleting the stories deletes their items too
		if _, err := tx.Exec(ctx, "DELETE FROM openmos.stories WHERE ro_id = $1", roID); err != nil {
			return dbError(err, "replace stories in "+roID)
		}
		for i, story := range list {
			story.RunningOrderID = roID
			if _, err := insertStory(ctx, tx, story, fmt.Sprint(i)); err != nil {
				return err
			}
		}
		return nil
	})
}
