# Developer Reference

This document provides a comprehensive guide to the OpenMOS codebase, including architecture, project structure, and feature roadmap.

## Overview

OpenMOS is an implementation of the Media Object Server (MOS) Protocol 4.0 using TCP socket communication. It is designed to manage running orders, stories, items, and media objects for broadcast and media environments, with the goal of achieving compliance with MOS Profile 7.

## Technology Stack

| Component | Technology | Version |
|-----------|------------|---------|
| Language | Go | 1.25+ |
| Database | PostgreSQL (in progress; in-memory until then) | 17 |
| Observability | Sentry | v0.50 |
| Communication | TCP Sockets | Port 10540 |
| Configuration | YAML | go.yaml.in/yaml/v3 |

## Architecture

OpenMOS follows a **Clean Layered Architecture** with clear separation of concerns:

```
┌─────────────────────────────────────────────────────────────┐
│                     TCP Clients (NCS)                       │
└─────────────────────────┬───────────────────────────────────┘
                          │
┌─────────────────────────▼───────────────────────────────────┐
│                    Server Layer                             │
│              (TCPServer, ClientConnection)                  │
└─────────────────────────┬───────────────────────────────────┘
                          │
┌─────────────────────────▼───────────────────────────────────┐
│                   Service Layer                             │
│                    (MOSService)                             │
└─────────────────────────┬───────────────────────────────────┘
                          │
┌─────────────────────────▼───────────────────────────────────┐
│                  Repository Layer                           │
│        (RunningOrder, Story, Item, Object Repos)            │
└─────────────────────────┬───────────────────────────────────┘
                          │
┌─────────────────────────▼───────────────────────────────────┐
│                   Database Layer                            │
│        (PostgreSQL; in-memory store until MOS-16)           │
└─────────────────────────────────────────────────────────────┘
```

### Layer Responsibilities

| Layer | Location | Responsibility |
|-------|----------|----------------|
| **Entry Point** | `main.go` | App initialization, config loading, graceful shutdown |
| **Server** | `internal/server/` | TCP connections, client management, heartbeat monitoring |
| **Service** | `internal/service/` | Business logic, MOS operations, event publishing |
| **Repository** | `internal/repository/` | Data access abstraction, CRUD operations |
| **Store** | `internal/store/postgres/` | PostgreSQL connection pool and schema migrations (repositories to follow) |
| **Model** | `internal/model/` | Domain entities and value objects |
| **XML** | `internal/xml/` | Message parsing and serialization |
| **Events** | `internal/events/` | Pub-sub event bus for real-time notifications |
| **Utilities** | `pkg/` | Logging, monitoring, shared helpers |

## Project Structure

