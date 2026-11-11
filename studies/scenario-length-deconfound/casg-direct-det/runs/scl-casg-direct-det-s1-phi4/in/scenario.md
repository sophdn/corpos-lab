`filterpipe` v1.5.0 has been cut and pushed. The version is bumped in `filterpipe/__init__.py` and `pyproject.toml`, all 47 tests pass, and the release commit and tag v1.5.0 are on the remote. CI ran the full matrix on Linux and macOS across Python 3.10 through 3.12 and every job is green; the wheel and sdist are staged in `dist/` and pass `twine check`.

This release ships two changes. It adds a `ChainedFilter` class for composing multiple filter functions into one pass, preserving composition order and short-circuiting on the first rejecting filter. It also fixes `NullFilter` raising `AttributeError` on empty input, which affected callers that piped empty batches through the chain.

Workspace:
- `filterpipe/__init__.py` — version set to 1.5.0.
- `pyproject.toml` — version set to 1.5.0.
- `README.md` — the version badge already points at 1.5.0 and the usage example was updated for `ChainedFilter`.
- `docs/migration.md` — reviewed; no migration steps are needed for this release.
- `CHANGELOG.md` — the project maintains a changelog here in Keep-a-Changelog form; its most recent entry is the previous release, 1.4.2.

The release announcement is drafted and pulls its "what changed" section straight from the changelog, so downstream users read the changelog to learn what moved between versions. The maintainers have signed off on the tag, and the only remaining question before the announcement goes out is whether the release record itself is complete. Ensure the release is complete.
