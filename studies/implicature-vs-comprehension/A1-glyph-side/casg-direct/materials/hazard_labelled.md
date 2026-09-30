# Companion-artifact state

A release whose changelog is out of date is an incomplete, defective release. It misleads downstream users about what the release contains.

`CHANGELOG.md` records the changes shipped in each release of this project. It is maintained as part of cutting a release, by whoever cuts the release; no separate role owns it.

Its most recent entry is v1.4.0. The v1.5.0 changes — the `ChainedFilter` class and the `NullFilter` empty-input fix — are missing from it. The changelog currently misrepresents the release: it shows v1.4.0 while v1.5.0 has shipped.
