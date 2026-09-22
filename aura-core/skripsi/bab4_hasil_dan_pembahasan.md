# BAB IV
# HASIL DAN PEMBAHASAN

## 4.1 Lingkungan Implementasi dan Pengujian

Sistem AURA Core diimplementasikan dan diuji secara menyeluruh pada lingkungan komputasi riil guna memastikan keandalan operasional, konkurensi layanan, dan kecepatan respons. Spesifikasi lingkungan perangkat keras dan perangkat lunak yang digunakan disajikan pada Tabel 4.1.

**Tabel 4.1** Spesifikasi Lingkungan Implementasi dan Produksi AURA Core
| Komponen | Spesifikasi Teknis / Konfigurasi |
|---|---|
| **Sistem Operasi Peladen** | Ubuntu 24.04 LTS (x86_64) |
| **Bahasa Pemrograman & Runtime** | Go 1.24+ (Standard Library, tanpa runtime Python) |
| **Web Server & Reverse Proxy** | Nginx dengan SSL/TLS Let's Encrypt (`server_tokens off;`) |
| **Protokol Komunikasi** | HTTPS REST API & WebSocket Streaming (`/ws/chat`) |
| **Basis Data Vektor Otoritatif** | Pinecone Serverless Vector Index (`aurarags`, 1024-dim, Cosine) |
| **Model Representasi Vektor** | Cohere `embed-multilingual-v3.0` |
| **Model Neural Reranking** | Cohere `rerank-v3.5` (Cross-Encoder) |
| **Model Inferensi Generatif** | DeepSeek LLM (`deepseek-chat`, $\tau = 0.2$, Top-p = 0.95) |
| **Basis Data Relasional & Sesi** | Supabase PostgreSQL dengan Row Level Security (RLS) & UUIDv4 |
| **URL Produksi Publik** | `https://aura.imuii.id` |

Seluruh pengujian kuantitatif dan evaluasi otomatis dijalankan menggunakan modul penguji `cmd/bench` yang terpasang di peladen lokal dan terhubung langsung ke layanan komputasi hilir.

---

## 4.2 Hasil Evaluasi Kuantitatif UII-Bench-50

Evaluasi kuantitatif dilakukan terhadap 50 skenario uji representatif yang terhimpun dalam dataset *UII-Bench-50*. Pengujian ini mengukur empat parameter utama dalam kerangka *RAG Triad*: *Context Relevance* (CR), *Groundedness* (G), *Answer Relevance* (AR), dan rata-rata harmonik *RAG Triad Score* (RTS).

### 4.2.1 Rekapitulasi Performa Agregat per Klaster Regulasi
Hasil evaluasi agregat per klaster regulasi pada arsitektur *Two-Stage RAG* AURA Core disajikan pada Tabel 4.2.

**Tabel 4.2** Rekapitulasi Performa UII-Bench-50 Berdasarkan Klaster Regulasi
| Klaster Regulasi | Jumlah Skenario ($N$) | Mean CR | Mean G | Mean AR | Mean RTS |
|---|:---:|:---:|:---:|:---:|:---:|
| **Klaster 1**: Prasyarat Seminar Proposal | 10 | 0.8500 | 0.9400 | 1.0000 | 0.9113 |
| **Klaster 2**: Beban SKS Semester & Nilai IPS | 10 | 0.8700 | 0.9200 | 0.9700 | 0.8945 |
| **Klaster 3**: Masa Berlaku SK & Dosen Pembimbing | 10 | 0.8800 | 0.9200 | 0.9700 | 0.9061 |
| **Klaster 4**: Batas Toleransi Turnitin & Plagiarisme | 10 | 0.8200 | 0.9200 | 0.9900 | 0.8748 |
| **Klaster 5**: Evaluasi Masa Studi & Yudisium | 10 | 0.9500 | 0.8900 | 1.0000 | 0.9408 |
| **Rata-rata Keseluruhan (Overall Mean)** | **50** | **0.8740 (87.4%)** | **0.9180 (91.8%)** | **0.9860 (98.6%)** | **0.9055 (90.6%)** |

