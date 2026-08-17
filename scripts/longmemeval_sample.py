#!/usr/bin/env python3
"""LongMemEval slice helpers. Prefix --limit is temporal-heavy; mixed is stratified.

Prefix-n is not overall V1. overlap ≠ gpt-4o ≠ V2 LAFS. Not Memory GA.
"""

from __future__ import annotations

from collections import Counter, defaultdict
from typing import Any, Dict, Iterable, List


def example_type(example: Dict[str, Any]) -> str:
    for key in ("question_type", "type", "category"):
        val = example.get(key)
        if val:
            return str(val)
    return "unknown"


def type_histogram(examples: Iterable[Dict[str, Any]]) -> Dict[str, int]:
    return dict(Counter(example_type(ex) for ex in examples))


def sample_mixed(examples: List[Dict[str, Any]], n: int) -> List[Dict[str, Any]]:
    """Round-robin by question_type so small n is not 100% temporal-reasoning."""
    if n <= 0 or n >= len(examples):
        return list(examples)
    buckets: Dict[str, List[Dict[str, Any]]] = defaultdict(list)
    for ex in examples:
        buckets[example_type(ex)].append(ex)
    keys = sorted(buckets)
    out: List[Dict[str, Any]] = []
    i = 0
    while len(out) < n:
        progressed = False
        for key in keys:
            bucket = buckets[key]
            if i < len(bucket):
                out.append(bucket[i])
                progressed = True
                if len(out) >= n:
                    break
        if not progressed:
            break
        i += 1
    return out


PREFIX_LIMIT_WARNING = (
    "warning: --limit N uses dataset prefix order; official V1 first rows are "
    "temporal-reasoning. Use --sample mixed for a mixed-type slice. "
    "Prefix-n is not overall V1. overlap ≠ gpt-4o ≠ V2 LAFS."
)


def apply_limit(
    examples: List[Dict[str, Any]],
    limit: int,
    sample: str,
    warn=None,
) -> List[Dict[str, Any]]:
    if limit <= 0 or limit >= len(examples):
        return list(examples)
    mode = (sample or "prefix").strip().lower()
    if mode == "mixed":
        return sample_mixed(examples, limit)
    if warn is not None:
        warn(PREFIX_LIMIT_WARNING)
    return examples[:limit]
