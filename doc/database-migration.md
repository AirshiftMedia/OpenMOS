# Database migration: MongoDB to PostgreSQL

OpenMOS is replacing MongoDB with PostgreSQL. MongoDB support will be removed,
not kept alongside, once the PostgreSQL store passes the same tests.

## Why

- The MOS data model is relational: a running order holds ordered stories, a
  story holds ordered items, and items reference media objects.
- MOS messages such as `roCreate`, `roReplace` and `roStoryMove` change many
  rows at once and must succeed or fail as a whole, which calls for
  transactions.
- Only the storage layer (about 480 lines) uses MongoDB today, so switching
  now is cheaper than after more profiles are built.

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

- `internal/db`: drop the MongoDB client; a new `internal/store/postgres`
  package provides the pool, transactions and the four repositories.
- `internal/repository`: interfaces stay, with composite keys, ordering
  operations and shared errors (`ErrNotFound`, `ErrVersionConflict`).
- Services run each MOS message in one transaction, and running order updates
  use version-checked writes (optimistic locking).
- Configuration: the `mongo` block becomes a `postgres` block, with
  `OPENMOS_DATABASE_URL` or the standard `PG*` environment variables.
- Startup can apply migrations automatically (`--migrate`) or only migrate and
  exit (`--migrate-only`) for deployment jobs.

## Steps

1. Fix the build, add CI and an implementation-independent repository test
   suite.
2. Add the schema migration and the PostgreSQL repositories, passing that
   suite.
3. Switch the server to PostgreSQL and add transactions.
4. Remove the MongoDB code and driver, and update the documentation.

<hr />

[BACK](./readme.md)