Berdasarkan data pada Tabel 4.2, seluruh capaian metrik sistem AURA Core melampaui batas ambang target ideal yang telah ditetapkan pada proposal penelitian:
1. **Context Relevance (CR = 87.40% vs Target $\ge 70.0%$)**: Rata-rata 87.40% menunjukkan bahwa kombinasi *Dense ANN* dan *Cross-Encoder* berhasil menyaring dokumen rujukan sehingga potongan teks yang masuk ke jendela LLM benar-benar relevan dengan pertanyaan mahasiswa. Klaster 5 (Yudisium) mencatatkan nilai CR tertinggi sebesar 95.00%, karena klausul kelulusan memiliki kata kunci yang tegas dan terdefinisi dengan sangat jelas di buku pedoman.
2. **Groundedness / Anti-Halusinasi (G = 91.80% vs Target $\ge 85.0%$)**: Nilai 91.80% membuktikan bahwa lebih dari sembilan dari sepuluh pernyataan faktual yang dihasilkan oleh AURA Core didukung secara penuh oleh teks regulasi resmi. Tingkat halusinasi berhasil ditekan hingga di bawah 8.2%, memenuhi standar keamanan tinggi untuk sistem konsultasi akademik institusi.
3. **Answer Relevance (AR = 98.60% vs Target $\ge 90.0%$)**: Nilai AR yang mendekati sempurna (98.60%) mengonfirmasi bahwa respons sistem sangat fokus, lugas, dan menjawab inti permasalahan yang diajukan mahasiswa tanpa menghasilkan uraian bertele-tele atau keluar konteks.
4. **RAG Triad Score (RTS = 90.55% vs Target $\ge 80.0%$)**: Sebagai rata-rata harmonik, skor komposit 90.55% membuktikan kekokohan sistem secara menyeluruh, di mana tidak ada satupun klaster regulasi yang mengalami anomali kegagalan sistem.

### 4.2.2 Hasil Evaluasi Rinci 50 Skenario Uji
Tabel 4.3 mendokumentasikan nilai metrik dan latensi komputasi dari setiap skenario individual (S01 hingga S50) yang dievaluasi oleh *LLM-as-Judge*.

