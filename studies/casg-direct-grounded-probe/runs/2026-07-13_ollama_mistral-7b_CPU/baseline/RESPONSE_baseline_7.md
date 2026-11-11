 To ensure that the release is complete, follow these steps:

1. Update the CHANGELOG.md file with the details of the new release (version 1.5.0). Include a description of the changes, such as the addition of `ChainedFilter` class for composing multiple filter functions and the fix for `NullFilter` raising `AttributeError` on empty input.

Here's an example entry for your change:
```markdown
## 1.5.0 (YYYY-MM-DD)

### Added
- ChainedFilter class for composing multiple filter functions

### Fixed
- NullFilter no longer raises AttributeError on empty input
```

Replace `(YYYY-MM-DD)` with the date of the release.

2. Push the updated CHANGELOG.md to your remote repository.

3. Verify that the tag v1.5.0 exists in your repository's tags. If it doesn't, create and push the tag using:
```bash
git tag v1.5.0
git push origin v1.5.0
```

4. Lastly, update any documentation or readme files that may require information about the new release.

Now, your release should be complete and ready to share with users!