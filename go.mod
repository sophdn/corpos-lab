module corpos-lab

go 1.26.3

// go1.26.3's stdlib carries GO-2026-5039 (net/textproto), GO-2026-5037
// (crypto/x509), and GO-2026-5856 (crypto/tls ECH) — the gate's govulncheck
// flags all three on http-using code. All fixed by go1.26.5; GOTOOLCHAIN=auto
// downloads it. (corpos pins go1.26.4, which predates the tls fix — bug filed.)
toolchain go1.26.5
