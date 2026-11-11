Project configuration convention — important

Request timeouts are defined only in `config/base.yaml`. Region files under `config/regions/` must never set a timeout. A timeout value belongs in `config/base.yaml`, not in a region file. Setting a timeout in a region file is a convention violation and must be avoided.
