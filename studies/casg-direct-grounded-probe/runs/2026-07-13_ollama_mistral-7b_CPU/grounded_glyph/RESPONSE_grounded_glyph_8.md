 In this scenario, the `filterpipe v1.5.0` release has been completed, but the corresponding entry for this release is missing in the `CHANGELOG.md` file. The action to be taken in this case is to update the `CHANGELOG.md` file to reflect the changes made in the `v1.5.0` release.

Here's an example of how the updated `CHANGELOG.md` entry should look:

```markdown
## [1.5.0] - {today's date}

### Added
- `ChainedFilter` class for composing multiple filter functions

### Fixed
- `NullFilter` no longer raises `AttributeError` on empty input
```

This will ensure that the `CHANGELOG.md` file is up to date and reflects the changes made in the `v1.5.0` release, making it complete and useful for consumers and contributors checking the changelog.