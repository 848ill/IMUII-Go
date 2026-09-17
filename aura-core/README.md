# AURA Core: Pure Go Native Two-Stage RAG Orchestrator 🚀

**AURA** (*Academic Universal Regulatory Assistant*) adalah sistem kecerdasan buatan berbasis *Retrieval-Augmented Generation* (RAG) dua-tahap yang dirancang khusus untuk konsultasi dan kepatuhan regulasi akademik Universitas Islam Indonesia (UII), khususnya Fakultas Teknologi Industri (FTI).

Sistem ini diorkestrasi **100% secara native dalam bahasa Go (Golang)** tanpa ketergantungan pada no-code tool (n8n), Docker container berat, ataupun cloudflare tunnel.

---

## 🏛️ Arsitektur Two-Stage Neural RAG

```
                                      [ PERTANYAAN MAHASISWA ]
                                                 │
                                                 ▼
                             ┌───────────────────────────────────────┐
                             │  Cohere Embeddings Multilingual v3.0  │ (1024-dim dense vector)
                             └───────────────────────────────────────┘
                                                 │
                                                 ▼
                             ┌───────────────────────────────────────┐
                             │  Stage 1: Pinecone ANN Retrieval      │ (Top-20 candidate clauses)
                             └───────────────────────────────────────┘
                                                 │
                                                 ▼
                             ┌───────────────────────────────────────┐
                             │  Stage 2: Cohere Cross-Encoder Rerank │ (Top-8 highly-relevant clauses)
                             └───────────────────────────────────────┘
                                                 │
                                                 ▼
                             ┌───────────────────────────────────────┐
                             │  Prompt Assembler + Citation Grounding│
                             └───────────────────────────────────────┘
                                                 │
                                                 ▼
                             ┌───────────────────────────────────────┐
                             │  DeepSeek-V3 Inference Engine         │ (Anti-hallucination compliance)
                             └───────────────────────────────────────┘
                                                 │
                                                 ▼
                             ┌───────────────────────────────────────┐
                             │  Real-Time SSE Streaming ke Web UI    │ (Typing token stream)
                             └───────────────────────────────────────┘
```

---

## 📁 Struktur Direktori

```
aura-core/
├── go.mod                     # Modul Go independen
├── .env                       # Konfigurasi API keys (DeepSeek, Cohere, Jina, Pinecone, Supabase)
├── cmd/
│   ├── server/                # Web Server & SSE Streaming API (Port 8090)
│   │   └── main.go
│   └── ingest/                # CLI Ingestion Tool untuk batch upload PDF
│       └── main.go
├── pkg/                       # Modul Client API Ringan & Zero Heavy Dependencies
│   ├── jina/                  # Jina Reader API (ekstraksi PDF ke clean Markdown)
│   ├── cohere/                # Cohere API (Embeddings 1024-dim & Rerank v3.5)
│   ├── pinecone/              # Pinecone REST API (Upsert & Top-K Vector Query)
│   ├── deepseek/              # DeepSeek API (Chat & Realtime Token Streaming)
│   ├── chunker/               # Recursive Character Splitter murni Go (1000/200)
│   └── supabase/              # Supabase REST API (Database dokumen & sesi)
├── internal/
│   ├── config/                # Environment loader
│   ├── models/                # Domain data models & telemetry timing
│   ├── rag/                   # Engine Inti Two-Stage RAG (Ingestor, Retriever, Generator)
│   └── api/                   # HTTP Handlers (REST & Server-Sent Events)
└── web/                       # Antarmuka Pengguna Modern & Responsif
    └── templates/
        └── index.html         # Web Chat UI dengan efek ketik real-time & sitasi
```

---

## ⚡ Cara Menjalankan

### 1. Prasyarat
- Go version 1.22 atau lebih baru.
- Isi `PINECONE_API_KEY` pada file `.env`.

### 2. Jalankan Server Web & API
```bash
cd aura-core
go run cmd/server/main.go
```
Buka di browser: **[http://localhost:8090](http://localhost:8090)**

### 3. Jalankan Ingesti Dokumen (CLI)
* **Ingest semua dokumen pending dari Supabase:**
  ```bash
  go run cmd/ingest/main.go --from-supabase
  ```
* **Ingest 1 dokumen via URL:**
  ```bash
  go run cmd/ingest/main.go --url https://link-dokumen.pdf --filename "pedoman_fti.pdf" --title "Pedoman FTI UII"
  ```

---

## 📡 Daftar Endpoint API

| Endpoint | Method | Deskripsi |
| :--- | :--- | :--- |
| `/api/health` | `GET` | Healthcheck & operational telemetry |
| `/api/chat` | `POST` | Tanya jawab sinkron (JSON response) |
| `/api/chat/stream` | `POST` / `GET` | Server-Sent Events (SSE) token streaming real-time |
| `/api/ingest` | `POST` | Trigger ingesti dokumen baru |
| `/` | `GET` | Web Chat UI |
