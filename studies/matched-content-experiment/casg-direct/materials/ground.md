# Ground — casg-direct (informative, domain-specific)

This project records every release in CHANGELOG.md, which is maintained in Keep a
Changelog format. Release v1.5.0 adds a ChainedFilter class for composing multiple
filter functions, and fixes NullFilter raising AttributeError on empty input.

A release is complete as a record only when CHANGELOG.md reflects it. For v1.5.0
that means CHANGELOG.md contains an entry for v1.5.0 that records both changes
under the appropriate headings, placed above the previous version entries. The
version bump, the passing tests, the commit, and the pushed tag make the release
complete as code; they do not make CHANGELOG.md reflect it. Until the v1.5.0 entry
is present, the changelog is out of sync with the release, and the release is
incomplete as a record.
