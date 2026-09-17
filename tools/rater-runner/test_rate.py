#!/usr/bin/env python3
"""Regression tests for rate.py.

The lab gate (scripts/gate.sh) is Go-only, so these do NOT run at commit time.
Run them by hand from this directory:

    python3 test_rate.py

The load-bearing case is test_relative_paths_complete (bug 1338): rate.py points
each rater at an isolated scratch cwd, so a relative --slices-dir/--out-dir or a
relative script inside --rater-cmd used to fail every slice. The fix resolves all
paths to absolute before the rater runs.
"""
import importlib.util
import json
import os
import shutil
import tempfile
import unittest
from pathlib import Path

HERE = Path(__file__).resolve().parent
_spec = importlib.util.spec_from_file_location("rate", HERE / "rate.py")
rate = importlib.util.module_from_spec(_spec)
_spec.loader.exec_module(rate)

# A tiny mechanical rater: read the slice JSONL, code every id "C", write the map.
# It is referenced by a RELATIVE path in --rater-cmd on purpose.
RATER_SCRIPT = '''import json, sys
a = sys.argv[1:]
inp = a[a.index("--in") + 1]
out = a[a.index("--out") + 1]
codes = {}
for line in open(inp):
    line = line.strip()
    if line:
        codes[json.loads(line)["id"]] = "C"
json.dump(codes, open(out, "w"))
'''


class RateTest(unittest.TestCase):
    def setUp(self):
        self._cwd = os.getcwd()
        self.tmp = tempfile.mkdtemp(prefix="rate-test-")
        os.chdir(self.tmp)

    def tearDown(self):
        os.chdir(self._cwd)
        shutil.rmtree(self.tmp, ignore_errors=True)

    def _write_slice(self, name, ids):
        d = Path("slices")
        d.mkdir(exist_ok=True)
        p = d / name
        p.write_text("".join(json.dumps({"id": i, "text": "x"}) + "\n" for i in ids))
        return p

    def test_relative_paths_complete(self):
        """Bug 1338: relative --slices-dir/--out-dir and a relative rater script."""
        self._write_slice("s0.jsonl", ["a1", "b2"])
        Path("score.py").write_text(RATER_SCRIPT)
        rc = rate.main([
            "--slices-dir", "slices",
            "--out-dir", "scores",
            "--rater-id", "mech",
            "--rater-cmd", "python3 score.py --in {slice} --out {out}",
        ])
        self.assertEqual(rc, 0)
        out = json.loads(Path("scores/mech/s0.json").read_text())
        self.assertEqual(set(out), {"a1", "b2"})

    def test_relative_single_slice(self):
        """The --slice single-file form also takes a relative path."""
        self._write_slice("s1.jsonl", ["z9"])
        Path("score.py").write_text(RATER_SCRIPT)
        rc = rate.main([
            "--slice", "slices/s1.jsonl",
            "--out-dir", "scores",
            "--rater-id", "mech",
            "--rater-cmd", "python3 score.py --in {slice} --out {out}",
        ])
        self.assertEqual(rc, 0)
        out = json.loads(Path("scores/mech/s1.json").read_text())
        self.assertEqual(set(out), {"z9"})

    def test_absolute_paths_still_work(self):
        """Absolute invocation is unchanged by the fix."""
        self._write_slice("s2.jsonl", ["m3", "m4"])
        Path("score.py").write_text(RATER_SCRIPT)
        rc = rate.main([
            "--slices-dir", os.path.abspath("slices"),
            "--out-dir", os.path.abspath("scores"),
            "--rater-id", "mech",
            "--rater-cmd", f"python3 {os.path.abspath('score.py')} --in {{slice}} --out {{out}}",
        ])
        self.assertEqual(rc, 0)
        out = json.loads(Path("scores/mech/s2.json").read_text())
        self.assertEqual(set(out), {"m3", "m4"})


if __name__ == "__main__":
    unittest.main()
