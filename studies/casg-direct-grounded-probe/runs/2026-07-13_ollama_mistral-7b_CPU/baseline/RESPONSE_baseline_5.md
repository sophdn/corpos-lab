 To ensure that the release of version 1.5.0 of `filterpipe` is complete, follow these steps:

1. Update the version in the CHANGELOG.md file:
   - Find the most recent section (probably a newline followed by '## v1.5.0')
   - Add a brief description of the changes made in this release under that section:
     ```
     ## v1.5.0 (2023-04-18)

     - Add `ChainedFilter` class for composing multiple filter functions
     - Fix `NullFilter` raising `AttributeError` on empty input
     ```

2. Commit the changes to CHANGELOG.md:
   ```
   git add CHANGELOG.md
   git commit -m "Update changelog for release v1.5.0"
   ```

3. Push the commit to the remote repository:
   ```
   git push origin main
   ```

4. If necessary, create a new pull request or merge the current branch into the main branch (depending on your workflow).

Now, the release of version 1.5.0 for `filterpipe` is complete and properly documented in the CHANGELOG.md file.