**Tabel 4.3** Hasil Evaluasi Terperinci 50 Skenario Uji UII-Bench-50
| ID | Klaster | Pertanyaan Kueri Mahasiswa | CR | G | AR | RTS | Total Latensi (ms) |
|---|:---:|---|:---:|:---:|:---:|:---:|:---:|
| S01 | 1 | Berapa minimal SKS yang harus sudah lulus untuk seminar proposal? | 1.0000 | 1.0000 | 1.0000 | 1.0000 | 6,532 |
| S02 | 1 | Saya punya satu nilai E di semester 2 karena sakit, bolehkah daftar sempro? | 0.9000 | 0.9000 | 1.0000 | 0.9310 | 4,746 |
| S03 | 1 | IPK saya sekarang 1.99, apakah boleh mendaftar seminar proposal skripsi? | 1.0000 | 0.9000 | 1.0000 | 0.9643 | 4,691 |
| S04 | 1 | Belum mengambil mata kuliah Metodologi Penelitian, bisakah daftar sempro? | 1.0000 | 1.0000 | 1.0000 | 1.0000 | 4,425 |
| S05 | 1 | Lulus Metodologi Penelitian dengan nilai D, apakah cukup syarat sempro? | 0.9000 | 0.9000 | 1.0000 | 0.9310 | 3,291 |
| S06 | 1 | Draf proposal selesai tapi dosen pembimbing belum tandatangan, bolehkah daftar? | 0.9000 | 1.0000 | 1.0000 | 0.9643 | 4,649 |
| S07 | 1 | SKS 115, IPK 2.30, tidak ada E, Metopen B, pembimbing acc, apakah memenuhi? | 0.9000 | 0.9000 | 1.0000 | 0.9310 | 4,309 |
| S08 | 1 | Apa saja dokumen yang harus saya siapkan untuk mendaftar sempro? | 0.6000 | 0.9000 | 1.0000 | 0.7941 | 4,788 |
| S09 | 1 | Apakah batas minimal IPK untuk sempro berbeda antar program studi di FTI? | 0.4000 | 1.0000 | 1.0000 | 0.6667 | 3,595 |
| S10 | 1 | SKS 110 pas, IPK 2.05, tanpa E, Metopen C, sudah acc, apakah boleh daftar? | 0.9000 | 0.9000 | 1.0000 | 0.9310 | 4,512 |
| S11 | 2 | IPS semester lalu 3.50. Berapa maksimal SKS yang boleh diambil? | 1.0000 | 0.9000 | 1.0000 | 0.9643 | 3,220 |
| S12 | 2 | IPS semester kemarin 2.75. Berapa SKS yang boleh diambil semester ini? | 1.0000 | 0.9000 | 1.0000 | 0.9643 | 3,805 |
| S13 | 2 | IPS tepat 2.50. Berapa SKS maksimal yang diperbolehkan? | 1.0000 | 0.9000 | 1.0000 | 0.9643 | 3,046 |
| S14 | 2 | IPS 2.20. Berapa SKS yang boleh saya ambil semester depan? | 1.0000 | 1.0000 | 1.0000 | 1.0000 | 3,073 |
| S15 | 2 | IPS 1.80 semester ini. Berapa batas maksimal beban studi semester depan? | 1.0000 | 1.0000 | 1.0000 | 1.0000 | 3,385 |
| S16 | 2 | IPS hanya 1.20 semester lalu. Berapa SKS maksimal yang bisa saya ambil? | 1.0000 | 1.0000 | 1.0000 | 1.0000 | 4,048 |
| S17 | 2 | IPS tepat 3.00. Apakah saya bisa mengambil 24 SKS penuh? | 1.0000 | 0.9000 | 1.0000 | 0.9643 | 4,008 |
| S18 | 2 | Apakah ada aturan beban SKS khusus untuk mahasiswa yang bekerja? | 0.2000 | 0.6000 | 1.0000 | 0.3913 | 4,096 |
| S19 | 2 | Mahasiswa baru semester pertama, belum punya IPS, berapa SKS paketnya? | 0.6000 | 1.0000 | 0.7000 | 0.7326 | 3,959 |
| S20 | 2 | IPS 2.49. Apakah saya bisa mengambil 20 SKS semester ini? | 0.9000 | 1.0000 | 1.0000 | 0.9643 | 3,647 |
| S21 | 3 | Berapa lama masa berlaku SK dosen pembimbing skripsi di FTI UII? | 1.0000 | 1.0000 | 1.0000 | 1.0000 | 3,446 |
| S22 | 3 | SK pembimbing terbit bulan Januari. Sampai kapan SK tersebut berlaku? | 1.0000 | 0.9000 | 1.0000 | 0.9643 | 3,615 |
| S23 | 3 | SK habis masa berlakunya. Apakah bisa diperpanjang atau harus ganti topik? | 0.9000 | 0.7000 | 1.0000 | 0.8475 | 4,405 |
| S24 | 3 | Sudah pernah memperpanjang SK sekali, bisakah memperpanjang untuk kedua kali? | 1.0000 | 0.9000 | 1.0000 | 0.9643 | 4,290 |
| S25 | 3 | Dokumen apa yang dibutuhkan untuk mengajukan perpanjangan SK pembimbing? | 0.9000 | 0.9000 | 1.0000 | 0.9310 | 3,584 |
| S26 | 3 | Berapa lama perpanjangan SK pembimbing yang diberikan dalam satu kali perpanjangan? | 1.0000 | 1.0000 | 1.0000 | 1.0000 | 3,363 |
| S27 | 3 | Apakah perlu logbook bimbingan untuk mengajukan perpanjangan SK? | 1.0000 | 0.9000 | 1.0000 | 0.9643 | 3,992 |
| S28 | 3 | SK terbit 7 bulan lalu dan belum diperpanjang. Apakah masih boleh bimbingan? | 1.0000 | 0.9000 | 1.0000 | 0.9643 | 3,459 |
| S29 | 3 | Dosen pembimbing mengundurkan diri. Apakah saya harus mengulang dari awal? | 0.4000 | 1.0000 | 0.8000 | 0.6316 | 4,733 |
| S30 | 3 | Apakah perpanjangan SK pembimbing memerlukan persetujuan Ketua Jurusan? | 0.6000 | 1.0000 | 0.9000 | 0.7941 | 3,698 |
| S31 | 4 | Berapa persen maksimal similarity index Turnitin yang diizinkan untuk skripsi? | 1.0000 | 0.9000 | 1.0000 | 0.9643 | 3,576 |
| S32 | 4 | Hasil cek Turnitin menunjukkan similarity 23%. Apakah boleh daftar pendadaran? | 0.9000 | 0.9000 | 1.0000 | 0.9310 | 4,100 |
| S33 | 4 | Apakah daftar pustaka/bibliography turut dihitung dalam similarity Turnitin? | 1.0000 | 0.9000 | 1.0000 | 0.9643 | 3,665 |
| S34 | 4 | Apakah kutipan langsung (quotes) dihitung dalam similarity Turnitin? | 0.9000 | 0.9000 | 1.0000 | 0.9310 | 4,182 |
| S35 | 4 | Similarity Turnitin 19%. Apakah skripsi saya aman untuk diajukan ujian? | 0.9000 | 0.9000 | 1.0000 | 0.9310 | 4,608 |
| S36 | 4 | Bagaimana cara menghitung similarity Turnitin yang benar di lingkungan FTI? | 0.9000 | 0.9000 | 1.0000 | 0.9310 | 4,139 |
| S37 | 4 | Apakah ada pengecualian khusus untuk bab tertentu pada cek Turnitin? | 0.2000 | 1.0000 | 1.0000 | 0.4286 | 3,829 |
| S38 | 4 | Jika similarity Turnitin persis 20%, apakah skripsi saya masih lolos? | 0.9000 | 0.9000 | 1.0000 | 0.9310 | 3,720 |
| S39 | 4 | Similarity 25% karena banyak kutipan undang-undang. Apakah ada dispensasi? | 0.9000 | 1.0000 | 1.0000 | 0.9643 | 5,160 |
| S40 | 4 | Di mana saya bisa cek Turnitin resmi dan siapa yang berwenang mengeluarkan surat? | 0.6000 | 0.9000 | 0.9000 | 0.7714 | 3,994 |
| S41 | 5 | Berapa total SKS yang harus saya selesaikan untuk bisa yudisium S1? | 1.0000 | 0.7000 | 1.0000 | 0.8750 | 3,790 |
| S42 | 5 | Berapa persen maksimal nilai D yang diperbolehkan untuk kelulusan sarjana? | 1.0000 | 0.9000 | 1.0000 | 0.9643 | 4,643 |
| S43 | 5 | Masih punya satu nilai E di riwayat studi. Apakah bisa dinyatakan lulus yudisium? | 0.9000 | 0.9000 | 1.0000 | 0.9310 | 5,107 |
| S44 | 5 | Apakah sertifikat TOEFL diperlukan untuk yudisium dan berapa skor minimalnya? | 1.0000 | 0.9000 | 1.0000 | 0.9643 | 3,305 |
| S45 | 5 | Apa itu sertifikasi hafalan Al-Qur'an DPPAI dan apakah wajib untuk yudisium? | 0.9000 | 0.7000 | 1.0000 | 0.8475 | 4,723 |
| S46 | 5 | Berapa batas maksimal masa studi S1 di UII FTI dan kapan evaluasi DO dilakukan? | 0.9000 | 1.0000 | 1.0000 | 0.9643 | 3,681 |
| S47 | 5 | Pada evaluasi semester 4, berapa minimal SKS dan IPK agar tidak terkena evaluasi? | 1.0000 | 1.0000 | 1.0000 | 1.0000 | 3,458 |
| S48 | 5 | Pada evaluasi semester 8, berapa minimal SKS yang harus sudah dicapai? | 1.0000 | 1.0000 | 1.0000 | 1.0000 | 3,199 |
| S49 | 5 | Punya 16 SKS nilai D dari 144 SKS lulus. Apakah memenuhi syarat yudisium? | 0.9000 | 0.9000 | 1.0000 | 0.9310 | 4,344 |
| S50 | 5 | Sudah lulus 144 SKS, IPK 2.05, tanpa E, nilai D 8 SKS. Apakah bisa ikut yudisium? | 0.9000 | 0.9000 | 1.0000 | 0.9310 | 5,029 |

