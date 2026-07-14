 Based on the provided information, it seems that the `filterpipe v1.5.0` release has been successfully completed and pushed to the remote repository. However, there is no update in the `CHANGELOG.md` file to reflect this new version.

According to the Ground register for the casg-direct decision class, the action to take in this scenario is to prepend a new entry to the `CHANGELOG.md` file immediately below the `# Changelog` header (or at the top of the file if no `[Unreleased]` section is present).

Here's an example of what that updated `CHANGELOG.md` might look like:

```markdown
## [1.5.0] - {today's date}

### Added
- `ChainedFilter` class for composing multiple filter functions

### Fixed
- `NullFilter` no longer raises `AttributeError` on empty input
```

This update ensures that the changelog is consistent with the latest release and provides clear documentation for contributors and consumers of the project.