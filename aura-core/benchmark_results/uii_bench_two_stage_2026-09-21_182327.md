# Hasil Evaluasi UII-Bench-50 (Automated RAG Triad Scoring)

## Tabel Skor Per-Skenario

| ID | Cluster | Nama Kluster | Query | CR | G | AR | RTS | Total (ms) |
|---|---|---|---|---|---|---|---|---|
| CS01 | 1 | Prasyarat Seminar Proposal Skripsi | Saya mahasiswa Informatika FTI UII semester 6 deng… | 0.9000 | 0.9000 | 1.0000 | 0.9310 | 7181 |
| CS02 | 1 | Prasyarat Seminar Proposal Skripsi | Total SKS lulus saya 120 SKS dengan IPK 2.80, tapi… | 0.9000 | 0.9000 | 1.0000 | 0.9310 | 7680 |
| CS03 | 1 | Prasyarat Seminar Proposal Skripsi | Saya semester 7 dengan 114 SKS lulus dan IPK 2.65,… | 0.9000 | 0.9000 | 1.0000 | 0.9310 | 6964 |
| CS04 | 2 | Beban SKS Semester & IPS | Pada semester genap lalu IPS saya memperoleh 2.85.… | 1.0000 | 0.9000 | 1.0000 | 0.9643 | 4519 |
| CS05 | 2 | Beban SKS Semester & IPS | IPS semester lalu saya 3.90. Bolehkah saya mengamb… | 0.9000 | 1.0000 | 1.0000 | 0.9643 | 4035 |
| CS06 | 2 | Beban SKS Semester & IPS | Berapa maksimal SKS yang boleh diambil mahasiswa p… | 0.0000 | 1.0000 | 0.7000 | 0.0000 | 6721 |
| CS07 | 3 | Masa Berlaku SK & Dosen Pembimbing | SK Dosen Pembimbing Skripsi saya sudah terbit 2 se… | 0.9000 | 0.9000 | 1.0000 | 0.9310 | 6144 |
| CS08 | 3 | Masa Berlaku SK & Dosen Pembimbing | Berapa frekuensi minimal bimbingan dengan Dosen Pe… | 0.2000 | 1.0000 | 0.9000 | 0.4219 | 4655 |
| CS09 | 3 | Masa Berlaku SK & Dosen Pembimbing | Draf proposal skripsi saya sudah selesai tapi pemb… | 0.9000 | 0.7500 | 1.0000 | 0.8710 | 6347 |
| CS10 | 4 | Batas Turnitin & Plagiarisme | Hasil cek Turnitin naskah skripsi saya menunjukkan… | 0.9000 | 0.9000 | 1.0000 | 0.9310 | 5531 |
| CS11 | 4 | Batas Turnitin & Plagiarisme | Pengaturan apa saja yang harus diaktifkan saat mel… | 0.9000 | 0.9000 | 1.0000 | 0.9310 | 5059 |
| CS12 | 4 | Batas Turnitin & Plagiarisme | Apa sanksi akademik jika mahasiswa terbukti melaku… | 0.4000 | 0.9000 | 0.7000 | 0.5953 | 6145 |
| CS13 | 5 | Yudisium & Evaluasi Batas Studi | Saya sudah mengumpulkan 144 SKS lulus dengan IPK 2… | 0.9000 | 0.9000 | 1.0000 | 0.9310 | 6466 |
| CS14 | 5 | Yudisium & Evaluasi Batas Studi | Seluruh nilai skripsi dan SKS saya sudah lengkap 1… | 0.9000 | 0.9000 | 1.0000 | 0.9310 | 7491 |
| CS15 | 5 | Yudisium & Evaluasi Batas Studi | Saya sekarang berada di akhir semester 4 dengan to… | 0.9000 | 0.9000 | 1.0000 | 0.9310 | 7848 |

## Agregat Per-Kluster

| Kluster | N | Mean CR | Mean G | Mean AR | Mean RTS |
|---|---|---|---|---|---|
| Kluster 1 | 3 | 0.9000 | 0.9000 | 1.0000 | 0.9310 |
| Kluster 2 | 3 | 0.6333 | 0.9667 | 0.9000 | 0.6429 |
| Kluster 3 | 3 | 0.6667 | 0.8833 | 0.9667 | 0.7413 |
| Kluster 4 | 3 | 0.7333 | 0.9000 | 0.9000 | 0.8191 |
| Kluster 5 | 3 | 0.9000 | 0.9000 | 1.0000 | 0.9310 |
| **Rata-rata Keseluruhan** | **15** | **0.7667** | **0.9100** | **0.9533** | **0.8131** |

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