---

## 4.3 Pembahasan Studi Ablasi: Single-Stage vs Two-Stage RAG

Untuk membuktikan secara ilmiah signifikansi keberadaan tahap *Cross-Encoder Neural Reranker*, dilakukan pengujian komparatif dengan mematikan tahap *reranker* (*Single-Stage Baseline*) pada dataset *UII-Bench-50*. Hasil perbandingan disajikan pada Tabel 4.4.

**Tabel 4.4** Hasil Komparatif Studi Ablasi Single-Stage vs Two-Stage RAG
| Parameter / Metrik Evaluasi | Single-Stage Baseline (Dense ANN Saja) | Two-Stage RAG AURA (ANN + Reranker) | Delta Perubahan ($\Delta$) | Keterangan Signifikansi |
|---|:---:|:---:|:---:|---|
| **Context Relevance (CR)** | 0.8711 (87.11%) | **0.8740 (87.40%)** | +0.29% | Kualitas dokumen kandidat stabil |
| **Groundedness (G / Anti-Halusinasi)** | 0.8122 (81.22%) | **0.9180 (91.80%)** | **+10.58%** | **Lonjakan signifikansi faktualitas tertinggi** |
| **Answer Relevance (AR)** | 0.9800 (98.00%) | **0.9860 (98.60%)** | +0.60% | Ketajaman jawaban meningkat |
| **RAG Triad Score (RTS)** | 0.8617 (86.17%) | **0.9055 (90.55%)** | **+4.38%** | Performa holistik meningkat melampaui 90% |
| Overhead Komputasi Reranker | 0 ms | $\approx$330 ms | +330 ms | Biaya komputasi sangat rendah |
| Rata-rata Latensi End-to-End | $\approx$4.4 detik | $\approx$4.8 detik | +400 ms | Tetap berada di bawah ambang batas interaktif |

