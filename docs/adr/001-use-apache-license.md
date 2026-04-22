# ADR-001: Use Apache License 2.0

- **Status**: Accepted
- **Date**: 2026-04-22
- **Deciders**: Kane

## Context

CloudSentinel needs a license that:
1. Allows broad adoption in commercial and open-source settings
2. Aligns with the dominant license used by Canonical and the broader cloud-native ecosystem
3. Provides patent protection for contributors and users

## Decision

Use Apache License 2.0 for all source code and documentation.

## Consequences

### Positive
- Compatible with most permissive and copyleft projects used as dependencies
- Explicit patent grant protects users from patent trolls
- Aligns with Kubernetes, Terraform, Docker, and most Canonical projects
- Recognized and trusted by enterprise legal teams

### Negative
- Less viral than GPL — downstream projects can keep modifications private

### Neutral
- Requires license headers in source files (can be automated)

## Alternatives Considered

- **MIT**: simpler but no patent protection — risky for a project touching kernel networking
- **GPL v3**: too viral for a testing library intended for broad use
- **BSD-3**: similar to MIT, same patent concerns

## References

- [Apache License 2.0 full text](https://www.apache.org/licenses/LICENSE-2.0)
- [Kubernetes LICENSE](https://github.com/kubernetes/kubernetes/blob/master/LICENSE)
