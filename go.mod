module corpos-lab

go 1.26.3

// go1.26.5's stdlib carries GO-2026-6218 (net/url), GO-2026-6090 (crypto/tls),
// GO-2026-5972 (encoding/asn1), and GO-2026-5026 (net/http idna) — the gate's
// govulncheck flags them on http-using code (internal/model, internal/substrate).
// All fixed by go1.26.6; GOTOOLCHAIN=auto downloads it when the local toolchain
// is older. Bump this directive as new stdlib advisories land.
toolchain go1.26.6

require github.com/BurntSushi/toml v1.6.0