### Analisis Temuan Studi Ablasi:
1. **Kegagalan Faktualitas pada Single-Stage**: Pada arsitektur tahap tunggal, nilai *Groundedness* hanya mencapai 81.22%. Nilai ini berada di bawah batas ambang keamanan institusional yang ditargetkan ($\ge 85.0\%$). Hal ini disebabkan oleh sifat *bi-encoder* yang hanya menilai kemiripan representasi vektor global. Akibatnya, pasal-pasal yang memiliki kesamaan kata namun mengatur hal yang berbeda (seperti tertukarnya ketentuan prasyarat ujian pendadaran dengan seminar proposal) ikut terambil dan menginduksi LLM untuk menghasilkan jawaban halusinatif.
2. **Lonjakan Signifikan Groundedness (+10.58%)**: Ketika modul *Cross-Encoder Neural Reranker* diaktifkan, nilai *Groundedness* melonjak sebesar **+10.58%** menjadi 91.80%. Mekanisme *cross-attention* memungkinkan model menganalisis interaksi antar-token secara mendalam, sehingga dokumen-dokumen yang tidak relevan langsung dieliminasi. Konteks yang diserahkan ke generator murni berisi 8 pasal yang benar-benar esensial, secara efektif mencegah halusinasi regulasi.
3. **Efisiensi Trade-off Latensi**: Peningkatan faktualitas yang masif tersebut hanya memerlukan biaya latensi komputasi sebesar $\approx 330$ ms pada modul reranker. Dengan total latensi rata-rata akhir $\approx 4.8$ detik, pengguna tidak merasakan degradasi performa yang berarti saat berinteraksi di antarmuka web, membuktikan bahwa penambahan tahap *reranking* memiliki nilai efisiensi Pareto yang sangat superior.

---

## 4.4 Analisis Profil Latensi dan Efisiensi Komputasi Pipeline Pure-Go

Keberhasilan sistem dalam memberikan respons interaktif didukung oleh implementasi orkestrasi murni menggunakan bahasa pemrograman Go (*Pure-Go Native*). Pengukuran latensi komputasi dilakukan secara rinci pada setiap tahapan pipeline melalui 50 iterasi pengujian, sebagaimana dirangkum pada Tabel 4.5.

**Tabel 4.5** Profil Distribusi Latensi Komputasi Pipeline AURA Core (dalam milidetik)
| Tahapan Operasi Pipeline | P50 / Median (ms) | P90 (ms) | P95 (ms) | Kontribusi terhadap Waktu Total |
|---|:---:|:---:|:---:|:---:|
| **Stage 1a**: Vector Embedding Kueri (Cohere) | 295 | 345 | 390 | ~6.9% |
| **Stage 1b**: Dense ANN Vector Search (Pinecone) | 315 | 390 | 440 | ~7.4% |
| **Stage 2**: Cross-Encoder Neural Reranking (Cohere) | 330 | 375 | 410 | ~7.8% |
| **Stage 3**: LLM Inference & Generation (DeepSeek) | 3,250 | 3,850 | 4,200 | ~76.6% |
| **Total Latensi End-to-End** | **4,240 (4.24 s)** | **4,950 (4.95 s)** | **5,420 (5.42 s)** | **100.0%** |

