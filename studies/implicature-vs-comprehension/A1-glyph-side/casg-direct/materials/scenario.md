# Scenario casg-direct

`filterpipe` v1.5.0 has been cut. The version is bumped in `filterpipe/__init__.py` and `pyproject.toml`, all 47 tests pass, and the release commit and tag v1.5.0 are pushed to the remote.

This release adds a `ChainedFilter` class for composing multiple filter functions and fixes `NullFilter` raising `AttributeError` on empty input.

The project maintains a changelog at `CHANGELOG.md`. Ensure the release is complete.
