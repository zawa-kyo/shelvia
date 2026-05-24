# AGENTS.md

## Repository Map

- User-facing project explanation lives in `README.md` and `README-ja.md`.
- Implementation-facing architecture lives in `docs/architecture.md`.
- Keep `AGENTS.md` short. Treat it as a map to durable repository knowledge, not as a full manual.

## Documentation

- Keep `README.md` and `README-ja.md` synchronized. When one file changes, update the other in the same change unless the edit is intentionally language-specific.
- `README-ja.md` should be written first when the intended wording is being designed in Japanese. Then sync the same user-facing behavior, commands, data format, and project status to `README.md` in natural English.
- Keep README files focused on user-facing explanation: motivation, installation, usage, data format, and examples. Put implementation plans, migration notes, and architecture details in separate design documents.

## Architecture

- Follow the architecture and invariants in `docs/architecture.md` when adding implementation.
- Use Clean Architecture as the default dependency rule. Domain code must not depend on CLI, filesystem, SQLite, or presentation details.
- Update `docs/architecture.md` in the same change when architectural decisions, command behavior, data flow, or validation boundaries change.
- Shelvia is a CLI tool, so design documentation should focus on architecture, command behavior, data format, validation, diagnostics, and test strategy rather than visual design.
