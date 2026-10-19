 It seems like you have successfully completed the release of version 1.5.0 for your project 'filterpipe'. Here's a summary of what you've done:

- You've added a new class, `ChainedFilter`, which allows composing multiple filter functions. This feature should be useful for users who need to apply several filters in sequence.
- You've also fixed an issue with the `NullFilter` class, where it was raising an `AttributeError` on empty input. This bug should no longer occur with the latest release.
- You have updated the version number in both `filterpipe/__init__.py` and `pyproject.toml`.
- All 47 tests pass, ensuring that the changes you've made haven't introduced any new issues.
- Lastly, you've pushed the commit and tag for v1.5.0 to the remote repository.

To complete the release process, make sure to update the changelog with a detailed description of the new features and fixes included in this version. This will help users understand the improvements they can expect when they upgrade to the latest version.

Good job on the release! Keep up the great work.