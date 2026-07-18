# ADR-0003: LowCode canvas export

- **Status**: Pending (Open)
- **Date**: 2026-07-17
- **Suggested by**: OpenStrata Architecture Group
- **Repository**: ai-sdk-go
- **Source**: `docs/DESIGN.md` §11 Open Issue
- **Association**: (within this repository)

##Context

Is the `react-flow → AgentSpec` exporter of the `LowCode` port down to the SDK, or is it only provided on the platform side? Affects cross-language consistency.

## Decision Options (Options Considered)

1. **Maintain status quo / conservative default**: Maintain current behavior, controlled by configuration switches or explicit parameters, and do not introduce destructive changes.
2. **Unified implementation after cross-repository alignment**: Agree on a clear contract with the relevant service (`corresponding governance service`) before implementation.
3. **Phased introduction**: Leave a placeholder/default switch in the current stage, and solidify it in subsequent stages after the dependent capabilities are ready (see Related Architecture §).

## Recommended decision (Decision)

This ADR solidifies "LowCode Canvas Export" as an architectural decision record and incorporates it into `docs/adr/` for continuous tracking. This issue stems from the `docs/DESIGN.md` §11 open issue and is still open.

**Conservative Default Principle**: Before the final decision is made, the "minimum available + explicit configuration switch" shall prevail, maintain the current behavior, and not destroy the existing contract and cross-repository SPI interface; this ADR status will be written back after review by the relevant team.



## To be aligned / Follow-ups (Follow-ups)

- Solidify the decision before the review at the corresponding stage, and write the final conclusion back into this ADR (the status is changed from "Pending" to "Adopted").

## Traceback

- Upstream design: `docs/DESIGN.md` §11 Open issue
- Relevance index: see `docs/adr/README.md`
