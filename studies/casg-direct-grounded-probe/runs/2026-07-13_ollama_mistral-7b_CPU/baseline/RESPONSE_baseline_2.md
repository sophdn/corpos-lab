 To ensure the release is complete, follow these steps:

1. Update the version in the CHANGELOG.md file to reflect the new v1.5.0 release. The entry should include a description of the changes made in this release, such as adding the `ChainedFilter` class and fixing the `NullFilter` issue.

2. Add a new section for version 1.5.0 at the top of the CHANGELOG.md file. This section can be formatted as follows:

```markdown
## [1.5.0] - YYYY-MM-DD

### Changes

* Added ChainedFilter class for composing multiple filter functions ([#PR_NUMBER])
* Fixed NullFilter raising AttributeError on empty input ([#PR_NUMBER])
```

3. Replace `[#PR_NUMBER]` with the appropriate GitHub issue numbers related to the changes made in this release. If you haven't created issues for these changes yet, create them now and replace the placeholders accordingly.

4. Commit and push your changes to the CHANGELOG.md file:

```bash
git add CHANGELOG.md
git commit -m "Update CHANGELOG for 1.5.0 release"
git push origin main
```

5. Finally, create a new GitHub release on the `filterpipe` repository with the following steps:

a) Go to the [GitHub web interface](https://github.com/yourusername/filterpipe) for the `filterpipe` repository.
b) Click on the "Releases" tab and then click on the green "Draft a new release" button.
c) Fill out the form with the following information:
   - Title: `v1.5.0`
   - Tag version: `v1.5.0`
   - Release notes: Paste the description of changes you added to the CHANLOG.md file.
   - Publish the release draft.

Your v1.5.0 release is now complete, and the changelog has been updated accordingly.