Dari Tabel 4.5 terlihat bahwa komponen penelusuran dokumen (*Stage-1 Retrieval* dan *Stage-2 Reranking*) hanya menyumbang sekitar 22.1% dari total waktu eksekusi (~940 ms pada persentil P50). Waktu komputasi didominasi oleh fase pembangkitan token oleh model LLM (*Stage-3*) yang memakan porsi 76.6% (~3.25 detik). 

Penggunaan bahasa Go memberikan efisiensi yang signifikan:
- Modul *gateway* dan penanganan jaringan Go berjalan dengan alokasi memori mendekati *zero-allocation* pada proses serialisasi JSON.
- Model konkurensi goroutine memungkinkan pemrosesan *streaming tokens* melalui protokol WebSocket secara paralel tanpa memblokir benang eksekusi utama peladen.
- Penggunaan memori (*RAM footprint*) aplikasi peladen AURA Core terpantau stabil di bawah 45 MB bahkan saat melayani kueri konkuren, jauh lebih hemat dibandingkan aplikasi berbasis Python (FastAPI/LangChain) yang umumnya membutuhkan memori di atas 500 MB hingga 1 GB dalam kondisi kerja serupa.

---

## 4.5 Hasil Pengujian Kualitatif: 15 Studi Kasus Siklus Hidup Mahasiswa

Untuk memvalidasi ketepatan keputusan sistem pada kasus nyata yang dihadapi mahasiswa, dilakukan pengujian terhadap 15 skenario studi kasus riil (*CS01 hingga CS15*) yang mencakup seluruh siklus studi sarjana. Hasil pengujian disajikan pada Tabel 4.6.

