# ADR-0005: Streaming contract

- **Status**: Pending Alignment
- **Date**: 2026-07-17
- **Suggested by**: OpenStrata Architecture Group
- **Repository**: ai-sdk-go
- **Source**: `design/DESIGN.md` §11 Open Issue
- **Association**: (within this repository)

##Context

Three SDK's `LLMProvider.Stream` return type (Go channel / Java Flow / Python async iterator) how to ensure semantic equivalence requires the specs layer to supplement the cross-language streaming contract. --- ### Tail #### Open Issues See §11 (5 pending, converging with ADR). #### Change Record | Version | Date | Description | | v1.0.0-Draft | 2026-07-17 | First version of detailed design, covering the placeholder skeleton; 11 sections + meta information + traceability matrix | #### Traceability matrix (this document section ↔ Architectural Design Document §) | This document | Architectural Design Document § | Description | | §1 Positioning | §15.5.1 / §15.6.2 | SDK role, only relies on standard library | | §2 Core abstraction | §4.3.5 / §15.5.2 / §15.5.4 | AgentSpec / DDD four layers / Port=SPI | | §3 Installation initialization | §15.5.1 / §12 | Framework convergence, configuration driver | | §4 Usage example | §4.3.5 / §4.4.1 / §4.4.4 / §10.4 | AgentSpec / Gateway / LLMProvider / VectorStore | | §5 SPI extension point | §10.4 / §15.5.4 | Same name port, dependency inversion, ACL | | §6 Platform SPI mapping | §10.4 / §10.6 / §16 | 15 port name + interface_versions | §7 Configuration and Credentials | §12 / §4.4.6 | Manifest / Key Vault / Outbound Control | | §8 Version Compatibility | §16.1 / §16.2 | SemVer / interface_versions | | §9 Observability | §4.8 / §16 | OTel / Langfuse / Audit Baseline | | §10 Test Release | §15.6.4 / §16.4 | CI/BOM pin versions per bin | | §11 Open issues | §4.3.3 / §4.3.5 / §4.8 | Pending items |

## Decision Options (Options Considered)

1. **Maintain status quo / conservative default**: Maintain current behavior, controlled by configuration switches or explicit parameters, and do not introduce destructive changes.
2. **Unified implementation after cross-repository alignment**: Agree on a clear contract with the relevant service (`corresponding governance service`) before implementation.
3. **Phased introduction**: Leave a placeholder/default switch in the current stage, and solidify it in subsequent stages after the dependent capabilities are ready (see Related Architecture §).

## Recommended decision (Decision)

This ADR solidifies the "stream contract" into an architectural decision record and incorporates it into `design/adr/` for continuous tracking. This issue stems from the `design/DESIGN.md` §11 open issue and is still open.

**Conservative Default Principle**: Before the final decision is made, the "minimum available + explicit configuration switch" shall prevail, maintain the current behavior, and not destroy the existing contract and cross-repository SPI interface; this ADR status will be written back after review by the relevant team.



## To be aligned / Follow-ups (Follow-ups)

- Associated architecture documents §1 (as a basis for decision-making and a source of consistency verification).
- Associated architecture documents §10 (as a basis for decision-making and a source of consistency verification).
- Associated architecture documents §10.4 (as a basis for decision-making and a source of consistency verification).
- Associated architecture documents §10.6 (as a basis for decision-making and a source of consistency verification).
- Associated architecture documents §11 (as a basis for decision-making and a source of consistency verification).
- Associated architecture documents §12 (as a basis for decision-making and a source of consistency verification).
- Associated architecture documents §15.5.1 (as a basis for decision-making and a source of consistency verification).
- Associated architecture documents §15.5.2 (as a basis for decision-making and a source of consistency verification).
- Associated architecture documents §15.5.4 (as a basis for decision-making and a source of consistency verification).
- Associated architecture documents §15.6.2 (as a basis for decision-making and a source of consistency verification).
- Associated architecture documents §15.6.4 (as a basis for decision-making and a source of consistency verification).
- Associated architecture documents §16 (as a basis for decision-making and a source of consistency verification).
- Associated architecture documents §16.1 (as a basis for decision-making and a source of consistency verification).
- Associated architecture documents §16.2 (as a basis for decision-making and a source of consistency verification).
- Associated architecture documents §16.4 (as a basis for decision-making and a source of consistency verification).
- Associated architecture documents §2 (as a basis for decision-making and a source of consistency verification).
- Associated architecture documents §3 (as a basis for decision-making and a source of consistency verification).
- Associated architecture documents §4 (as a basis for decision-making and a source of consistency verification).
- Associated architecture documents §4.3.3 (as a basis for decision-making and a source of consistency verification).
- Associated architecture documents §4.3.5 (as a basis for decision-making and a source of consistency verification).
- Associated architecture documents §4.4.1 (as a basis for decision-making and a source of consistency verification).
- Associated architecture documents §4.4.4 (as a basis for decision-making and a source of consistency verification).
- Associated architecture documents §4.4.6 (as a basis for decision-making and a source of consistency verification).
- Associated architecture documents §4.8 (as a basis for decision-making and a source of consistency verification).
- Associated architecture documents §5 (as a basis for decision-making and a source of consistency verification).
- Associated architecture documents §6 (as a basis for decision-making and a source of consistency verification).
- Associated architecture documents §7 (as a basis for decision-making and a source of consistency verification).
- Associated architecture documents §8 (as a basis for decision-making and a source of consistency verification).
- Associated architecture documents §9 (as a basis for decision-making and a source of consistency verification).
- Solidify the decision before the review at the corresponding stage, and write the final conclusion back into this ADR (the status is changed from "Pending" to "Adopted").

## Traceback

- Upstream design: `design/DESIGN.md` §11 Open issue
- Relevance index: see `design/adr/README.md`
