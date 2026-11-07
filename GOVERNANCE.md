# Governance

## Principles

- Open and transparent decision-making
- Meritocratic: contributions over tenure
- Consensus-seeking, with structured escalation
- Vendor-neutral: no single company controls the roadmap

## Roles

### Maintainers

- At least **3 maintainers from 2+ different organizations** at all times.
- Can merge PRs, cut releases, manage issues.
- Listed in `MAINTAINERS.md`.

### Contributors

- Anyone who has submitted a merged PR.
- Can open issues, request reviews, propose features.

### Steering Committee (future)

- Formed when the project reaches 10+ active contributors.
- 5 members, elected by maintainers + contributors.
- Meets monthly, sets roadmap, resolves disputes.

## Adding a Maintainer

1. Contributor must have 5+ merged PRs over 3+ months.
2. 2 existing maintainers nominate.
3. 3/4 maintainer approval (no single-org majority).
4. Update `MAINTAINERS.md` via PR.

## Removing a Maintainer

1. Self-resignation: update `MAINTAINERS.md`.
2. Inactivity > 6 months: 2/3 maintainer vote.
3. Misconduct: 3/4 vote + CNCF Code of Conduct review.

## Decision Process

| Decision Type | Method |
|--------------|--------|
| API / CRD changes | RFC in `docs/rfcs/`, 2-week comment, maintainer consensus |
| Roadmap | Quarterly planning, open to community input |
| Releases | Semantic versioning, 2-maintainer sign-off |
| Breaking changes | 2-week deprecation window, major version bump |
| Disputes | Steering Committee (or TOC if pre-incubation) |

## CNCF Alignment

This project follows the CNCF Technical Oversight Committee (TOC) guidelines:
- [CNCF IP Policy](https://github.com/cncf/foundation/blob/main/policies-guidance/cncf-ip-policy.md)
- [Technical Leadership Principles](https://github.com/cncf/toc/blob/main/toc_subprojects/project-reviews-subproject/technical-leadership-principles.md)
- [Code of Conduct](https://github.com/cncf/foundation/blob/main/policies-guidance/cncf-code-of-conduct.md)
