# LongMemEval locked mixed baseline (internal)

**Not a README number** · **not official V1**.

Official V1 BGE mixed n=500 runs on pin `9bee542` are in [`LONGMEMEVAL.md`](LONGMEMEVAL.md) (run 1 **388/500**, run 2 **384/500**, INTERNAL unpublished). Two runs are **not identical**; do not publish a single %. n=12 / n=60 remain unpublished **improvement baseline**, not V1.

Same 12 question IDs. Reader `gpt-4o-mini`. Judge **`gpt-4o-2024-08-06`**. Retrieve `session_id` = official `conv_id`. Isolated palace per embed mode.

## n=60 remesure — hash-only current kernel after #143+#146+#150+#152 (kernel `77b2839`, 2026-09-14T17:35Z–17:37Z)

Internal locked mixed **n=60** (`testdata/longmemeval_baseline_ids_n60.json`, 10 of each of 6 types). First 12 IDs = n=12 lock. Reader `gpt-4o-mini`. Judge `gpt-4o-2024-08-06`. Isolated palace `:8788`. Health `embed_mode=hash` verified. PersistEmbeddings OFF. Qdrant off. Retrieve timeout 600. **Hash-only** (MiniLM/BGE **not run**; hash ≠ official V1 BGE n=500). **Complete 60/60** (0 generate errors). **Not a README number.** **Not official V1.**

