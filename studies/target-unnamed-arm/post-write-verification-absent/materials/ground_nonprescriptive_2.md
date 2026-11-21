Each `UPDATE — statement OK` confirms the statement was issued and accepted. It does not confirm how many rows it changed, or that the customer records now hold the region set to 'EU' for the DE, FR, and NL customers: a statement that matched zero rows — a country stored in a different case or form — returns the same OK.

The authoritative source for the resulting rows is a read of the customer records, not the acknowledgments. The acknowledgments are identical whether rows changed or not, so on their own they leave the resulting rows unverified.
