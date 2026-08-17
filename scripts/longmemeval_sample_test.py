#!/usr/bin/env python3
"""stdlib unittest for mixed-type LongMemEval slice (#58)."""

from __future__ import annotations

import io
import unittest

from longmemeval_sample import PREFIX_LIMIT_WARNING, apply_limit, sample_mixed, type_histogram


def _ex(qid: str, typ: str) -> dict:
    return {"question_id": qid, "question_type": typ}


class SampleTests(unittest.TestCase):
    def test_mixed_is_not_prefix_temporal_only(self) -> None:
        rows = (
            [_ex(f"t{i}", "temporal-reasoning") for i in range(60)]
            + [_ex(f"m{i}", "multi-session-synthesis") for i in range(10)]
            + [_ex(f"k{i}", "knowledge-update") for i in range(10)]
        )
        mixed = sample_mixed(rows, 12)
        hist = type_histogram(mixed)
        self.assertGreater(len(hist), 1, hist)
        self.assertIn("temporal-reasoning", hist)
        self.assertLess(hist["temporal-reasoning"], 12)

    def test_prefix_warns(self) -> None:
        buf = io.StringIO()
        rows = [_ex(f"t{i}", "temporal-reasoning") for i in range(20)]
        out = apply_limit(rows, 5, "prefix", warn=buf.write)
        self.assertEqual(len(out), 5)
        self.assertIn("prefix", buf.getvalue().lower())
        self.assertIn("mixed", PREFIX_LIMIT_WARNING.lower())


if __name__ == "__main__":
    unittest.main()