Kernel `77b2839d4a77eedc63b250c346c748a889a61517` (`origin/main` at remesure start, `v1.5.12-36-g77b2839`): [#143](https://github.com/iome-sh/memory/pull/143) `5d3aca5` latest-value clip, [#146](https://github.com/iome-sh/memory/pull/146) `89c1dc0` restaurant tried-count, [#150](https://github.com/iome-sh/memory/pull/150) `76175a0` question_date-ago, [#152](https://github.com/iome-sh/memory/pull/152) `85d44c8` stop-`as` (includes Wave M card [#154](https://github.com/iome-sh/memory/pull/154) and official V1 run 2 card [#153](https://github.com/iome-sh/memory/pull/153)). Compare to n=60 `5154a76` (hash **49/60**; `852ce960` FAIL $350k; restaurants `6aeb4375` FAIL 3) **before** #136/#143/#146/#150/#152.

| Embed | Judge-true | Rate | health `embed_mode` | knowledge-update | multi-session | ss-assistant | ss-preference | ss-user | temporal |
|-------|------------|------|---------------------|------------------|---------------|--------------|---------------|---------|----------|
| hash | **52/60** | 0.867 | `hash` | **9/10** | **9/10** | **10/10** | **6/10** | **10/10** | **8/10** |
| MiniLM ONNX | SKIP | — | not run | — | — | — | — | — | — |
| BAAI BGE ONNX | SKIP | — | not run | — | — | — | — | — | — |

n=12 prefix vs Wave N (`77b2839` hash 11/12): hash **11/12** (miss `0a995998` only; `6aeb4375` PASS). MiniLM/BGE not this remesure — do not invent those columns.

| ID | Role | hash |
|----|------|------|
| `0a995998` | clothes gold 3 | FAIL (summed 2) |
| `gpt4_59c863d7` | kits gold 5 | PASS |
| `aae3761f` | hours gold 15 | PASS |
| `3a704032` | plants | PASS |
| `852ce960` | KU mortgage gold $400k | FAIL ($350k) |
| `6aeb4375` | KU restaurants gold 4 | **PASS** (hyp four) |
| `08f4fc43` | days mass→Ash | FAIL (hyp 30 days; judge) |
| `2a1811e2` | days Holi→mass | PASS |
| `2c63a862` | days until house | PASS |

Clothes (`0a995998`) FAIL: retrieve prepends numbered `Count evidence (3 distinct items)` with `1. [dry-clean]` / `2. [return]` / `3. [pick-up]`; reader still summed 2. KU `852ce960` FAIL: latest-value evidence lists `$400,000` (Nov 30) first then stale `$350,000` (clip [#143](https://github.com/iome-sh/memory/pull/143) kept the later amount); reader still `$350,000`. Does not NLP-supersede. KU `6aeb4375` **PASS** (hyp **four**): latest-first tried-N `[time: 2023-09-30T18:01:00Z] I've tried four different ones so far` then stale “three”; no `[restaurant:korean-style-bbq]` / `[restaurant:if]` / `[restaurant:as]` (after [#152](https://github.com/iome-sh/memory/pull/152)). Days `08f4fc43` FAIL: isolated-palace retrieve appends `text dates 30 days apart (January 2nd → February 1st)`; hyp **30 days** (gold accepts 30/31); judge false. JSONL retrieve text is 400-char truncated.

Miss IDs: `08f4fc43` `0a995998` `35a27287` `75832dbd` `852ce960` `afdc33df` `caf03d32` `gpt4_2312f94c`. vs `5154a76` hash: recovered `0edc2aef` `51a45a95` `6aeb4375` `gpt4_76048e76`; new miss `75832dbd` (pref).

Artifacts (gitignored palaces): worktree `data/baseline-n60-hash-now/` (health, hypotheses JSONL, eval-results-gpt-4o, server logs). Log `/tmp/lme-n60-hash-now.log`. Script table `/tmp/lme-n60-hash-now-table.md`.

## Wave N — restaurant stop-token `as` remesure hash-only after #152 (kernel `77b2839`, 2026-09-14T17:35Z)

Kernel `77b2839d4a77eedc63b250c346c748a889a61517` (`origin/main` at remesure start): [#152](https://github.com/iome-sh/memory/pull/152) `85d44c8` drop leftover `[restaurant:as]` (on [#150](https://github.com/iome-sh/memory/pull/150) / Wave M `7a9b956`; includes Wave M card [#154](https://github.com/iome-sh/memory/pull/154) and official V1 run 2 [#153](https://github.com/iome-sh/memory/pull/153)). **Hash-only** (MiniLM/BGE **not run**). Health `embed_mode=hash` verified. Isolated palace. PersistEmbeddings OFF. Qdrant off. **Not a README number.** **Not official V1.**

| Embed | Judge-true | Rate | health `embed_mode` | multi-session |
|-------|------------|------|---------------------|---------------|
| hash | **11/12** | 0.917 | `hash` | **1/2** (clothes miss, projects pass) |
| MiniLM ONNX | SKIP | — | not run | — |
| BAAI BGE ONNX | SKIP | — | not run | — |

Temporal `gpt4_2487a7cb` (webinar first) PASS: retrieve prepends `Temporal evidence (2 distinct events)` (`two months ago` webinar then `last Saturday` workshop) and full evidence appends T6++ `text dates earliest: two months ago · latest: last Saturday` (hypotheses JSONL retrieve text is 400-char truncated). n=12 temporal is **which-first**, not ago — no `question_date` evidence line (not a TimeTo filter). Clothes (`0a995998`, gold 3) FAIL: retrieve prepends numbered `Count evidence (3 distinct items)` with `1. [dry-clean]` / `2. [return]` / `3. [pick-up]`; reader still summed 2. Projects (`6d550036`, gold 2) PASS (`Count evidence:` without N). Knowledge-update `6aeb4375` (Korean restaurants, gold 4) **PASS** (hyp **four**): latest-first tried-N `[time: 2023-09-30T18:01:00Z] I've tried four different ones so far` then stale “three”; no `[restaurant:korean-style-bbq]` / `[restaurant:if]` / `[restaurant:as]` (Wave M leftover `[restaurant:as]` is **before** #152). Other types 2/2. MiniLM/BGE not this remesure — do not invent those columns.

## Wave M — question_date-ago remesure hash-only after #150 (kernel `7a9b956`, 2026-09-14T17:04Z)

Kernel `7a9b956ebbd79b98019622f7ccf38624f0501d1c` (`origin/main` at remesure start): [#150](https://github.com/iome-sh/memory/pull/150) `76175a0` dated-span ago vs `question_date` (on [#147](https://github.com/iome-sh/memory/pull/147) T6++ / Wave L `d5bec89`; includes Wave L card [#151](https://github.com/iome-sh/memory/pull/151)). **Hash-only** (MiniLM/BGE **not run**; official V1 run 2 on `:8782` expected done — this remesure does not start BGE). Health `embed_mode=hash` verified. Isolated palace. PersistEmbeddings OFF. Qdrant off. **Not a README number.** **Not official V1.**

| Embed | Judge-true | Rate | health `embed_mode` | multi-session |
|-------|------------|------|---------------------|---------------|
| hash | **11/12** | 0.917 | `hash` | **1/2** (clothes miss, projects pass) |
| MiniLM ONNX | SKIP | — | not run | — |
| BAAI BGE ONNX | SKIP | — | not run | — |

Temporal `gpt4_2487a7cb` (webinar first) PASS: retrieve prepends `Temporal evidence (2 distinct events)` (`two months ago` webinar then `last Saturday` workshop) and full evidence appends T6++ `text dates earliest: two months ago · latest: last Saturday` (hypotheses JSONL retrieve text is 400-char truncated). n=12 temporal is **which-first**, not ago — no `question_date` evidence line (not a TimeTo filter). Clothes (`0a995998`, gold 3) FAIL: retrieve prepends numbered `Count evidence (3 distinct items)` with `1. [dry-clean]` / `2. [return]` / `3. [pick-up]`; reader still summed 2. Projects (`6d550036`, gold 2) PASS (`Count evidence:` without N). Knowledge-update `6aeb4375` (Korean restaurants, gold 4) **PASS** (hyp **four**): latest-first tried-N `[time: 2023-09-30T18:01:00Z] I've tried four different ones so far` then stale “three”; no `[restaurant:korean-style-bbq]` / `[restaurant:if]`. Leftover `[restaurant:as]` from an Indian-cuisine turn is still clustered (kernel `7a9b956` is **before** [#152](https://github.com/iome-sh/memory/pull/152) `85d44c8` drop-`as`). Other types 2/2. MiniLM/BGE not this remesure — do not invent those columns.

## Wave L — T6++ order-extrema remesure hash-only after #147 (kernel `d5bec89`, 2026-09-14T09:01Z)

Kernel `d5bec8948ae8ec8ec4afb225edf2f93df4fda3c9` (`origin/main`): [#147](https://github.com/iome-sh/memory/pull/147) `298584c` T6++ temporal-order extrema (on [#146](https://github.com/iome-sh/memory/pull/146) restaurant tried-count). **Hash-only** (MiniLM/BGE **not run**; official V1 run 2 occupies BGE RAM). Health `embed_mode=hash` verified. Isolated palace. PersistEmbeddings OFF. Qdrant off. **Not a README number.** **Not official V1.**

| Embed | Judge-true | Rate | health `embed_mode` | multi-session |
|-------|------------|------|---------------------|---------------|
| hash | **11/12** | 0.917 | `hash` | **1/2** (clothes miss, projects pass) |
| MiniLM ONNX | SKIP | — | not run | — |
| BAAI BGE ONNX | SKIP | — | not run | — |

Temporal `gpt4_2487a7cb` (webinar first) PASS: retrieve prepends `Temporal evidence (2 distinct events)` (`two months ago` webinar then `last Saturday` workshop) and full evidence appends T6++ `text dates earliest: two months ago · latest: last Saturday` (hypotheses JSONL retrieve text is 400-char truncated). Clothes (`0a995998`, gold 3) FAIL: retrieve prepends numbered `Count evidence (3 distinct items)` with `1. [dry-clean]` / `2. [return]` / `3. [pick-up]`; reader still summed 2. Projects (`6d550036`, gold 2) PASS (`Count evidence:` without N). Knowledge-update `6aeb4375` (Korean restaurants, gold 4) **PASS** (hyp **four**): latest-first tried-N `[time: 2023-09-30T18:01:00Z] I've tried four different ones so far` then stale “three”; no `[restaurant:korean-style-bbq]` / `[restaurant:if]`. Leftover `[restaurant:as]` from an Indian-cuisine turn is still clustered. Other types 2/2. MiniLM/BGE not this remesure — do not invent those columns.

## Wave K — restaurant remesure hash-only after #146 (kernel `89c1dc0`, 2026-09-14T08:31Z)

Kernel `89c1dc0bb7d16423a790ba5402c88ead7cb4c5e4` (`origin/main`): [#146](https://github.com/iome-sh/memory/pull/146) restaurant tried-count mentions + cuisine-BBQ not venue (on [#136](https://github.com/iome-sh/memory/pull/136) clusters + [#143](https://github.com/iome-sh/memory/pull/143) clip). **Hash-only** (MiniLM/BGE **not run**; official V1 run 2 occupies BGE RAM). Health `embed_mode=hash` verified. Isolated palace. PersistEmbeddings OFF. Qdrant off. **Not a README number.** **Not official V1.**

| Embed | Judge-true | Rate | health `embed_mode` | multi-session |
|-------|------------|------|---------------------|---------------|
| hash | **11/12** | 0.917 | `hash` | **1/2** (clothes miss, projects pass) |
| MiniLM ONNX | SKIP | — | not run | — |
| BAAI BGE ONNX | SKIP | — | not run | — |

Temporal `gpt4_2487a7cb` (webinar first) PASS. Clothes (`0a995998`, gold 3) FAIL: retrieve prepends numbered `Count evidence (3 distinct items)` with `1. [dry-clean]` / `2. [return]` / `3. [pick-up]`; reader still summed 2. Projects (`6d550036`, gold 2) PASS (`Count evidence:` without N). Knowledge-update `6aeb4375` (Korean restaurants, gold 4) **PASS** (hyp **four**): latest-first tried-N `[time: 2023-09-30T18:01:00Z] I've tried four different ones so far` then stale “three”; no `[restaurant:korean-style-bbq]` / `[restaurant:if]`. Leftover `[restaurant:as]` from an Indian-cuisine turn is still clustered. Other types 2/2. MiniLM/BGE not this remesure — do not invent those columns.

## Wave J — restaurant remesure hash-only after #136+#143 (kernel `5d3aca5`, 2026-09-14T07:12Z)

Kernel `5d3aca5c876da487eea7e56a73ea513209e610dc` (`origin/main`): [#136](https://github.com/iome-sh/memory/pull/136) unique-entity restaurant clusters + [#143](https://github.com/iome-sh/memory/pull/143) latest-value clip. **Hash-only** (MiniLM/BGE **not run**; official V1 run 2 occupies BGE RAM). Health `embed_mode=hash` verified. Isolated palace. PersistEmbeddings OFF. Qdrant off. **Not a README number.** **Not official V1.**

| Embed | Judge-true | Rate | health `embed_mode` | multi-session |
|-------|------------|------|---------------------|---------------|
| hash | **10/12** | 0.833 | `hash` | **1/2** (clothes miss, projects pass) |
| MiniLM ONNX | SKIP | — | not run | — |
| BAAI BGE ONNX | SKIP | — | not run | — |

Temporal `gpt4_2487a7cb` (webinar first) PASS. Clothes (`0a995998`, gold 3) FAIL: retrieve prepends numbered `Count evidence (3 distinct items)` with `1. [dry-clean]` / `2. [return]` / `3. [pick-up]`; reader still summed 2. Projects (`6d550036`, gold 2) PASS (`Count evidence:` without N). Knowledge-update `6aeb4375` (Korean restaurants, gold 4) FAIL (hyp **three**): count-evidence clustered `[restaurant:korean-style-bbq]` / `[restaurant:if]`; haystack still has stale “three” ahead of later “four”. Other types 2/2. MiniLM/BGE not this remesure — do not invent those columns.

## Wave I — skip-vector / clothing-only N / numbered clothes / T6 / T7 (kernel `e094bec` / #122+#124+#127+#129+#132, 2026-09-13T00:39Z–00:44Z)

Kernel `e094bec` (`origin/main` at remesure start): [#122](https://github.com/iome-sh/memory/pull/122) skip-vector + latest-value, [#124](https://github.com/iome-sh/memory/pull/124) clothing-only N, [#127](https://github.com/iome-sh/memory/pull/127) numbered clothes, [#129](https://github.com/iome-sh/memory/pull/129) T6 text-date delta, [#132](https://github.com/iome-sh/memory/pull/132) T7 unique-entity generalize. Health `embed_mode` verified (MiniLM/BGE did **not** hash-fall-back). Isolated palace per embed. **Not a README number.** [#134](https://github.com/iome-sh/memory/pull/134) T4-perf landed **after** this remesure and is not this card. `852ce960` is **not** in the n=12 lock.

| Embed | Judge-true | Rate | health `embed_mode` | multi-session |
|-------|------------|------|---------------------|---------------|
| hash | **10/12** | 0.833 | `hash` | **1/2** (clothes miss, projects pass) |
| MiniLM ONNX | **10/12** | 0.833 | `onnx-minilm-l6-v2` | **1/2** (clothes miss, projects pass) |
| BAAI BGE ONNX | **10/12** | 0.833 | `onnx-bge-small-en-v1.5` | **1/2** (clothes miss, projects pass) |

Temporal `gpt4_2487a7cb` (webinar first) pass all three. Clothes (`0a995998`, gold 3) miss all three: retrieve prepends numbered `Count evidence (3 distinct items)` with `1. [dry-clean]` / `2. [return]` / `3. [pick-up]`; reader still summed 2. Projects (`6d550036`, gold 2) pass all three (`Count evidence:` without N; MiniLM recovered vs Wave H `8 distinct items`). Knowledge-update `6aeb4375` (Korean restaurants, gold 4) miss all three (reader 3; retrieve leads with stale “three”). Other types 2/2. Not official V1.

## n=60 remesure — after #119+#122+#124+#129+#132 (kernel `5154a76`, 2026-09-13T00:54Z–01:19Z)

Internal locked mixed **n=60** (`testdata/longmemeval_baseline_ids_n60.json`, 10 of each of 6 types). First 12 IDs = n=12 lock. Reader `gpt-4o-mini`. Judge `gpt-4o-2024-08-06`. Isolated palace. Health `embed_mode` verified (MiniLM/BGE did **not** hash-fall-back). PersistEmbeddings default OFF. Qdrant off. Retrieve timeout 600; BGE workers 1. **Complete 60/60** (no timeouts). **Not a README number** · **not official V1**.

This kernel is **after** [#119](https://github.com/iome-sh/memory/pull/119)+[#122](https://github.com/iome-sh/memory/pull/122)+[#124](https://github.com/iome-sh/memory/pull/124)+[#129](https://github.com/iome-sh/memory/pull/129)+[#132](https://github.com/iome-sh/memory/pull/132) (unlike v1.5.12 `e90a82d`, which was **before** #119). Worktree HEAD `5154a76395fcf25312ba82dfab4c7494fdcd24b4` (`v1.5.12-17-g5154a76`) includes [#134](https://github.com/iome-sh/memory/pull/134) T4-perf and Wave I docs [#135](https://github.com/iome-sh/memory/pull/135). [#136](https://github.com/iome-sh/memory/pull/136) restaurant clusters landed **after** this remesure and is not this card.

| Embed | Judge-true | Rate | health `embed_mode` | knowledge-update | multi-session | ss-assistant | ss-preference | ss-user | temporal |
|-------|------------|------|---------------------|------------------|---------------|--------------|---------------|---------|----------|
| hash | **49/60** | 0.817 | `hash` | **8/10** | **9/10** | **10/10** | **6/10** | **9/10** | **7/10** |
| MiniLM ONNX | **51/60** | 0.850 | `onnx-minilm-l6-v2` | **8/10** | **9/10** | **10/10** | **6/10** | **10/10** | **8/10** |
| BAAI BGE ONNX | **47/60** | 0.783 | `onnx-bge-small-en-v1.5` | **8/10** | **8/10** | **10/10** | **5/10** | **9/10** | **7/10** |

n=12 prefix vs Wave I (`e094bec` 10/12 all three): hash **10/12** (miss `0a995998` `6aeb4375`); MiniLM **11/12** (miss `6aeb4375` only; clothes pass this run); BGE **10/12** (miss `0a995998` `6aeb4375`). MiniLM clothes pass is reader/judge variance vs Wave I; do not treat Wave I MiniLM clothes miss as closed.

| ID | Role | hash | MiniLM | BGE |
|----|------|------|--------|-----|
| `0a995998` | clothes gold 3 | FAIL (summed 2) | PASS | FAIL (summed 2) |
| `gpt4_59c863d7` | kits gold 5 | PASS | PASS | PASS |
| `aae3761f` | hours gold 15 | PASS | PASS | PASS |
| `3a704032` | plants | PASS | PASS | PASS |
| `852ce960` | KU mortgage gold $400k | FAIL ($350k) | FAIL ($350k) | FAIL ($350k) |
| `6aeb4375` | KU restaurants gold 4 | FAIL (3) | FAIL (3) | FAIL (3) |
| `08f4fc43` | days mass→Ash | FAIL | FAIL | FAIL |
| `2a1811e2` | days Holi→mass | PASS | PASS | PASS |
| `2c63a862` | days until house | PASS | PASS | PASS |

Shared misses: `08f4fc43` `35a27287` `6aeb4375` `852ce960` `afdc33df` `caf03d32` `gpt4_2312f94c`. Hash-only: `0edc2aef`. MiniLM-only: none. BGE-only: `54026fce`.

v1.5.12 `e90a82d` (before #119): hash 47/60 MS 8/10; MiniLM 48/60 MS 7/10; BGE **46/58 incomplete**. This remesure is complete /60.

Artifacts (gitignored palaces): worktree `data/baseline-n60-wave-i/` (health, hypotheses JSONL, eval-results-gpt-4o, server logs). Log `/tmp/lme-n60-wave-i-run.log`. Summary `/tmp/lme-n60-wave-i-summary.md`.

## Wave H — T1 distinct-item count header (kernel `895f255` / #120, 2026-09-12T22:41Z–23:01Z)

Kernel `#120` (`Count evidence (N distinct items):` / temporal `N distinct events`) on `#119`. Health `embed_mode` verified. **Not a README number.** `#122` (latest-value recency) is **not** this kernel.

| Embed | Judge-true | Rate | health `embed_mode` | multi-session |
|-------|------------|------|---------------------|---------------|
| hash | **11/12** | 0.917 | `hash` | **1/2** (clothes miss, projects pass) |
| MiniLM ONNX | **10/12** | 0.833 | `onnx-minilm-l6-v2` | **0/2** (clothes miss, projects miss) |
| BAAI BGE ONNX | **11/12** | 0.917 | `onnx-bge-small-en-v1.5` | **1/2** (clothes miss, projects pass) |

Temporal `gpt4_2487a7cb` (webinar first) pass all three. Clothes (`0a995998`, gold 3) miss all three: retrieve prepends `Count evidence (3 distinct items)` with dry-clean/return/pick-up; reader still summed 2. Projects (`6d550036`, gold 2) pass hash/BGE; MiniLM miss (header `8 distinct items`; reader split Data Mining). Other types 2/2. Not official V1.

## Wave G — T1 unique-entity + temporal dated evidence (kernel `b02abaf` / #119, 2026-09-12T22:02Z–22:36Z)

Kernel `#119` (`AssembleTemporalEvidence` text dates vs ingest Timestamp; unique-entity count clusters). Health `embed_mode` verified. **Not a README number.** `#120` (distinct-item header) is **not** this kernel.

| Embed | Judge-true | Rate | health `embed_mode` | multi-session |
|-------|------------|------|---------------------|---------------|
| hash | **11/12** | 0.917 | `hash` | **1/2** (clothes miss, projects pass) |
| MiniLM ONNX | **11/12** | 0.917 | `onnx-minilm-l6-v2` | **1/2** (clothes miss, projects pass) |
| BAAI BGE ONNX | **11/12** | 0.917 | `onnx-bge-small-en-v1.5` | **1/2** (clothes miss, projects pass) |

Temporal `gpt4_2487a7cb` (webinar first) pass all three. Clothes (`0a995998`, gold 3) miss all three: retrieve still prepends 3 action clusters; reader summed 2. Projects (`6d550036`, gold 2) pass all three. Other types 2/2. Not official V1.

## Wave F — T1 clothing-errand clusters (kernel `aa64dbc` / #115 on #114, 2026-09-12T20:35Z–20:57Z)

Kernel `#115` (`AssembleCountEvidence` clusters dry-clean / return / pick-up) on `#114` (harness-only batch ONNX retrieve scoring). Health `embed_mode` verified. **Not a README number.**

| Embed | Judge-true | Rate | health `embed_mode` | multi-session |
|-------|------------|------|---------------------|---------------|
| hash | **11/12** | 0.917 | `hash` | **2/2** (clothes pass, projects pass) |
| MiniLM ONNX | **12/12** | 1.000 | `onnx-minilm-l6-v2` | **2/2** |
| BAAI BGE ONNX | **12/12** | 1.000 | `onnx-bge-small-en-v1.5` | **2/2** |

Clothes (`0a995998`, gold 3) pass all three (return + pickup exchanged boots + dry-cleaning). Projects (`6d550036`, gold 2) pass all three. Hash miss is temporal `gpt4_2487a7cb` (event order inverted). Other types 2/2. n=12 clothes residual from Wave E is closed. Not official V1.

## n=60 scout — `origin/main` `1f5bb20` (kernel `f99c140` + Wave E docs, 2026-09-12T20:18Z–20:29Z)

Internal locked mixed **n=60** (`testdata/longmemeval_baseline_ids_n60.json`, 10 of each of 6 types). First 12 IDs = n=12 lock. Reader `gpt-4o-mini`. Judge `gpt-4o-2024-08-06`. Isolated palace. Health `embed_mode` verified. **BGE skipped** (too slow; parent runs BGE after kernel PRs). **Not a README number** · **not official V1**.

| Embed | Judge-true | Rate | health `embed_mode` | knowledge-update | multi-session | ss-assistant | ss-preference | ss-user | temporal |
|-------|------------|------|---------------------|------------------|---------------|--------------|---------------|---------|----------|
| hash | **48/60** | 0.800 | `hash` | **9/10** | **7/10** | **10/10** | **6/10** | **10/10** | **6/10** |
| MiniLM ONNX | **47/60** | 0.783 | `onnx-minilm-l6-v2` | **9/10** | **6/10** | **10/10** | **6/10** | **10/10** | **6/10** |
| BAAI BGE ONNX | SKIP | — | not run | — | — | — | — | — | — |

n=12 prefix **did not fully hold** vs Wave E (`f99c140`): Wave E hash **11/12** / MiniLM **12/12**. This scout hash **10/12** / MiniLM **10/12**. Extra miss `gpt4_2487a7cb` (event order inverted) on both; clothes `0a995998` gold 3 also miss on MiniLM this run (counted 2). Projects `6d550036` still pass. Reader/judge variance on the lock; do not treat Wave E MiniLM 12/12 as reproduced.

Miss IDs shared: `08f4fc43` `0a995998` `0edc2aef` `2a1811e2` `2c63a862` `35a27287` `852ce960` `aae3761f` `afdc33df` `caf03d32` `gpt4_2487a7cb` `gpt4_59c863d7`. MiniLM-only: `3a704032` (plants). Hash-only: none.

What n=60 breaks: incomplete multi-session counts (clothes 3, kits 5, driving hours, MiniLM plants 3); temporal date-diff / order; preference-conditioned answers 4/10 both; one stale knowledge-update (`852ce960` $350k vs $400k). ss-user and ss-assistant 10/10.

Artifacts (gitignored palaces): worktree `data/baseline-n60-scout/` (health, hypotheses JSONL, eval-results-gpt-4o, server logs). Summary `/tmp/lme-n60-scout-1f5bb20-summary.md`.

## Wave E — T1 count-assembly (kernel `f99c140` / #111, 2026-09-12T19:43Z–19:57Z)

Kernel `#111` (`AssembleCountEvidence` + palace-wide count-fact union). Health `embed_mode` verified.

| Embed | Judge-true | Rate | health `embed_mode` | multi-session |
|-------|------------|------|---------------------|---------------|
| hash | **11/12** | 0.917 | `hash` | **1/2** (clothes miss, projects pass) |
| MiniLM ONNX | **12/12** | 1.000 | `onnx-minilm-l6-v2` | **2/2** |
| BAAI BGE ONNX | **11/12** | 0.917 | `onnx-bge-small-en-v1.5` | **1/2** (clothes miss, projects pass) |

Clothes (`0a995998`, gold 3) pass on MiniLM (return + pickup exchanged boots + dry-cleaning). Hash counts pickup + dry-clean = 2; BGE counts pickup + return = 2. Projects (`6d550036`, gold 2) pass all three (Marketing Research led-team + Data Mining solo). Other types 2/2. T1 done-when (MiniLM **and** BGE multi-session no longer 0/2) **met**. Not a README number.

## Wave D — T1 overlap-rank + led extract (kernel `2695e02` / #110, 2026-09-12T19:25Z–19:36Z)

Kernel `#110` only. Does **not** include `f99c140` (#111 count-assembly).

| Embed | Judge-true | Rate | health `embed_mode` | multi-session |
|-------|------------|------|---------------------|---------------|
| hash | **9/12** | 0.750 | `hash` | **0/2** (clothes miss, projects miss) |
| MiniLM ONNX | **11/12** | 0.917 | `onnx-minilm-l6-v2` | **1/2** (clothes pass, projects miss) |
| BAAI BGE ONNX | **10/12** | 0.833 | `onnx-bge-small-en-v1.5` | **0/2** (clothes miss, projects miss) |

Clothes (`0a995998`, gold 3) pass on MiniLM only. Projects (`6d550036`, gold 2) fail all three. Hash also misses temporal `gpt4_2487a7cb`. T1 done-when (MiniLM **and** BGE multi-session no longer 0/2) **not met**. Not a README number.

## Wave C — T1 + count-query fact promotion (kernel `a25a883`, 2026-09-12T19:12Z)

| Embed | Judge-true | Rate | multi-session |
|-------|------------|------|---------------|
| hash | **11/12** | 0.917 | **1/2** (clothes pass, projects miss) |
| MiniLM ONNX | **10/12** | 0.833 | **0/2** |
| BAAI BGE ONNX | **11/12** | 0.917 | **1/2** (clothes pass, projects miss) |

Clothes (`0a995998`, gold 3) now counts boots pick-up + return + dry-cleaning on hash and BGE. Projects (`6d550036`, gold 2) still fails: reader sees no “led/leading” count. Other types 2/2. Hash temporal 2/2 this run.

## Wave B — T1 retrieve only (kernel `ef6a3e9`, 2026-09-12T18:52Z)

| Embed | Judge-true | multi-session |
|-------|------------|---------------|
| hash | 9/12 | 0/2 |
| MiniLM ONNX | 10/12 | 0/2 |
| BAAI BGE ONNX | 10/12 | 0/2 |

Retrieve covered all inner haystack sessions; judge still 0/2.

## Wave A — pre-T1 (kernel `cb93b08`, 2026-09-12T17:33Z)

| Embed | Judge-true | multi-session |
|-------|------------|---------------|
| hash | 9/12 | 0/2 |
| MiniLM ONNX | 10/12 | 0/2 |
| BAAI BGE ONNX | 10/12 | 0/2 |

## Protocol

| Field | Value |
|-------|--------|
| Slice | locked mixed n=12 (`testdata/longmemeval_baseline_ids.json`) |
| Histogram | 2 of each of 6 types |
| BGE source | `BAAI/bge-small-en-v1.5` hugot layout |

```bash
make longmemeval-baseline
```

## Notes

- hash-overlap unpublished · n=12 is not overall V1 · not a README number
- Official V1 BGE mixed n=500 runs on pin `9bee542` are in [`LONGMEMEVAL.md`](LONGMEMEVAL.md) (run 1 **388/500**, run 2 **384/500**, INTERNAL unpublished). Two runs are **not identical**; do not publish a single %. n=12 / n=60 remain unpublished improvement baseline, not V1.
- TTFH / cite-both walking skeleton is a different clock
- Wave N is kernel `77b2839` (#152 restaurant stop-token `as` on #150 / Wave M). Hash-only **11/12** MS **1/2**. Clothes `0a995998` FAIL (numbered 1–3 + N=3 in retrieve; reader summed 2). Projects `6d550036` PASS. KU `6aeb4375` **PASS** (hyp four; latest-first tried-N; no korean-style-bbq/if/`as`). Temporal `gpt4_2487a7cb` PASS (`text dates earliest: two months ago · latest: last Saturday`; which-first, not ago). MiniLM/BGE not this remesure. Not a README number.
- Wave M is kernel `7a9b956` (#150 dated-span ago vs `question_date` on #147 / Wave L). Hash-only **11/12** MS **1/2**. Clothes `0a995998` FAIL (numbered 1–3 + N=3 in retrieve; reader summed 2). Projects `6d550036` PASS. KU `6aeb4375` **PASS** (hyp four; latest-first tried-N; no korean-style-bbq/if). Temporal `gpt4_2487a7cb` PASS (`text dates earliest: two months ago · latest: last Saturday`; which-first, not ago). MiniLM/BGE not this remesure. Not a README number.
- Wave L is kernel `d5bec89` (#147 T6++ order extrema on #146). Hash-only **11/12** MS **1/2**. Clothes `0a995998` FAIL (numbered 1–3 + N=3 in retrieve; reader summed 2). Projects `6d550036` PASS. KU `6aeb4375` **PASS** (hyp four; latest-first tried-N; no korean-style-bbq/if). Temporal `gpt4_2487a7cb` PASS (`text dates earliest: two months ago · latest: last Saturday`). MiniLM/BGE not this remesure. Not a README number.
- Wave K is kernel `89c1dc0` (#146 tried-count + cuisine-BBQ not venue). Hash-only **11/12** MS **1/2**. Clothes `0a995998` FAIL (numbered 1–3 + N=3 in retrieve; reader summed 2). Projects `6d550036` PASS. KU `6aeb4375` **PASS** (hyp four; latest-first tried-N; no korean-style-bbq/if). Temporal `gpt4_2487a7cb` PASS. MiniLM/BGE not this remesure. Not a README number.
- Wave J is kernel `5d3aca5` (#136 restaurant clusters + #143 latest-value clip). Hash-only **10/12** MS **1/2**. Clothes `0a995998` FAIL (numbered 1–3 + N=3 in retrieve; reader summed 2). Projects `6d550036` PASS. KU `6aeb4375` FAIL (3 vs gold 4). Temporal `gpt4_2487a7cb` PASS. MiniLM/BGE not this remesure. Not a README number.
- Wave I is kernel `e094bec` (#122+#124+#127+#129+#132). Hash/MiniLM/BGE **10/12** MS **1/2**. Clothes `0a995998` miss all three (numbered 1–3 + N=3 in retrieve; reader summed 2). Projects `6d550036` pass all three (MiniLM recovered). KU `6aeb4375` miss all three (3 vs gold 4). `852ce960` not in n=12. #134 not this remesure. Not a README number.
- n=60 remesure after #119+#122+#124+#129+#132 is kernel `5154a76` (includes #134+#135; **before** #136). hash **49/60** MS **9/10**; MiniLM **51/60** MS **9/10**; BGE **47/60** MS **8/10**. Complete 60/60. Unlike v1.5.12 `e90a82d` (before #119). Not a README number.
