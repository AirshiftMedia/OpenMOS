package postgres

import (
	"slices"
	"testing"
)

// tables are the OpenMOS tables, sorted, without goose's version table.
var tables = []string{"items", "mos_objects", "running_orders", "stories"}

func TestMigrateUpDownUp(t *testing.T) {
	db := testDB(t)
	ctx := t.Context()

	listTables := func() []string {
		t.Helper()
		rows, err := db.QueryContext(ctx, `SELECT tablename FROM pg_tables
			WHERE schemaname = $1 AND tablename <> 'goose_db_version' ORDER BY tablename`, Schema)
		if err != nil {
			t.Fatal(err)
		}
		defer rows.Close()
		var names []string
		for rows.Next() {
			var n string
			if err := rows.Scan(&n); err != nil {
				t.Fatal(err)
			}
			names = append(names, n)
		}
		if err := rows.Err(); err != nil {
			t.Fatal(err)
		}
		return names
	}

	// Up on an empty database creates the schema and every table
	applied, err := Migrate(ctx, db)
	if err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	if !slices.Equal(applied, []int64{1}) {
		t.Errorf("applied = %v, want [1]", applied)
	}
	if got := listTables(); !slices.Equal(got, tables) {
		t.Errorf("tables after up = %v, want %v", got, tables)
	}

	// Up again is a no-op
	applied, err = Migrate(ctx, db)
	if err != nil {
		t.Fatalf("second Migrate: %v", err)
	}
	if len(applied) != 0 {
		t.Errorf("second Migrate applied %v, want nothing", applied)
	}

	// Down removes every table
	if err := MigrateDownTo(ctx, db, 0); err != nil {
		t.Fatalf("MigrateDownTo(0): %v", err)
	}
	if got := listTables(); len(got) != 0 {
		t.Errorf("tables after down = %v, want none", got)
	}
	if v, err := Version(ctx, db); err != nil || v != 0 {
		t.Errorf("Version after down = %d, %v; want 0", v, err)
	}

	// And up works again after a full rollback
	if _, err := Migrate(ctx, db); err != nil {
		t.Fatalf("Migrate after down: %v", err)
	}
	if got := listTables(); !slices.Equal(got, tables) {
		t.Errorf("tables after re-up = %v, want %v", got, tables)
	}
	if v, err := Version(ctx, db); err != nil || v != 1 {
		t.Errorf("Version = %d, %v; want 1", v, err)
	}
}
