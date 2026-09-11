# Score Grid — assay-blank-claude-definition-work-v10

*Auto-populated: Verdict (C/I/N), Field Source (FS). Fill in Observable (O) and Evidence (E) quality (C/P/I) after reading response files.*

| Scenario | GT | R1-V | R1-FS | R2-V | R2-FS | R3-V | R3-FS | R4-V | R4-FS | Verdict | O | E | Scope-cited |
|----------|----| ------ | --- | ------ | --- | ------ | --- | ------ | --- |---------|---|---|-------------|
| scb-a | yes | C | Scope - operative when | C | Scope — operative when | C | Scope | C | Scope | 4/4 | · | · | N/A |
| scb-b | no | C | Scope — not operative when | C | Scope — not operative when | C | Scope - not operative when | C | Scope - Not Operative When | 4/4 | · | · | 2/4 |
| cgu-a | yes | C | Scope — operative when | C | Scope — operative when | C | Scope - operative when | C | Scope — operative when | 4/4 | · | · | N/A |
| cgu-b | no | C | Scope — operative when | C | Scope — operative when | I | Scope - not operative | C | Scope — operative when | 3/4 | · | · | 3/4 |
| cas-a | yes | C | Marker axis | C | Marker axis | I | Aim axis - Invariant | C | Marker axis | 3/4 | · | · | N/A |
| cas-b | no | C | Pull character | C | Pull character | C | Marker axis | C | Scope — operative when | 4/4 | · | · | 1/4 |
| fsb-a | yes | C | Scope — operative when | C | Scope — operative when | C | Scope — operative when | I | Scope — not operative when | 3/4 | · | · | N/A |
| fsb-b | no | I | Pull character | C | Scope — operative when | I | Pull character | C | Scope — operative when | 2/4 | · | · | 2/4 |
| gop-a | yes | C | Scope - operative when | C | Scope | C | Scope — operative when | C | Scope — operative when | 4/4 | · | · | N/A |
| gop-b | no | I | Scope — operative when | I | Scope — operative when | I | Scope - operative when | C | Scope — not operative when | 1/4 | · | · | 3/4 |
| psc-a | yes | I | Scope — not operative when | C | Scope — operative when | I | Scope — operative when | C | Scope — operative when | 2/4 | · | · | N/A |
| psc-b | no | C | Scope — operative when | C | Scope - operative when | C | Scope — operative when | C | Scope - operative when | 4/4 | · | · | 2/4 |

## Field source key

FS values come directly from the model's FIELD SOURCE response. Scope-cited = yes when FS references Scope-operative or Scope-not-operative. For b-scenarios, this column tracks carve-out field engagement.
