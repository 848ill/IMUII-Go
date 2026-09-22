# BAB V
# KESIMPULAN DAN SARAN

## 5.1 Kesimpulan

Berdasarkan keseluruhan tahapan penelitian, perancangan arsitektur, implementasi, serta serangkaian evaluasi kuantitatif dan kualitatif yang telah dilaksanakan pada sistem **AURA Core** (*Academic Universal Regulatory Assistant*), dapat ditarik kesimpulan sebagai berikut:

1. **Keberhasilan Rancang Bangun Arsitektur *Pure-Go Native Two-Stage RAG***:  
   Sistem AURA Core berhasil dibangun secara berdaulat (*sovereign*) menggunakan bahasa pemrograman Go (Go 1.24+ standard library) tanpa ketergantungan pada kerangka kerja eksternal yang berat seperti LangChain atau LlamaIndex. Arsitektur ini mengintegrasikan tahap penelusuran awal *Dense Vector ANN Retrieval* (Top-20 kandidat) dengan tahap penyaringan lanjutan *Cross-Encoder Neural Reranking* (Top-8 klausul). Implementasi murni berbasis Go terbukti sangat efisien dengan jejak memori peladen (*memory footprint*) stabil di bawah 45 MB dan latensi median *end-to-end* sebesar 4.24 detik, menjadikannya sangat andal untuk melayani beban kueri konkuren di peladen institusi.

2. **Jaminan Privasi Data Akademik Mahasiswa Melalui *In-Memory Parsing***:  
   Mekanisme isolasi privasi data mahasiswa berhasil diimplementasikan melalui pendekatan *in-memory zero-persistence*. Dokumen transkrip privat (KHS dan KRS) diproses strictly di dalam memori kerja (RAM) per sesi transaksi aktif. Data akademik mahasiswa tidak pernah diubah menjadi vektor semantik dan **tidak pernah disimpan ke dalam basis data vektor publik bersama**, sehingga sepenuhnya meniadakan risiko kebocoran data akademik privat lintas pengguna.

3. **Bukti Empiris Keunggulan *Two-Stage RAG* melalui Studi Ablasi**:  
   Pengujian studi ablasi (*ablation study*) membuktikan secara ilmiah bahwa penambahan tahap *Cross-Encoder Neural Reranking* merupakan komponen krusial dalam mengeliminasi halusinasi regulasi. Dibandingkan dengan arsitektur *Single-Stage RAG* (yang hanya mencapai nilai *Groundedness* 81.22% dan gagal memenuhi batas aman institusional), arsitektur *Two-Stage RAG* AURA Core menghasilkan lonjakan *Groundedness* sebesar **+10.58% menjadi 91.80%** serta peningkatan skor komposit *RAG Triad Score* (RTS) sebesar **+4.38% menjadi 90.55%**. Lonjakan akurasi faktual ini dicapai dengan penambahan biaya latensi komputasi yang sangat efisien, yaitu hanya sekitar 330 milidetik.

4. **Pencapaian Kinerja Kuantitatif dan Ketahanan Kualitatif Sistem**:  
   Pada evaluasi kuantitatif dataset *UII-Bench-50* yang mencakup 50 skenario uji di 5 klaster regulasi FTI UII, AURA Core mencatatkan performa rata-rata yang melampaui seluruh target ideal skripsi:
   - *Context Relevance* (CR): **87.40%** (Target: $\ge 70.0\%$).
   - *Groundedness / Anti-Halusinasi* (G): **91.80%** (Target: $\ge 85.0\%$).
   - *Answer Relevance* (AR): **98.60%** (Target: $\ge 90.0\%$).
   - *RAG Triad Score* (RTS): **90.55%** (Target: $\ge 80.0\%$).  
   Selain itu, pengujian kualitatif terhadap 15 studi kasus riil mahasiswa (CS01–CS15) membuktikan ketahanan sistem dalam menegakkan batas hukum absolut (seperti penolakan 26 SKS pada IPS 3.90), ketegasan menolak upaya manipulasi aturan fiktif (injeksi surat rekomendasi RT/RW), kecermatan interpretasi status nilai E yang diulang, serta disiplin menolak spekulasi pada domain di luar regulasi FTI UII (*out-of-domain isolation*).

---

## 5.2 Saran

Untuk penyempurnaan dan pengembangan sistem AURA Core pada penelitian selanjutnya, beberapa saran teknis dan strategis yang direkomendasikan adalah:

1. **Implementasi Penelusuran Hibrida (*Hybrid Lexical-Dense Retrieval*)**:  
   Disarankan untuk mengintegrasikan algoritma pencarian leksikal tradisional (seperti BM25 atau SPLADE) yang digabungkan dengan *Dense ANN Vector Search* melalui mekanisme *Reciprocal Rank Fusion* (RRF) pada Tahap 1. Hal ini akan semakin meningkatkan nilai *Context Relevance* (CR) pada kueri mahasiswa yang memuat singkatan langka, kode mata kuliah spesifik, atau istilah administratif lokal yang tidak terpetakan secara optimal dalam model embedding semantik global.

2. **Fitur Ekspor Bukti Konsultasi Digital Terverifikasi (*Cryptographic Consultation Export*)**:  
   Sistem dapat dikembangkan lebih lanjut dengan menambahkan modul pembangkit berkas PDF otomatis yang memuat rangkuman hasil konsultasi regulasi mahasiswa. Berkas tersebut dapat dilengkapi dengan tanda tangan digital berbasis stempel kriptografis (*HMAC-SHA256 signature* atau *QR Code verification*) yang dapat dijadikan bukti pertimbangan awal saat mahasiswa mengajukan permohonan dispensasi resmi ke pihak fakultas.

3. **Integrasi Antarmuka Terprogram (*API Gateway*) dengan Sistem Informasi Akademik Kampus**:  
   Untuk mempermudah mahasiswa tanpa perlu mengunggah berkas KHS/KRS secara manual, penelitian masa depan dapat mengintegrasikan AURA Core langsung ke sistem informasi akademik universitas (seperti UIIRAS / Gateway Akademik UII) melalui protokol otorisasi terpusat (OAuth 2.0 / SAML) dengan tetap mempertahankan standar enkripsi ujung-ke-ujung (*end-to-end encryption*).
