# AGENTS.md

## Repository Map

- User-facing project explanation lives in `README.md` and `README-ja.md`.
- Implementation-facing design lives in `docs/`.
- Start from `docs/architecture.md`, then follow the linked documents for shelf behavior, implementation structure, and testing strategy.
- Keep `AGENTS.md` short. Treat it as a map to durable repository knowledge, not as a full manual.

## Documentation

- Keep `README.md` and `README-ja.md` synchronized. When one file changes, update the other in the same change unless the edit is intentionally language-specific.
- `README-ja.md` should be written first when the intended wording is being designed in Japanese. Then sync the same user-facing behavior, commands, data format, and project status to `README.md` in natural English.
- Keep README files focused on user-facing explanation: motivation, installation, usage, data format, and examples. Put implementation plans, migration notes, and architecture details in separate design documents.
- Do not stage changes unless the user explicitly asks for staging. Editing files is allowed, but leave staging decisions to the user.

## Architecture

- Follow the architecture and invariants in `docs/architecture.md` when adding implementation.
- Follow `docs/shelf.md` for `shelf root`, `config.toml`, file discovery, and command behavior.
- Follow `docs/implementation.md` for package layout, layer responsibilities, and dependency direction.
- Follow `docs/testing.md` for the test strategy.
- Use Clean Architecture as the default dependency rule. Domain code must not depend on CLI, filesystem, SQLite, or presentation details.
- Update the relevant document in `docs/` in the same change when architectural decisions, command behavior, data flow, validation boundaries, or test strategy change.
- Shelvia is a CLI tool, so design documentation should focus on architecture, command behavior, data format, validation, diagnostics, and test strategy rather than visual design.
