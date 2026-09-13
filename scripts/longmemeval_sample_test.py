#!/usr/bin/env python3
"""stdlib unittest for mixed-type LongMemEval slice (#58)."""

from __future__ import annotations

import io
import os
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

    def test_limit_zero_is_full_file(self) -> None:
        rows = [_ex(f"t{i}", "temporal-reasoning") for i in range(8)] + [
            _ex(f"m{i}", "multi-session") for i in range(8)
        ]
        self.assertEqual(len(apply_limit(rows, 0, "mixed")), len(rows))
        self.assertEqual(len(apply_limit(rows, 12, "mixed")), 12)


class V1CardHarnessTests(unittest.TestCase):
    """Static checks: scored-run LIMIT unset/0 must not coerce to n=12."""

    def test_v1_card_does_not_coerce_unset_limit_to_12(self) -> None:
        root = os.path.dirname(os.path.abspath(__file__))
        path = os.path.join(root, "longmemeval_v1_card.sh")
        with open(path, encoding="utf-8") as f:
            src = f.read()
        self.assertNotIn("RUN_LIMIT=12", src)
        self.assertNotIn("using mixed n=12 for scored sample", src)
        self.assertIn(
            "official V1 scored run: mixed n=500 (full oracle); n=12 is not V1",
            src,
        )
        self.assertIn("LONGMEMEVAL_QA_LIMIT=12 is a mixed sample, not V1", src)
        # --limit only on the positive-LIMIT branch, not the unset/0 branch.
        unset_branch = src.split('log "official V1 scored run: mixed n=500', 1)[1]
        unset_branch = unset_branch.split("bash scripts/longmemeval_judge.sh", 1)[0]
        self.assertNotIn("--limit", unset_branch)


if __name__ == "__main__":
    unittest.main()
