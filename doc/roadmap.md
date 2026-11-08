# Roadmap

OpenMOS is organised around the MOS 4.0 profiles. A profile is a numbered set of
messages that a MOS device can claim to support; the project goal is Profile 7.
This page shows how ready each profile is and which milestone delivers it.

Status as of 2026-11-08 (commit `3a73754`).

## Readiness by profile

| Profile | Scope | Status | Milestone |
| --- | --- | --- | --- |
| 0 | Basic communication: `heartbeat`, `reqMachInfo`, `listMachInfo` | Partial: `heartbeat` only, with non-standard attributes | M2 |
| 1 | Basic object workflow: `mosObj`, `mosReqObj`, `mosReqAll`, `mosListAll` | Not started (storage layer exists) | M3 |
| 2 | Basic running order workflow: `roCreate`, `roReplace`, `roDelete`, `roReq`, `roList`, `roAck`, `roMetadataReplace`, `roElementStat`, `roReadyToAir`, `roElementAction` | Partial: `roCreate` stores ROs and stories, items are dropped | M3 |
| 3 | Advanced object workflow: `mosObjCreate`, `mosItemReplace`, searchable schema, object lists | Not started | M4 |
| 4 | Advanced running order workflow: `roReqAll`, `roListAll`, `roStorySend` | Partial and incorrect (`roReq`/`roReqAll` swapped) | M4 |
| 5 | Item control: `roItemCue`, `roCtrl` | Not started | M5 |
| 6 | MOS redirection | Not started | M5 |
| 7 | RO modification requested by the MOS device: `roReqStoryAction` | Not started | M5 |

## Milestones

Milestones run in order; each one builds on the previous.

1. **M1 Foundations and PostgreSQL.** Build passes for the whole module, CI runs
   build, vet and tests, and storage moves from MongoDB to PostgreSQL.
   See [database-migration.md](./database-migration.md).
2. **M2 Wire protocol and Profile 0.** Messages use the `<mos>` envelope
   (`mosID`, `ncsID`, `messageID`), UTF-16BE encoding, robust stream framing,
   and the lower (10540) and upper (10541) ports. Messages are tested against
   [`res/mosv4.xsd`](../res/mosv4.xsd).
3. **M3 Profiles 1 and 2.** Object and running order basics, with `roAck`
   replies.
4. **M4 Profiles 3 and 4.** Advanced object and running order workflows.
5. **M5 Profiles 5, 6 and 7.** Item control, redirection and RO modification.
   The MQTT-based tally light idea from the README is a separate experiment
   within M5.

## Known issues

Tracked until the corresponding milestone fixes them:

- `go build ./...` fails in `pkg/monitoring` (uses a sentry-go function that
  no longer exists). The package is not used by the server. (M1)
- `roCreate` for a new running order with a numeric duration panics the
  connection handler. (M1)
- Running order updates are not protected against concurrent writes. (M1)
- `roReq` and `roReqAll` are handled the wrong way round. (M2)
- Durations are stored as whole seconds; MOS uses frames and a time base. (M1)

<hr />

[BACK](./readme.md)
