 The scenario describes a software release of a Python package called `filterpipe` version 1.5.0. The release involves updating the version numbers in `filterpipe/__init__.py` and `pyproject.toml`, running tests, committing and tagging the changes, and pushing them to a remote repository.

The new features added in this release are:
1. A `ChainedFilter` class for composing multiple filter functions
2. Fixing the `NullFilter` to no longer raise an `AttributeError` on empty input

However, the changelog at `CHANGELOG.md` has not been updated yet to reflect these changes. To ensure the release is complete as a communication artifact, it is recommended to prepend the following entry to `CHANGELOG.md`:

```markdown
## [1.5.0] - {today's date}

### Added
- `ChainedFilter` class for composing multiple filter functions

### Fixed
- `NullFilter` no longer raises `AttributeError` on empty input
```

This will bring the changelog and the actual release state into sync, providing a complete record of the changes in this version.