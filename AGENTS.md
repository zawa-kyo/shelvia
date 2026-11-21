# AGENTS.md

## Repository Map

- **User documentation**: Maintained in `README.md` and `README-ja.md`.
- **Architecture and design**: Maintained in `docs/`.
  - Start with `docs/architecture.md`, then refer to the linked documents for CLI behavior, implementation details, and testing strategy.
- Keep `AGENTS.md` brief—treat it as a navigation index to durable repository knowledge, not an exhaustive manual.
- Tracked README and `docs/` files are the canonical specification. Disregard untracked `.idea/` content or legacy booklog materials.

## Documentation Guidelines

- Keep `README.md` and `README-ja.md` synchronized. When updating one, update the other in the same change unless the edit is intentionally language-specific.
- Draft changes in `README-ja.md` first when designing wording in Japanese, then reflect the same user-facing behavior, commands, data format, and project status in `README.md` in natural English.
- README files must focus strictly on user-facing concerns: motivation, installation, usage, data format, and examples. Place implementation plans, migration notes, and architectural decisions in dedicated documents under `docs/`.
- Do not stage git changes unless the user explicitly requests it. Editing files is permitted, but leave staging and committing to the user.

## Architecture Guidelines

- Follow the Clean Architecture principles and invariants in `docs/architecture.md`:
  - `docs/cli-behavior.md`: `shelf root` resolution, `config.toml`, file discovery, and command behavior.
  - `docs/implementation.md`: Package layout, layer responsibilities, and dependency direction.
  - `docs/testing.md`: Unit, integration, and E2E testing strategies.
- Maintain inward dependency direction: domain code must never depend on the CLI, local filesystem, query engine, or presentation layers.
- Update the relevant document in `docs/` whenever architectural decisions, command behavior, data flow, validation rules, or test strategies change.
- Shelvia is a CLI tool; design documentation should prioritize architecture, command behavior, data formats, validation rules, diagnostics, and testing rather than visual design.
