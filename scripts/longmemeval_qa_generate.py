#!/usr/bin/env python3
"""
Efficient LongMemEval QA hypothesis generator for github.com/iome-sh/memory.

Ingests haystack_sessions via the Go HTTP server and generates one OpenAI answer
per question (single combined prompt — no separate summarize step).

Prerequisites:
  1. Start the benchmark server:
       export MEMORY_ONNX_MODEL_PATH=testdata/models/KnightsAnalytics_bge-small-en-v1.5
       go run cmd/longmemeval-server/main.go
  2. pip install -r requirements-bench.txt
  3. export OPENAI_API_KEY=sk-...

Usage:
  python scripts/longmemeval_qa_generate.py \\
    --dataset data/longmemeval_oracle.json \\
    --output hypotheses.jsonl \\
    --limit 500 --workers 4
"""

from __future__ import annotations

import argparse
import json
import os
import sys
from concurrent.futures import ThreadPoolExecutor, as_completed
from pathlib import Path
from typing import Any, Dict, List, Optional

import requests
from tqdm import tqdm

_SCRIPTS = Path(__file__).resolve().parent
if str(_SCRIPTS) not in sys.path:
    sys.path.insert(0, str(_SCRIPTS))
from longmemeval_sample import apply_limit, type_histogram  # noqa: E402

try:
    import openai
except ImportError:
    print("pip install -r requirements-bench.txt", file=sys.stderr)
    sys.exit(1)

SERVER_URL = os.environ.get("LONGMEMEVAL_SERVER", "http://localhost:8765")
OPENAI_MODEL = os.environ.get("OPENAI_MODEL", "gpt-4o-mini")
RETRIEVE_K = int(os.environ.get("LONGMEMEVAL_RETRIEVE_K", "40"))


def check_server(session: requests.Session) -> bool:
    try:
        r = session.get(f"{SERVER_URL}/health", timeout=5)
        return r.status_code == 200
    except Exception as e:
        print(f"Server not reachable at {SERVER_URL}: {e}", file=sys.stderr)
        return False


def load_dataset(path: str) -> List[Dict[str, Any]]:
    with open(path, "r", encoding="utf-8") as f:
        data = json.load(f)
    if isinstance(data, list):
        return data
    if isinstance(data, dict):
        for key in ("examples", "data", "items", "questions"):
            if key in data:
                return data[key]
    return data


def haystack_history(example: Dict[str, Any]) -> List[Dict[str, Any]]:
    sessions = example.get("haystack_sessions")
    if not sessions:
        return (
            example.get("history")
            or example.get("conversation")
            or example.get("turns", [])
        )

    dates = example.get("haystack_dates") or []
    session_ids = example.get("haystack_session_ids") or []
    turns: List[Dict[str, Any]] = []
    for sess_idx, session in enumerate(sessions):
        ts = dates[sess_idx] if sess_idx < len(dates) else ""
        sid = session_ids[sess_idx] if sess_idx < len(session_ids) else f"sess-{sess_idx}"
        for turn_idx, turn in enumerate(session):
            turns.append({
                "role": turn.get("role", "user"),
                "content": turn.get("content", ""),
                "timestamp": ts,
                "cycle": turn_idx + 1,
                "session_id": sid,
            })
    return turns


def ingest_history(session: requests.Session, conv_id: str, history: List[Dict[str, Any]]) -> None:
    turns = [
        {
            "role": turn.get("role", "user"),
            "content": turn.get("content", ""),
            "timestamp": turn.get("timestamp", ""),
            "cycle": turn.get("cycle", 0),
        }
        for turn in history
    ]
    payload = {"conv_id": conv_id, "turns": turns}
    r = session.post(f"{SERVER_URL}/ingest", json=payload, timeout=120)
    r.raise_for_status()


def retrieve_memories(
    session: requests.Session,
    query: str,
    k: int,
    session_id: str = "",
) -> List[Dict[str, Any]]:
    payload: Dict[str, Any] = {"query": query, "limit": k}
    if session_id:
        payload["session_id"] = session_id
    r = session.post(f"{SERVER_URL}/retrieve", json=payload, timeout=60)
    r.raise_for_status()
    return r.json().get("memories", [])


def server_embed_mode(session: requests.Session) -> str:
    try:
        r = session.get(f"{SERVER_URL}/health", timeout=5)
        if r.status_code == 200:
            return str(r.json().get("embed_mode") or "unknown")
    except Exception:
        pass
    return "unknown"


