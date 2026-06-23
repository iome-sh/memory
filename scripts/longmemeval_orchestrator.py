#!/usr/bin/env python3
"""
LongMemEval Orchestrator for github.com/sudo-jin/memory

This script drives the Go HTTP server (cmd/longmemeval-server) to benchmark
the PalaceStore + RecMem features on the official LongMemEval dataset.

Usage (after starting the Go server):
    python scripts/longmemeval_orchestrator.py --dataset data/longmemeval_s_cleaned.json --output hypotheses.jsonl

Then run the official evaluator from the LongMemEval repo.
"""

import argparse
import json
import os
import sys
from typing import Any, Dict, List

import requests
from tqdm import tqdm

try:
    import openai
except ImportError:
    print("pip install openai tqdm requests", file=sys.stderr)
    sys.exit(1)


SERVER_URL = os.environ.get("LONGMEMEVAL_SERVER", "http://localhost:8765")
OPENAI_MODEL = os.environ.get("OPENAI_MODEL", "gpt-4o-mini")


def check_server() -> bool:
    try:
        r = requests.get(f"{SERVER_URL}/health", timeout=5)
        return r.status_code == 200
    except Exception as e:
        print(f"Server not reachable at {SERVER_URL}: {e}", file=sys.stderr)
        return False


def ingest_history(conv_id: str, history: List[Dict[str, Any]]):
    """Send conversation turns to the Go memory server."""
    turns = []
    for turn in history:
        turns.append({
            "role": turn.get("role", "user"),
            "content": turn.get("content", ""),
            "timestamp": turn.get("timestamp", ""),
            "cycle": turn.get("cycle", 0),
        })
    payload = {"conv_id": conv_id, "turns": turns}
    r = requests.post(f"{SERVER_URL}/ingest", json=payload, timeout=30)
    r.raise_for_status()


def retrieve_memories(query: str, k: int = 40) -> List[Dict[str, Any]]:
    payload = {"query": query, "limit": k}
    r = requests.post(f"{SERVER_URL}/retrieve", json=payload, timeout=30)
    r.raise_for_status()
    return r.json().get("memories", [])


def summarize_memories(memories: List[Dict[str, Any]]) -> str:
    """Post-retrieval LLM summarization for richer context assembly (EmergenceMem/Zep style)."""
    if not memories:
        return "No relevant memories retrieved."
    # Build rich context for summarizer
    context_for_summary = "\n".join(
        f"- Memory {m.get('id', 'unknown')}: {m.get('full', m.get('summary', ''))}"
        for m in memories[:10]  # limit for token
    )
    client = openai.OpenAI()
    resp = client.chat.completions.create(
        model=OPENAI_MODEL,
        messages=[{"role": "user", "content": f"Extract and list the key personal facts, entities, quantities, events, and preferences from these memories in concise bullet points. Focus on information relevant to answering personal questions:\n{context_for_summary}"}],
        temperature=0.0,
        max_tokens=400,
    )
    return resp.choices[0].message.content.strip()


def generate_answer(question: str, memories: List[Dict[str, Any]]) -> str:
    """Use OpenAI to answer using retrieved memories as context."""
    # Improved context_text prioritizing full + tags + scores (line ~68 area)
    context_text = "\n".join(
        f"[Score: {m.get('score', 0.0):.2f}] ID: {m.get('id', 'N/A')} Full: {m.get('full', m.get('summary', ''))} Tags: {m.get('tags', [])}"
        for m in memories
    )
    summary = summarize_memories(memories)
    prompt = (
        "You are a helpful assistant with access to long-term memory. "
        "Use the provided memories and summary to answer the question as accurately as possible. "
        "If the memories contain partial information, synthesize the most accurate answer possible; only abstain if truly absent.\n\n"
        f"Memories:\n{context_text}\n\nSummary of Key Facts:\n{summary}\n\nQuestion: {question}\n\nAnswer:"
    )
    client = openai.OpenAI()
    resp = client.chat.completions.create(
        model=OPENAI_MODEL,
        messages=[{"role": "user", "content": prompt}],
        temperature=0.0,
        max_tokens=400,
    )
    return resp.choices[0].message.content.strip()


