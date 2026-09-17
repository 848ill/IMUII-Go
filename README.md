---
title: AURA UII
emoji: 🎓
colorFrom: green
colorTo: blue
sdk: docker
app_port: 7860
pinned: false
---

# AURA UII — Academic Universal Regulatory Assistant 🎓

Sistem kecerdasan buatan berbasis **Two-Stage RAG** (Retrieval-Augmented Generation) untuk konsultasi regulasi akademik Universitas Islam Indonesia (UII).

## Arsitektur

```
Query → Cohere Embed (1024-dim) → Pinecone ANN Top-20 → Cohere Rerank Top-8 → DeepSeek-V3 → SSE Stream
```

## Fitur Utama

- **Zero-Hallucination Grounding** — menolak menjawab jika tidak ada di basis regulasi resmi
- **DPA Virtual** — evaluasi kelayakan studi mahasiswa (KRS/KHS cross-referencing)
- **Real-time SSE Streaming** — respons huruf-demi-huruf
- **UII-Bench-50** — automated evaluation runner dengan LLM-as-Judge

## Tech Stack

Go 1.27 • DeepSeek-V3 • Cohere Multilingual v3.0 • Pinecone Serverless • Supabase