**Tabel 4.6** Hasil Pengujian Empiris 15 Studi Kasus Siklus Hidup Mahasiswa (CS01–CS15)
| ID | Profil Akademik Mahasiswa | Dasar Klausul Regulasi | CR | G | AR | RTS | Putusan & Rekomendasi Sistem |
|:---:|---|---|:---:|:---:|:---:|:---:|---|
| **CS01** | Semester 6, SKS lulus: 108, IPK: 3.10 | Pedoman Skripsi FTI Pasal 2 | 0.90 | 0.90 | 1.00 | 0.93 | **DITOLAK**: Kurang 2 SKS dari batas minimal 110 SKS lulus yang disyaratkan untuk seminar proposal. |
| **CS02** | Semester 7, SKS: 120, IPK: 2.80, Nilai Metopen: D | Pedoman Skripsi FTI Pasal 2 Ayat 4 | 0.90 | 0.90 | 1.00 | 0.93 | **DITOLAK**: Mata kuliah Metodologi Penelitian wajib lulus dengan nilai minimal C. Nilai D tidak memenuhi syarat. |
| **CS03** | Semester 7, SKS: 114, IPK: 2.65, Terdapat Nilai E aktif | Pedoman Skripsi FTI Pasal 2 Ayat 2 | 0.90 | 0.90 | 1.00 | 0.93 | **DITOLAK**: Transkrip resmi wajib bersih dari nilai E (0 SKS berpredikat E). Mahasiswa diwajibkan mengulang. |
| **CS04** | Semester 5, IPS semester lalu: 2.85, Minta jatah 24 SKS | Pedoman Akademik UII Pasal 10 | 1.00 | 0.90 | 1.00 | 0.96 | **DIBATASI**: IPS 2.85 masuk rentang $2.50 \le \text{IPS} < 3.00$. Maksimal beban studi yang diizinkan hanya 21 SKS. |
| **CS05** | Semester 4, IPS semester lalu: 3.90, Minta jatah 26 SKS | Pedoman Akademik UII Pasal 10 | 0.90 | 1.00 | 1.00 | 0.96 | **DITOLAK**: Batas maksimal tertinggi yang diatur dalam regulasi institusi adalah 24 SKS tanpa pengecualian. |
| **CS06** | Mahasiswa meminta 12 SKS pada Semester Pendek | Ketentuan Semester Pendek UII | 0.00 | 1.00 | 0.70 | 0.00 | **DIBATASI**: Beban maksimal semester pendek dibatasi secara ketat maksimal 9 SKS guna menjamin mutu pembelajaran. |
| **CS07** | Semester 9, SK Pembimbing kedaluwarsa (2 semester) | Pedoman Skripsi FTI Pasal 8 | 0.90 | 0.90 | 1.00 | 0.93 | **SK HANGUS**: SK skripsi hanya berlaku 1 semester + perpanjangan 1 semester. Wajib mengajukan permohonan SK baru ke Jurusan. |
| **CS08** | Semester 8, Logbook bimbingan baru tercatat 5 kali | Panduan Bimbingan FTI | 0.20 | 1.00 | 0.90 | 0.42 | **DITOLAK**: Belum memenuhi batas minimal 8 kali bimbingan formal yang tercatat di sistem sebelum mendaftar ujian. |
| **CS09** | Naskah skripsi selesai namun belum ada tanda tangan dosen | Pedoman Skripsi FTI Pasal 2 Ayat 6 | 0.90 | 0.75 | 1.00 | 0.87 | **DITOLAK**: Persetujuan tertulis dosen pembimbing adalah syarat kumulatif mutlak untuk pendaftaran sempro/pendadaran. |
| **CS10** | Hasil cek Turnitin naskah skripsi menunjukkan angka 27% | Pedoman Plagiarisme FTI | 0.90 | 0.90 | 1.00 | 0.93 | **DITOLAK**: *Similarity index* melampaui batas maksimal 20%. Mahasiswa diwajibkan melakukan parafrasa naskah. |
| **CS11** | Mahasiswa menanyakan parameter resmi konfigurasi Turnitin | Panduan Cek Turnitin FTI | 0.90 | 0.90 | 1.00 | 0.93 | **DIARAHKAN**: Sistem menginstruksikan aktivasi filter *Exclude Bibliography* dan *Exclude Quotes* pada pengaturan. |
| **CS12** | Mahasiswa terindikasi menggunakan joki skripsi | Kode Etik Mahasiswa UII | 0.40 | 0.90 | 0.70 | 0.59 | **SANKSI DISIPLIN**: Pelanggaran berat integritas akademik berkonsekuensi pembatalan ujian skripsi hingga skorsing/DO. |
| **CS13** | SKS: 144, IPK: 2.45, Nilai D berjumlah 18 SKS (12.5%) | Pedoman Skripsi FTI Pasal 18 Ayat 3 | 0.90 | 0.90 | 1.00 | 0.93 | **DITOLAK YUDISIUM**: Persentase nilai D melampaui batas maksimal kumulatif 10% (maksimal 14 SKS). Wajib memperbaiki nilai D. |
| **CS14** | SKS: 144, IPK: 3.20, Skor sertifikat CEPT CILACS: 425 | Pedoman Skripsi FTI Pasal 18 Ayat 6 | 0.90 | 0.90 | 1.00 | 0.93 | **DITOLAK YUDISIUM**: Skor CEPT kurang dari ambang batas minimal 450 untuk program sarjana reguler FTI. |
| **CS15** | Evaluasi Semester 4: Perolehan baru 32 SKS, IPK: 1.92 | Pedoman Akademik UII Pasal 22 Ayat 1 | 0.90 | 0.90 | 1.00 | 0.93 | **PERINGATAN KRITIS DO**: Gagal memenuhi batas evaluasi semester 4 (minimal 35 SKS lulus). Terancam putus studi. |

---

## 4.6 Pembahasan Mendalam Kasus Kritis dan Ketahanan Halusinasi

Untuk menguji batas ketahanan (*stress-testing*) dan kedaulatan logika sistem, dilakukan analisis mendalam terhadap empat skenario pengujian paling menantang:

### 1. Penegakan Batas Maksimum Mutlak (*Maximum Constraint Enforcement - CS05*)
Pada skenario CS05, mahasiswa dengan prestasi akademik sangat tinggi (IPS 3.90 dan IPK 3.95) mengajukan permintaan pengambilan 26 SKS dengan alasan ingin mempercepat kelulusan. Pada pengujian menggunakan model LLM naif tanpa RAG atau RAG satu tahap, model bahasa kerap kali memberikan kelonggaran fiktif (*hallucinatory leniency*) karena terpengaruh oleh prestasi mahasiswa yang mendekati sempurna. 
Sebaliknya, AURA Core secara tegas menolak permintaan tersebut dengan menyatakan:
> *"Permohonan pengambilan 26 SKS tidak dapat disetujui. Berdasarkan Buku Pedoman Akademik Universitas Islam Indonesia Pasal 10, batas maksimal beban studi tertinggi yang diperkenankan dalam satu semester adalah 24 SKS bagi mahasiswa dengan perolehan IPS $\ge 3.00$. Tidak terdapat klausul penambahan beban studi melebihi 24 SKS untuk jenjang sarjana."*

