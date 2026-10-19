# Scenario — config-default fix

A colleague reports a bug in the `settings` module: `load_settings()` raises `KeyError: 'timeout'` when the config file omits the `timeout` key. The agreed fix is to fall back to a default of 30 seconds when `timeout` is absent.

Here is the current function:

```python
def load_settings(config: dict) -> Settings:
    return Settings(
        host=config["host"],
        port=config["port"],
        timeout=config["timeout"],
    )
```

Return the corrected `load_settings` so it uses a default `timeout` of 30 when the key is absent.