def load_dataset(path: str) -> List[Dict[str, Any]]:
    with open(path, "r", encoding="utf-8") as f:
        data = json.load(f)
    if isinstance(data, dict) and "examples" in data:
        return data["examples"]
    if isinstance(data, list):
        return data
    for key in ["data", "items", "questions"]:
        if key in data:
            return data[key]
    return data


def haystack_history(example: Dict[str, Any]) -> List[Dict[str, Any]]:
    """Flatten official LongMemEval haystack_sessions into ingest turns."""
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


def answer_in_memories(answer: str, memories: List[Dict[str, Any]], topk: int) -> bool:
    if not answer or not memories:
        return False
    answer_lower = answer.strip().lower()
    corpus = " ".join(
        (m.get("full") or m.get("summary") or "").lower()
        for m in memories[:topk]
    )
    if answer_lower in corpus:
        return True
    tokens = [
        w
        for w in "".join(c if c.isalnum() else " " for c in answer_lower).split()
        if len(w) >= 3
    ]
    if not tokens:
        return False
    matched = sum(1 for t in tokens if t in corpus)
    threshold = len(tokens) if len(tokens) == 1 else (len(tokens) + 1) // 2
    return matched >= threshold


def main():
    parser = argparse.ArgumentParser(description="LongMemEval harness for sudo-jin/memory")
    parser.add_argument("--dataset", required=True, help="Path to longmemeval_s_cleaned.json or similar")
    parser.add_argument("--output", default="hypotheses.jsonl", help="Output hypothesis file")
    parser.add_argument("--limit", type=int, default=0, help="Limit number of examples (0 = all)")
    parser.add_argument(
        "--recall-only",
        action="store_true",
        help="Skip OpenAI answer generation; ingest+retrieve and print recall stats only",
    )
    parser.add_argument("--topk", type=int, default=5, help="Top-k memories for recall-only hit check")
    args = parser.parse_args()

    if not check_server():
        print("Start the Go server first: go run cmd/longmemeval-server/main.go", file=sys.stderr)
        sys.exit(1)

    examples = load_dataset(args.dataset)
    if args.limit > 0:
        examples = examples[: args.limit]

    hypotheses = []
    recall_hits = 0
    recall_total = 0

    for ex in tqdm(examples, desc="Processing examples"):
        qid = ex.get("question_id") or ex.get("id") or ex.get("qid")
        question = ex.get("question") or ex.get("query")
        history = haystack_history(ex)
        conv_id = ex.get("conv_id") or ex.get("conversation_id") or qid
        oracle = ex.get("answer") or ""

        if not qid or not question:
            print(f"Skipping example without question_id/question: {ex.keys()}", file=sys.stderr)
            continue

        try:
            ingest_history(str(conv_id), history)
        except Exception as e:
            print(f"Ingest error for {qid}: {e}", file=sys.stderr)
            continue

        memories = retrieve_memories(question, k=max(args.topk, 40))

        if args.recall_only:
            recall_total += 1
            hit = answer_in_memories(oracle, memories, args.topk)
            if hit:
                recall_hits += 1
            status = "HIT" if hit else "MISS"
            print(f"[{status}] {qid} answer={oracle!r}")
            continue

        try:
            answer = generate_answer(question, memories)
        except Exception as e:
            print(f"Generation error for {qid}: {e}", file=sys.stderr)
            answer = "[ERROR]"

        hypotheses.append({"question_id": qid, "hypothesis": answer})

    if args.recall_only:
        recall = (recall_hits / recall_total) if recall_total else 0.0
        print(f"\naggregate recall: {recall_hits}/{recall_total} = {recall:.2f}")
        return

    with open(args.output, "w", encoding="utf-8") as f:
        for h in hypotheses:
            f.write(json.dumps(h, ensure_ascii=False) + "\n")

    print(f"\nWrote {len(hypotheses)} hypotheses to {args.output}")
    print("Now run the official evaluator (adjust paths to match your layout):")
    print("  cd LongMemEval/src/evaluation")
    print(f"  python evaluate_qa.py gpt-4o ../../../{args.output} ../../../data/longmemeval_oracle.json")


if __name__ == "__main__":
    main()
