 It seems like you've successfully completed a new release of your project, `filterpipe`, version 1.5.0! Here are some steps to ensure that the release is complete:

1. Update the `__version__` variable in `filterpipe/__init__.py`. The value should be '1.5.0'.

```python
__version__ = "1.5.0"
```

2. Check if the version number is correctly reflected in your `pyproject.toml` file:

```toml
[tool.poetry.version]
"1.5.0"
```

3. Make sure that all 47 tests pass. You can run the tests using your preferred method (e.g., `pytest`, `unittest`, etc.).

4. Create a new section in the changelog (`CHANGELOG.md`) for this release, detailing the new features and fixes:

```markdown
## 1.5.0 - Release Date
### Features Added
- ChainedFilter class for composing multiple filter functions

### Bugs Fixed
- NullFilter now correctly handles empty input and no longer raises AttributeError
```

5. Commit the changes to your local repository:

```bash
git add .
git commit -m "Release v1.5.0: Add ChainedFilter, fix NullFilter"
```

6. Tag the commit as the release version:

```bash
git tag v1.5.0
```

7. Push both the commit and the tag to your remote repository:

```bash
git push origin master
git push --tags
```

8. Publish the new version on PyPI (if you use poetry):

```bash
poetry publish
```

That's it! You have now completed a successful release of `filterpipe` v1.5.0. Keep up the good work! 🚀