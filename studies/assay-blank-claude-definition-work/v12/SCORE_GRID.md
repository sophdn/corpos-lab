# Score Grid — assay-blank-claude-definition-work-v12

*Auto-populated: Verdict (C/I/N), Field Source (FS). Fill in Observable (O) and Evidence (E) quality (C/P/I) after reading response files.*

| Scenario | GT | R1-V | R1-FS | R2-V | R2-FS | R3-V | R3-FS | R4-V | R4-FS | Verdict | O | E | Scope-cited |
|----------|----| ------ | --- | ------ | --- | ------ | --- | ------ | --- |---------|---|---|-------------|
| gop-a | yes | C | Scope - Independently initiated | C | Scope - Independently initiated | C | Scope - Independently initiated | C | Scope — operative when | 4/4 | C | C | N/A |
| gop-b | no | C | Artifact coupling | C | Artifact coupling | C | Scope - Downstream-of-consultation | C | Scope — not operative when | 4/4 | C | P | 1/4 |
| cas-a | yes | C | Scope | C | Artifact coupling | C | Artifact coupling | I | Scope — not operative when | 3/4 | P | P | N/A |
| cas-b | no | I | Scope - operative when | C | Artifact coupling | C | Artifact coupling | C | Scope — not operative when | 3/4 | P | P | 2/4 |

## Field source key

FS values come directly from the model's FIELD SOURCE response. Scope-cited = yes when FS references Scope-operative or Scope-not-operative. For b-scenarios, this column tracks carve-out field engagement.
