# Database migration: MongoDB to PostgreSQL

OpenMOS stores its data in PostgreSQL only. The original MongoDB store was
removed outright (MOS-34) rather than kept alongside or behind a switch.
Until the PostgreSQL store lands, the server runs on the in-memory reference
store (`internal/repository/memory`), so data is not persisted.

## Why

- The MOS data model is relational: a running order holds ordered stories, a
  story holds ordered items, and items reference media objects.
- MOS messages such as `roCreate`, `roReplace` and `roStoryMove` change many
  rows at once and must succeed or fail as a whole, which calls for
  transactions.
- Only the storage layer (about 480 lines) used MongoDB, so switching early
  was cheaper than after more profiles were built.

## Tooling

| Need | Choice | Notes |
| --- | --- | --- |
| Driver and pool | [pgx v5](https://github.com/jackc/pgx) (`pgxpool`) | Actively maintained, native PostgreSQL types |
| Migrations | [goose v3](https://github.com/pressly/goose) | Plain SQL files embedded in the binary |
| Integration tests | [testcontainers-go](https://github.com/testcontainers/testcontainers-go) | Runs a real PostgreSQL in tests and CI |
| Database | PostgreSQL 17 | Local development via Docker Compose |

## Schema principles

- **Composite keys.** MOS IDs are unique only within their parent: a `storyID`
  within its running order, an `itemID` within its story, an `objID` within
  its MOS device.
- **Explicit order.** Stories and items carry a `position` column, with
  deferrable unique constraints so moves and swaps renumber safely inside one
  transaction.
- **MOS timing.** Durations are stored as frames plus time base (`objTB`),
  as the protocol defines them.
- **UTC timestamps.** All times are `timestamptz`, exchanged as RFC 3339.
- **Metadata.** `mosExternalMetadata` blocks are kept per row (format,
  `jsonb` or verbatim XML, still to be decided).

Tables: `running_orders`, `stories`, `items`, `mos_objects`.

## Code changes

- `internal/db` and the MongoDB repositories are gone; a new `internal/store/postgres`
  package provides the pool, transactions and the four repositories.
- `internal/repository`: interfaces stay, gaining composite keys and ordering
  operations. Shared errors live in `errors.go` and grow with each change.
- Services run each MOS message in one transaction, and running order updates
  use version-checked writes (optimistic locking).
- Configuration: the old `mongo` block is gone; a new `postgres` block, with
  `DATABASE_URL` or the standard `PG*` environment variables.
- Startup can apply migrations automatically (`--migrate`) or only migrate and
  exit (`--migrate-only`) for deployment jobs.

## Steps

1. Fix the build, add CI and an implementation-independent repository test
   suite (MOS-11), with an in-memory reference store.
2. Remove the MongoDB code, driver and configuration (MOS-34).
3. Add the schema migration and the PostgreSQL repositories, passing that
   suite (MOS-12 to MOS-14).
4. Add transactions and switch the server to PostgreSQL (MOS-15, MOS-16).

<hr />

[BACK](./readme.md)
