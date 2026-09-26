The edit tool's acknowledgment confirms the edit was issued and accepted. It does not confirm that the configuration now holds a default timeout of 60: a successful edit and a silently-failed one return the same acknowledgment.

The authoritative source for the resulting content is a read of the edited configuration itself, not the acknowledgment. The acknowledgment is identical whether the edit landed or not, so on its own it leaves the resulting content unverified.