```
OpenMOS/
├── src/                              # Source code root
│   ├── main.go                       # Application entry point
│   ├── go.mod                        # Go module definition
│   ├── go.sum                        # Dependency checksums
│   │
│   ├── internal/                     # Private application code
│   │   ├── config/
│   │   │   └── config.go             # YAML/env configuration loader
│   │   │
│   │   ├── events/
│   │   │   └── bus.go                # EventBus pub-sub implementation
│   │   │
│   │   ├── model/
│   │   │   ├── mos.go                # Core domain models
│   │   │   └── status.go             # Status type definitions
│   │   │
│   │   ├── repository/
│   │   │   ├── repository.go         # Interface definitions
│   │   │   ├── errors.go             # Shared errors (ErrNotFound, ...)
│   │   │   ├── memory/               # In-memory reference implementation
│   │   │   └── repotest/             # Contract suite for every implementation
│   │   │
│   │   ├── store/postgres/
│   │   │   ├── migrate.go            # Goose migrations (embedded SQL)
│   │   │   ├── pool.go               # pgx connection pool, TLS rules
│   │   │   └── migrations/           # Schema migrations (00001_init.sql, ...)
│   │   │
│   │   ├── server/
│   │   │   ├── server.go             # TCPServer main logic
│   │   │   ├── client.go             # ClientConnection management
│   │   │   └── client_story_handler.go # Story action handlers
│   │   │
│   │   ├── service/
│   │   │   ├── mos.go                # Main MOS service
│   │   │   └── story.go              # Story operations
│   │   │
│   │   └── xml/
│   │       ├── messages.go           # MOS message definitions
│   │       ├── story_messages.go     # Story-specific messages
│   │       ├── parser.go             # XML parser
│   │       ├── generator.go          # XML generator
│   │       └── heartbeat.go          # Heartbeat monitoring
│   │
│   └── pkg/                          # Shared utility packages
│       ├── logger/
│       │   ├── logger.go             # Base logger
│       │   └── sentry.go             # Sentry integration
│       ├── monitoring/               # Performance monitoring
│       └── utils/                    # Utility functions
│
├── doc/                              # Documentation
│   ├── readme.md                     # Documentation index
│   ├── developer-reference.md        # This file
│   ├── devtasks.md                   # Development milestones
│   ├── profiles.md                   # MOS Profile specifications
│   ├── mosfields.md                  # MOS field descriptions
│   └── releasenotes.md               # Release notes
│
├── res/                              # Resources
│   ├── mosCreate_example.xml         # Example MOS Create message
│   ├── mosModify_example.xml         # Example MOS Modify message
│   ├── mosReplace_example.xml        # Example MOS Replace message
│   ├── mosv4.dtd                     # MOS v4 DTD schema
│   ├── mosv4.xsd                     # MOS v4 XSD schema
│   └── mosproject-logo.jpg           # Project logo
│
├── README.md                         # Main project README
├── LICENSE                           # Project license
└── .gitignore                        # Git ignore rules
```

## Data Models

### Entity Relationships

```
RunningOrder (1) ──► (N) Story ──► (N) Item ──► MOSObject
     │                    │
     │                    └── PreviousStoryID / NextStoryID (linked list)
     │
     └── FirstStoryID / LastStoryID
```

### Core Entities

| Entity | Description | Key Fields |
|--------|-------------|------------|
| **RunningOrder** | Top-level container for broadcast content | ID, Slug, FirstStoryID, LastStoryID, Version |
| **Story** | Collection of items within a running order | ID, RunningOrderID, Slug, Order, PreviousStoryID, NextStoryID |
| **Item** | Individual element within a story | ID, StoryID, ObjectID, Slug, Order |
| **MOSObject** | Lowest-level media object | ID, Slug, Description, Status |

## Message Flow

1. **Client Connection**: TCP client connects to server on port 10540
2. **Heartbeat Monitoring**: Client heartbeat is tracked; timeout triggers disconnection
3. **Message Reception**: XML messages are parsed and validated
4. **Service Processing**: Business logic handles operations (create/update/replace)
5. **Storage**: Changes are stored through the repository interfaces (in memory until the PostgreSQL store lands)
6. **Event Publishing**: Service publishes events to the event bus
7. **Client Notification**: Connected clients receive real-time updates via subscriptions

## Implemented Features

### Core Infrastructure
- [x] TCP Socket Server with multi-client support
- [x] Repository pattern with a contract test suite and an in-memory store
- [ ] PostgreSQL persistence (M1, see database-migration.md)
- [x] YAML configuration with environment variable overrides
- [x] Multi-level logging with Sentry integration
- [x] Client heartbeat monitoring with timeout detection
- [x] Event bus for pub-sub real-time notifications
- [x] XML message parsing and generation
- [x] Graceful shutdown handling

### MOS Protocol
- [x] Running Order creation and basic management
- [x] Story create, update, and replace operations
- [x] Item creation and storage
- [x] MOS XML message processing

## Features To Be Implemented

### Profile Support

