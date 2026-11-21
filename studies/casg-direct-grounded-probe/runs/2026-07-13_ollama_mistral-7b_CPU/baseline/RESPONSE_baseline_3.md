 To ensure that the release of `filterpipe v1.5.0` is complete, follow these steps:

1. Update the version in the `__version__` variable in the `filterpipe/__init__.py` file with the new version number (v1.5.0). The updated code should look something like this:

```python
__version__ = '1.5.0'
```

2. Update the project version in the `pyproject.toml` file as well:

```toml
[tool.poetry]
name = "filterpipe"
version = "1.5.0"
```

3. Add an entry for this release in the `CHANGELOG.md` file, following the existing format. The new entry should include a brief description of the changes and fixes made in the v1.5.0 release:

```markdown
## 1.5.0 - Release Date

### Added
- ChainedFilter class for composing multiple filter functions

### Fixed
- NullFilter raising AttributeError on empty input
```

4. Commit the changes to your local repository, ensuring that all files have been updated:

```bash
git add .
git commit -m "Bump version to 1.5.0 and update CHANGELOG"
```

5. Tag the commit as a release on GitHub:

```bash
git tag v1.5.0
git push origin v1.5.0
```

6. Finally, push the tag to the remote repository:

```bash
git push --tags
```

Now, the release of `filterpipe v1.5.0` is complete and can be found on the project's GitHub page.