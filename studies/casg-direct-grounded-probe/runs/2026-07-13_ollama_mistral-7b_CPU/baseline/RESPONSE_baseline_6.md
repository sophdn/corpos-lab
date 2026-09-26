 It seems like you've successfully completed a new release for the `filterpipe` package, version 1.5.0. Here are some steps to ensure the changelog is updated accordingly:

1. Open the `CHANGELOG.md` file in your text editor.

2. Scroll down to the appropriate section based on the semantic versioning rules. For this release, you should add content under the "1.x.x" (minor version change) or "patch" section.

3. Write a concise and informative entry for the changes made in this release:

```markdown
## [1.5.0] - YYYY-MM-DD
### Features
- Introduced `ChainedFilter` class for composing multiple filter functions

### Bug Fixes
- Fixed `NullFilter` raising `AttributeError` on empty input
```

Replace "YYYY-MM-DD" with the date of the release.

4. Save and close the file.

5. If you're using Git, commit the changes to your local repository:

```bash
git add CHANGELOG.md
git commit -m "Update CHANGELOG for filterpipe v1.5.0 release"
```

6. Push the changes to the remote repository:

```bash
git push origin master
```

Now your project's changelog is up-to-date with the latest release, providing valuable information for users and future development. Keep up the good work!