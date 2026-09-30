# Score Grid — assay-blank-claude-definition-work-v11b

*Auto-populated: Verdict (C/I/N), Field Source (FS). Fill in Observable (O) and Evidence (E) quality (C/P/I) after reading response files.*

| Scenario | GT | R1-V | R1-FS | R2-V | R2-FS | R3-V | R3-FS | R4-V | R4-FS | Verdict | O | E | Scope-cited |
|----------|----| ------ | --- | ------ | --- | ------ | --- | ------ | --- |---------|---|---|-------------|
| cas-a | yes | I | Scope — operative when | I | Scope — not operative when | I | Scope — not operative when | I | Artifact coupling | 0/4 | · | · | N/A |
| cas-b | no | C | Scope — operative when | C | Scope — operative when | C | Scope — not operative when | C | Scope — not operative when | 4/4 | · | · | 4/4 |
| gop-a | yes | C | Scope | C | Scope | C | Scope - operative when | C | Scope | 4/4 | · | · | N/A |
| gop-b | no | C | Scope — operative when | I | Artifact coupling > Scope d… | I | Scope discriminator | I | Scope - Operative when | 1/4 | · | · | 1/4 |
| fsb-a | yes | C | Scope — not operative when | C | Scope — operative when | I | Scope — operative when | C | Scope - operative when | 3/4 | · | · | N/A |
| fsb-b | no | I | Artifact coupling | C | Marker axis, Invariant | C | Scope — operative when | C | Scope — operative when | 3/4 | · | · | 2/4 |

## Field source key

FS values come directly from the model's FIELD SOURCE response. Scope-cited = yes when FS references Scope-operative or Scope-not-operative. For b-scenarios, this column tracks carve-out field engagement.