| Profile | Name | Status | Priority |
|---------|------|--------|----------|
| Profile 0 | Basic Communication | Pending | High |
| Profile 1 | Basic Object Based Workflow | Pending | High |
| Profile 2 | Basic Running Order / Content List Workflow | Pending | High |
| Profile 3 | Advanced Object Based Workflow | Pending | Medium |
| Profile 4 | Advanced RO/Content List Workflow | Pending | Medium |
| Profile 5 | Item Control | Pending | Medium |
| Profile 6 | MOS Redirection | Pending | Low |
| Profile 7 | MOS RO/Content List Modification | Pending | High |

### Profile 0 - Basic Communication
- [ ] `keepAlive` - Connection keep-alive message
- [ ] `heartBeat` - Heartbeat exchange between devices
- [ ] `reqMachInfo` - Request machine information
- [ ] `listMachInfo` - List machine information response

### Profile 1 - Basic Object Based Workflow
- [ ] `mosObj` - MOS object definition
- [ ] `mosReqObj` - Request MOS object
- [ ] `mosReqAll` - Request all MOS objects
- [ ] `mosAck` - MOS acknowledgment
- [ ] `mosListAll` - List all MOS objects response

### Profile 2 - Basic Running Order Workflow
- [ ] `roCreate` - Create running order
- [ ] `roReplace` - Replace running order
- [ ] `roDelete` - Delete running order
- [ ] `roReq` - Request running order
- [ ] `roList` - Running order list response
- [ ] `roMetadataReplace` - Replace running order metadata

### Profile 3 - Advanced Object Based Workflow
- [ ] `mosObjCreate` - Create MOS object
- [ ] `mosItemReplace` - Replace MOS item
- [ ] `mosReqSearchableSchema` - Request searchable schema
- [ ] `mosListSearchableSchema` - List searchable schema response

### Profile 4 - Advanced RO/Content List Workflow
- [ ] `roStoryAppend` - Append story to running order
- [ ] `roStoryInsert` - Insert story in running order
- [ ] `roStoryReplace` - Replace story in running order
- [ ] `roStoryMove` - Move story in running order
- [ ] `roStoryDelete` - Delete story from running order
- [ ] `roStorySwap` - Swap stories in running order
- [ ] `roItemInsert` - Insert item in story
- [ ] `roItemReplace` - Replace item in story
- [ ] `roItemMoveMultiple` - Move multiple items
- [ ] `roItemDelete` - Delete item from story
- [ ] `roReadyToAir` - Mark running order ready to air
- [ ] `roElementStat` - Element status update

### Profile 5 - Item Control
- [ ] `roCtrl` - Running order control command
- [ ] MQTT IoT device integration (experimental)
- [ ] Red light control use case

### Profile 6 - MOS Redirection
- [ ] `roReqStoryAction` - Request story action
- [ ] `roStoryAction` - Story action response

### Profile 7 - MOS RO/Content List Modification
- [ ] `roReqStoryMod` - Request story modification
- [ ] `roListStoryMod` - List story modification response

### Infrastructure Enhancements
- [ ] Docker image for deployment
- [ ] Database index optimization
- [ ] Connection pooling improvements
- [ ] Unit test coverage
- [ ] Integration test suite
- [ ] API documentation generation
- [ ] Metrics and monitoring dashboard

### Experimental Features
- [ ] MQTT protocol integration for IoT control
- [ ] Ember+ protocol support for machine control
- [ ] VDCP protocol support for video device control

## Configuration Reference

