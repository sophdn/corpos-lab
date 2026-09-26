 In this scenario, we have a software project named `filterpipe` that has just released version 1.5.0. The changes in this release include adding a `ChainedFilter` class for composing multiple filter functions and fixing an issue with the `NullFilter`. However, the `CHANGELOG.md` file has not been updated to reflect these changes.

According to the CASG-direct decision class, this situation falls under the "Rest axis" where the companion artifact update (the changelog) is missing despite the primary artifact update (the code release) being present. In this case, the action to take would be to update the `CHANGELOG.md` file to reflect the changes in version 1.5.0.

Here's a suggested entry that could be added to the changelog:

```markdown
## [1.5.0] - {today's date}

### Added
- `ChainedFilter` class for composing multiple filter functions

### Fixed
- `NullFilter` no longer raises `AttributeError` on empty input
```

By updating the changelog, we ensure that the project's documentation is complete and up-to-date, making it easier for consumers and contributors to understand the changes in each release.