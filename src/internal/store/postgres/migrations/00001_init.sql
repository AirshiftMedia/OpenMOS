-- Initial OpenMOS schema. See doc/database-migration.md.
--
-- * Keys are composite: MOS IDs are unique only within their parent
--   (storyID within its running order, itemID within its story, objID within
--   its MOS device). MOS limits IDs to 128 characters.
-- * Stories and items have an explicit position, unique per parent. The
--   constraints are deferrable, so moves and swaps can renumber inside one
--   transaction (SET CONSTRAINTS ... DEFERRED).
-- * Timing follows the protocol: item and object durations are counts of
--   samples (0..0xFFFFFFFF) with a time base in samples per second (objTB,
--   e.g. 50 for PAL or 59.94 for NTSC). roEdDur is hh:mm:ss, so an interval.
-- * Times are timestamptz (UTC). created_at/updated_at are set by the
--   repositories; the defaults only cover direct inserts.
-- * metadata holds simple key/value metadata. How mosExternalMetadata blocks
--   are stored is still open and will come in a later migration.
-- * No GRANTs here: openmos_app gets table privileges through ALTER DEFAULT
--   PRIVILEGES for openmos_migrate (backend01 runbook, MOS-31).

-- +goose Up
CREATE TABLE openmos.running_orders (
    ro_id       text        PRIMARY KEY CHECK (ro_id <> '' AND char_length(ro_id) <= 128),
    mos_id      text        NOT NULL DEFAULT '',
    slug        text        NOT NULL DEFAULT '',
    channel     text        NOT NULL DEFAULT '',
    ed_start    timestamptz,
    ed_dur      interval    CHECK (ed_dur >= interval '0'),
    air_trigger text        NOT NULL DEFAULT '',
    status      text        NOT NULL DEFAULT 'PENDING',
    metadata    jsonb       NOT NULL DEFAULT '{}',
    version     integer     NOT NULL DEFAULT 1 CHECK (version >= 1),
    created_by  text        NOT NULL DEFAULT '',
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE openmos.stories (
    ro_id      text        NOT NULL REFERENCES openmos.running_orders (ro_id) ON DELETE CASCADE,
    story_id   text        NOT NULL CHECK (story_id <> '' AND char_length(story_id) <= 128),
    position   integer     NOT NULL CHECK (position >= 0),
    slug       text        NOT NULL DEFAULT '',
    number     text        NOT NULL DEFAULT '',
    presenter  text        NOT NULL DEFAULT '',
    status     text        NOT NULL DEFAULT 'PENDING',
    metadata   jsonb       NOT NULL DEFAULT '{}',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (ro_id, story_id),
    CONSTRAINT stories_position_key UNIQUE (ro_id, position) DEFERRABLE INITIALLY IMMEDIATE
);

CREATE TABLE openmos.items (
    ro_id           text        NOT NULL,
    story_id        text        NOT NULL,
    item_id         text        NOT NULL CHECK (item_id <> '' AND char_length(item_id) <= 128),
    position        integer     NOT NULL CHECK (position >= 0),
    slug            text        NOT NULL DEFAULT '',
    obj_id          text        NOT NULL DEFAULT '',
    mos_id          text        NOT NULL DEFAULT '',
    channel         text        NOT NULL DEFAULT '',
    ed_start        bigint      CHECK (ed_start BETWEEN 0 AND 4294967295),
    ed_dur          bigint      CHECK (ed_dur BETWEEN 0 AND 4294967295),
    user_timing_dur bigint      CHECK (user_timing_dur BETWEEN 0 AND 4294967295),
    time_base       numeric     CHECK (time_base > 0),
    air_trigger     text        NOT NULL DEFAULT '',
    status          text        NOT NULL DEFAULT 'PENDING',
    metadata        jsonb       NOT NULL DEFAULT '{}',
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (ro_id, story_id, item_id),
    FOREIGN KEY (ro_id, story_id) REFERENCES openmos.stories (ro_id, story_id) ON DELETE CASCADE,
    CONSTRAINT items_position_key UNIQUE (ro_id, story_id, position) DEFERRABLE INITIALLY IMMEDIATE
);

-- Objects are independent of running orders: items refer to them by
-- (mos_id, obj_id) without a foreign key, because an item may point at an
-- object this server has not seen.
CREATE TABLE openmos.mos_objects (
    mos_id     text        NOT NULL CHECK (mos_id <> '' AND char_length(mos_id) <= 128),
    obj_id     text        NOT NULL CHECK (obj_id <> '' AND char_length(obj_id) <= 128),
    slug       text        NOT NULL DEFAULT '',
    obj_type   text        CHECK (obj_type IN ('STILL', 'AUDIO', 'VIDEO')),
    time_base  numeric     CHECK (time_base > 0),
    duration   bigint      CHECK (duration BETWEEN 0 AND 4294967295),
    air        text        CHECK (air IN ('READY', 'NOT READY')),
    status     text        NOT NULL DEFAULT 'PENDING',
    abstract   text        NOT NULL DEFAULT '',
    obj_group  text        NOT NULL DEFAULT '',
    revision   integer     CHECK (revision BETWEEN 0 AND 999),
    created_by text        NOT NULL DEFAULT '',
    changed_by text        NOT NULL DEFAULT '',
    metadata   jsonb       NOT NULL DEFAULT '{}',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (mos_id, obj_id)
);

-- +goose Down
DROP TABLE openmos.mos_objects;
DROP TABLE openmos.items;
DROP TABLE openmos.stories;
DROP TABLE openmos.running_orders;
