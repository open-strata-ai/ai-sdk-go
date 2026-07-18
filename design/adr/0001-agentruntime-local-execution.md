# ADR-0001: AgentRuntime local execution

- **Status**: Accepted — see R-001 in `openstrata-meta/contracts/adr-resolutions.md`
- **Date**: 2026-07-17
- **Suggested by**: OpenStrata Architecture Group
- **Repository**: ai-sdk-go
- **Source**: `design/DESIGN.md` §11 Open Issue
- **Association**: (within this repository)

##Context

Does the SDK provide a "pure Go native map executor" as a lightweight implementation of the `AgentRuntime` port (without relying on Python LangGraph / Java Spring AI instances)? Or just do remote binding?

## Decision Options (Options Considered)

1. **Maintain status quo / conservative default**: Maintain current behavior, controlled by configuration switches or explicit parameters, and do not introduce destructive changes.
2. **Unified implementation after cross-repository alignment**: Agree on a clear contract with the relevant service (`corresponding governance service`) before implementation.
3. **Phased introduction**: Leave a placeholder/default switch in the current stage, and solidify it in subsequent stages after the dependent capabilities are ready (see Related Architecture §).

## Recommended decision (Decision)

This ADR solidifies "AgentRuntime local execution" into an architectural decision record and incorporates it into `design/adr/` for continuous tracking. This issue stems from the `design/DESIGN.md` §11 open issue and is still open.

**Conservative Default Principle**: Before the final decision is made, the "minimum available + explicit configuration switch" shall prevail, maintain the current behavior, and not destroy the existing contract and cross-repository SPI interface; this ADR status will be written back after review by the relevant team.



## To be aligned / Follow-ups (Follow-ups)

- **Resolution (R-001)**: Accepted — one `AgentRuntimePort` SPI, two bindings (local native executor + remote binding to `ai-gateway-core`), gateway-selected at runtime; default = remote, local executor opt-in for offline/air-gapped. See `openstrata-meta/contracts/adr-resolutions.md`.

## Traceback

- Upstream design: `design/DESIGN.md` §11 Open issue
- Relevance index: see `design/adr/README.md`
