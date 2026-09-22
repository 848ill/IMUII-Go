# Hasil Evaluasi UII-Bench-50 (Automated RAG Triad Scoring)

## Tabel Skor Per-Skenario

| ID | Cluster | Nama Kluster | Query | CR | G | AR | RTS | Total (ms) |
|---|---|---|---|---|---|---|---|---|
| S01 | 1 | Prasyarat Seminar Proposal | Berapa minimal SKS yang harus sudah lulus untuk bi… | 1.0000 | 1.0000 | 1.0000 | 1.0000 | 9827 |
| S02 | 1 | Prasyarat Seminar Proposal | Saya punya satu nilai E di semester 2 dulu karena … | 0.9000 | 0.9000 | 1.0000 | 0.9310 | 4771 |
| S03 | 1 | Prasyarat Seminar Proposal | IPK saya sekarang 1.99, kurang dari 2.00. Apakah s… | 0.9000 | 0.9000 | 1.0000 | 0.9310 | 5083 |
| S04 | 1 | Prasyarat Seminar Proposal | Saya belum mengambil mata kuliah Metodologi Peneli… | 1.0000 | 1.0000 | 1.0000 | 1.0000 | 4712 |
| S05 | 1 | Prasyarat Seminar Proposal | Saya sudah lulus Metodologi Penelitian dengan nila… | 0.9000 | 1.0000 | 1.0000 | 0.9643 | 4158 |
| S06 | 1 | Prasyarat Seminar Proposal | Draf proposal saya sudah selesai tapi dosen pembim… | 0.9000 | 1.0000 | 1.0000 | 0.9643 | 4106 |
| S07 | 1 | Prasyarat Seminar Proposal | Saya sudah 115 SKS lulus, IPK 2.30, tidak ada nila… | 0.9000 | 0.9000 | 1.0000 | 0.9310 | 4923 |
| S08 | 1 | Prasyarat Seminar Proposal | Apa saja dokumen yang harus saya siapkan untuk men… | 0.6000 | 1.0000 | 0.9000 | 0.7941 | 4412 |
| S09 | 1 | Prasyarat Seminar Proposal | Apakah batas minimal IPK untuk sempro berbeda anta… | 0.6000 | 1.0000 | 1.0000 | 0.8182 | 3467 |
| S10 | 1 | Prasyarat Seminar Proposal | Saya sudah 110 SKS lulus persis, IPK 2.05, tidak a… | 0.9000 | 0.9000 | 1.0000 | 0.9310 | 5868 |

## Agregat Per-Kluster

| Kluster | N | Mean CR | Mean G | Mean AR | Mean RTS |
|---|---|---|---|---|---|
| Kluster 1 | 10 | 0.8600 | 0.9600 | 0.9900 | 0.9265 |
| **Rata-rata Keseluruhan** | **10** | **0.8600** | **0.9600** | **0.9900** | **0.9265** |

## Penjelasan & Indikator Metrik RAG Triad

| Metrik | Singkatan | Nama Lengkap | Definisi & Indikator | Rentang Nilai | Target Ideal Skripsi |
|---|---|---|---|---|---|
| **CR** | CR | *Context Relevance* | Mengukur apakah pasal/dokumen yang diretrieve oleh Pinecone & Reranker benar-benar relevan dengan pertanyaan mahasiswa. | 0.00 – 1.00 (0–100%) | $\ge 0.70$ (70%) |
| **G** | G | *Groundedness* (Anti-Halusinasi) | Mengukur kesetiaan klaim jawaban terhadap dokumen rujukan resmi. Nilai 1.00 berarti 100% fakta bersumber dari regulasi tanpa halusinasi. | 0.00 – 1.00 (0–100%) | $\ge 0.85$ (85%) |
| **AR** | AR | *Answer Relevance* | Mengukur seberapa tepat, tuntas, dan langsung jawaban sistem dalam menyelesaikan masalah pertanyaan mahasiswa tanpa bertele-tele. | 0.00 – 1.00 (0–100%) | $\ge 0.90$ (90%) |
| **RTS** | RTS | *RAG Triad Score* | Skor komposit holistik RAG yang dihitung dari rata-rata harmonik: $\frac{3}{\frac{1}{\text{CR}} + \frac{1}{\text{G}} + \frac{1}{\text{AR}}}$. | 0.00 – 1.00 (0–100%) | $\ge 0.80$ (80%) |

### Indikator Latensi Pipeline (Pure Go Native Orchestrator)
- **EmbedMs**: Waktu embedding teks query menjadi vektor 1024-dimensi oleh Cohere Multilingual v3.
- **DenseMs**: Waktu pencarian kemiripan vektor ANN (*Approximate Nearest Neighbor*) Top-20 di Pinecone Vector Database.
- **RerankMs**: Waktu penyortiran ulang akurasi tinggi Top-8 kandidat oleh Cohere Cross-Encoder Neural Reranker v3.5.
- **GenMs**: Waktu inferensi teks dan penalaran anti-halusinasi oleh LLM DeepSeek.
- **TotalMs**: Total waktu eksekusi pipeline end-to-end dari query diterima hingga respons selesai (*wall-clock time*).