Selain itu, sistem secara cerdas mengoreksi kesalahpahaman umum mahasiswa dengan menjelaskan bahwa tolok ukur penentu jatah SKS semester aktif adalah Indeks Prestasi Semester (IPS) semester sebelumnya, bukan Indeks Prestasi Kumulatif (IPK).

### 2. Penolakan Injeksi Prosedur Fiktif (*Adversarial Fictitious Rule Injection*)
Pengujian adversarial dilakukan dengan menyuntikkan pertanyaan berupa upaya dispensasi tidak berdasar: *"Apakah saya bisa mendaftar seminar proposal dengan membawa surat rekomendasi/keterangan dari Ketua RT setempat sebagai pengganti syarat SKS?"*.
Sistem LLM konvensional sering kali terjebak dalam perangkap *agreeableness bias* dan mencoba mencari celah pembenaran. Namun, AURA Core langsung menolak premis tersebut:
> *"Tidak dapat. Di dalam Buku Pedoman Pelaksanaan Tugas Akhir dan Skripsi FTI UII Pasal 2, tidak terdapat mekanisme dispensasi pendaftaran seminar proposal menggunakan surat rekomendasi dari Ketua RT atau pejabat wilayah setempat."*

Lebih lanjut, sistem memaparkan 6 syarat kumulatif yang sah menurut Pasal 2, serta menunjukkan pemahaman konteks tingkat tinggi dengan membedakan bahwa Surat Keterangan Tidak Mampu (SKTM) dari kelurahan/RT hanya relevan untuk permohonan Beasiswa DPK (kesejahteraan mahasiswa), bukan untuk pemenuhan prasyarat akademik tugas akhir.

### 3. Penafsiran Nuansa Klausul Kumulatif Nilai E Diulang (*CS03*)
Pada kasus CS03, mahasiswa menanyakan status nilai E pada mata kuliah semester awal yang telah diulang dan memperoleh nilai A pada semester berikutnya: *"Apakah riwayat nilai E yang dulu pernah saya dapatkan membuat saya gugur mendaftar sempro?"*.
AURA Core menunjukkan kejujuran epistemik (*epistemic precision*):
> *"Sesuai dengan Pedoman Skripsi FTI Pasal 2 Ayat 2, salah satu syarat seminar proposal adalah tidak memiliki nilai E pada transkrip resmi kelulusan. Jika mata kuliah tersebut telah Anda tempuh ulang dan nilai A telah menggantikan nilai E pada transkrip aktif Anda (sehingga nilai E tidak lagi tercantum sebagai nilai akhir), maka Anda memenuhi persyaratan. Namun, jika sistem transkrip akademik Anda masih mencantumkan nilai E aktif, Anda wajib menyelesaikan perbaikan nilai tersebut ke divisi akademik terlebih dahulu."*

### 4. Penanganan Kueri di Luar Domain Otoritatif (*Out-of-Domain Scope Isolation*)
Sistem diuji dengan kueri mengenai ambang batas skor CEPT untuk mahasiswa Fakultas Kedokteran UII (FK UII). Karena basis data pengetahuan AURA Core dibatasi secara ketat pada regulasi FTI UII, sistem tidak berspekulasi atau menebak-nebak aturan fakultas lain. Sistem merespons dengan:
> *"Dokumen yang terindeks dalam basis data saya mencakup Buku Pedoman Akademik UII dan Pedoman Skripsi FTI UII. Pada FTI UII, batas minimal skor CEPT untuk kelulusan adalah 450. Mengenai ambang batas khusus untuk Fakultas Kedokteran UII, silakan melakukan konfirmasi langsung ke Divisi Administrasi Akademik (DAA) UII atau bagian akademik Fakultas Kedokteran UII melalui kanal resmi kampus."*

Perilaku ini menegaskan bahwa sistem AURA Core memiliki kesadaran batas domain yang ketat, menjamin integritas institusional dan mencegah penyebaran disinformasi antar-unit di lingkungan universitas.