def generate_answer(
    question: str,
    memories: List[Dict[str, Any]],
    question_date: str = "",
) -> str:
    """Single OpenAI call: extract key facts from memories and answer the question."""
    if not memories:
        context = "No relevant memories retrieved."
    else:
        lines = []
        for m in memories[:15]:
            full = m.get("full") or m.get("summary") or ""
            score = m.get("score", 0.0)
            ts = m.get("timestamp") or ""
            sid = m.get("session_id") or ""
            prefix = f"[score={score:.2f}"
            if ts:
                prefix += f" t={ts}"
            if sid:
                prefix += f" session={sid}"
            lines.append(f"{prefix}] {full}")
        context = "\n".join(lines)

    date_line = ""
    if question_date:
        date_line = f"Question date (use this as 'now' for temporal items): {question_date}\n\n"

    prompt = (
        "You are a helpful assistant with access to long-term personal memories.\n"
        "Read the retrieved memory snippets below. Extract relevant facts and answer "
        "the question as accurately as possible. If memories contain partial information, "
        "synthesize the best answer; only abstain if the information is truly absent.\n\n"
        f"{date_line}"
        f"Memories:\n{context}\n\n"
        f"Question: {question}\n\n"
        "Answer:"
    )
    client = openai.OpenAI()
    resp = client.chat.completions.create(
        model=OPENAI_MODEL,
        messages=[{"role": "user", "content": prompt}],
        temperature=0.0,
        max_tokens=400,
    )
    return (resp.choices[0].message.content or "").strip()


def snippet_from_memory(m: Dict[str, Any], max_len: int = 400) -> Dict[str, Any]:
    text = (m.get("full") or m.get("summary") or "")[:max_len]
    return {
        "id": m.get("id") or "",
        "session_id": m.get("session_id") or "",
        "timestamp": m.get("timestamp") or "",
        "text": text,
    }


def process_example(example: Dict[str, Any], retrieve_k: int) -> Optional[Dict[str, Any]]:
    qid = example.get("question_id") or example.get("id") or example.get("qid")
    question = example.get("question") or example.get("query")
    if not qid or not question:
        return None

    conv_id = str(example.get("conv_id") or example.get("conversation_id") or qid)
    question_date = str(example.get("question_date") or "")
    history = haystack_history(example)

    session = requests.Session()
    ingest_history(session, conv_id, history)
    memories = retrieve_memories(session, question, retrieve_k, session_id=conv_id)
    answer = generate_answer(question, memories, question_date=question_date)
    return {
        "question_id": str(qid),
        "hypothesis": answer,
        "question_date": question_date,
        "question_type": str(example.get("question_type") or ""),
        "retrieve": [snippet_from_memory(m) for m in memories[:15]],
    }


def main() -> None:
    parser = argparse.ArgumentParser(description="LongMemEval QA hypothesis generator")
    parser.add_argument("--dataset", required=True, help="Path to longmemeval_oracle.json or similar")
    parser.add_argument("--output", default="hypotheses.jsonl", help="Output hypotheses JSONL")
    parser.add_argument("--limit", type=int, default=0, help="Max examples (0 = all)")
    parser.add_argument(
        "--sample",
        choices=("prefix", "mixed"),
        default="prefix",
        help="prefix = first-N (official V1 start is temporal-only); mixed = stratified by question_type",
    )
    parser.add_argument("--workers", type=int, default=4, help="Parallel OpenAI+HTTP workers")
    parser.add_argument(
        "--server",
        default=os.environ.get("LONGMEMEVAL_SERVER", "http://localhost:8765"),
        help="LongMemEval server base URL",
    )
    args = parser.parse_args()

    global SERVER_URL
    SERVER_URL = args.server.rstrip("/")

    if not os.environ.get("OPENAI_API_KEY"):
        print("error: OPENAI_API_KEY is required for QA generation", file=sys.stderr)
        sys.exit(1)

    probe = requests.Session()
    if not check_server(probe):
        print(
            "Start the Go server first:\n"
            "  export MEMORY_ONNX_MODEL_PATH=testdata/models/KnightsAnalytics_bge-small-en-v1.5\n"
            "  go run cmd/longmemeval-server/main.go",
            file=sys.stderr,
        )
        sys.exit(1)

    examples = load_dataset(args.dataset)
    examples = apply_limit(examples, args.limit, args.sample, warn=lambda m: print(m, file=sys.stderr))
    hist = type_histogram(examples)
    print(f"slice n={len(examples)} sample={args.sample} types={hist}", file=sys.stderr)

    embed_mode = server_embed_mode(probe)
    print(f"embed_mode={embed_mode}", file=sys.stderr)

    hypotheses: List[Dict[str, Any]] = []
    errors = 0

    with ThreadPoolExecutor(max_workers=max(1, args.workers)) as pool:
        futures = {
            pool.submit(process_example, ex, RETRIEVE_K): ex
            for ex in examples
        }
        for fut in tqdm(as_completed(futures), total=len(futures), desc="QA generate"):
            ex = futures[fut]
            qid = ex.get("question_id") or ex.get("id") or "unknown"
            try:
                result = fut.result()
                if result is None:
                    errors += 1
                    continue
                result["embed_mode"] = embed_mode
                hypotheses.append(result)
            except Exception as e:
                errors += 1
                print(f"error for {qid}: {e}", file=sys.stderr)

    hypotheses.sort(key=lambda h: h["question_id"])
    with open(args.output, "w", encoding="utf-8") as f:
        for h in hypotheses:
            f.write(json.dumps(h, ensure_ascii=False) + "\n")

    print(f"Wrote {len(hypotheses)} hypotheses to {args.output} ({errors} errors)")
    print(
        "Official V1 Accuracy needs evaluate_qa.py. This file is not V2 LAFS. "
        "Prefix --limit is not mixed V1.",
        file=sys.stderr,
    )


if __name__ == "__main__":
    main()