```yaml
app:
    name: OpenMOS              # Application name
    version: 1.0.0             # Version string
    environment: development   # Environment (development/staging/production)

server:
    host: 0.0.0.0              # Listen address
    port: 10540                # TCP port (MOS default)
    readtimeout: 5s            # Read timeout duration
    writetimeout: 5s           # Write timeout duration
    shutdowntimeout: 30s       # Graceful shutdown timeout

postgres:
    url: ""                    # Connection URL; a secret, so use DATABASE_URL instead
    maxconns: 10               # Pool size (env POSTGRES_MAX_CONNS)
    minconns: 1                # Connections kept open (env POSTGRES_MIN_CONNS)
    connecttimeout: 5s         # Connect and startup ping timeout (POSTGRES_CONNECT_TIMEOUT)
    healthcheckperiod: 30s     # Idle connection health checks (POSTGRES_HEALTH_CHECK_PERIOD)

mos:
    id: mos01.station.com      # MOS server identifier
    heartbeatinterval: 30s     # Heartbeat interval
    clienttimeout: 2m0s        # Client timeout before disconnect

logging:
    level: info                # Log level (debug/info/warning/error/fatal)

sentry:
    dsn: ""                    # Sentry DSN (leave empty to disable)
    environment: development   # Sentry environment tag
    debug: false               # Enable Sentry debug mode
    attachstacktrace: true     # Attach stack traces to events
    samplerate: 1              # Error sampling rate (0.0-1.0)
    tracessamplerate: 0.2      # Performance trace sampling rate
```

### Database connection

- `DATABASE_URL` (the backend01 `.env` name) overrides `postgres.url`. With
  neither set, the standard `PG*` variables (`PGHOST`, `PGUSER`, ...) are used;
  with nothing set, the server runs without a database.
- When a database is configured, the server connects and pings it at startup
  and exits if that fails. Logs and errors name host, database and user, never
  the password.
- Outside `app.environment: development`, connections must use
  `sslmode=verify-full` with `sslrootcert` pointing at the server CA (on the
  server, `/secrets/aiven-ca.pem`). Connect to Aiven directly, not through
  its PgBouncer pool.

## Development Guidelines

### Code Organization
- Keep business logic in the `service` layer
- Data access should go through `repository` interfaces
- Models should be pure data structures without behavior
- Use the event bus for cross-component communication

### Error Handling
- Wrap errors with context using `fmt.Errorf("context: %w", err)`
- Log errors at the appropriate level
- Use Sentry for production error tracking

### Testing
- Write unit tests for service and repository layers
- Every repository implementation must pass the contract suite:
  call `repotest.Run` from its tests (see `repository/memory`)
- Use interfaces to enable mocking
- Test XML parsing with example messages from `/res/`
- PostgreSQL tests (`internal/store/postgres`) need a throwaway database in
  `OPENMOS_TEST_DATABASE_URL`; without it they skip locally (CI provides one).
  They drop and recreate the `openmos` schema, so never point it at real data:

  ```bash
  docker compose up -d   # from the repository root; also creates openmos_test
  export OPENMOS_TEST_DATABASE_URL='postgres://openmos:openmos@localhost:5432/openmos_test?sslmode=disable'
  go test ./internal/store/...
  ```

### Schema migrations
- Migrations are SQL files in `internal/store/postgres/migrations`, embedded
  in the binary and applied with goose (`postgres.Migrate`). Tables and the
  goose version table live in the `openmos` schema.
- Never edit a migration that has been applied anywhere; add a new one.
- Migrations run as the schema owner (`openmos_migrate` on Aiven) over a
  direct connection, not a PgBouncer pool. Do not add GRANTs: the runtime
  role `openmos_app` gets privileges through `ALTER DEFAULT PRIVILEGES`.

### Contributing
1. Create a feature branch from main
2. Follow existing code style and patterns
3. Add tests for new functionality
4. Update documentation as needed
5. Submit a pull request with a clear description

## Related Documentation

- [Development Tasks](./devtasks.md) - Current milestones and known issues
- [MOS Profiles](./profiles.md) - Profile specifications
- [MOS Fields](./mosfields.md) - Field descriptions from MOS v4 spec
- [Release Notes](./releasenotes.md) - Version history
- [MOS Protocol Official Site](https://www.mosprotocol.com) - Protocol documentation

<hr />

[BACK](./readme.md)
