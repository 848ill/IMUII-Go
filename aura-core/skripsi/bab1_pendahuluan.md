# BAB I
# PENDAHULUAN

## 1.1 Latar Belakang Masalah

Pendidikan tinggi menuntut tata kelola akademik yang tertib, transparan, dan akuntabel. Di tingkat fakultas dan universitas, regulasi akademik berfungsi sebagai instrumen normatif utama yang mengatur hak, kewajiban, dan tahapan studi mahasiswa. Regulasi ini mencakup berbagai spektrum prosedural yang sangat penting, meliputi batas pengambilan beban Satuan Kredit Semester (SKS) berdasarkan Indeks Prestasi Semester (IPS), prasyarat kelayakan pendaftaran seminar proposal tugas akhir, batas masa berlaku Surat Keputusan (SK) dosen pembimbing, ambang batas toleransi orisinalitas karya ilmiah (*Turnitin similarity index*), hingga evaluasi keberlanjutan studi pada semester 4 dan semester 8 guna mencegah putus studi (*drop out*), serta persyaratan yudisium kelulusan sarjana (Universitas Islam Indonesia, 2024; Fakultas Teknologi Industri UII, 2024).

Meskipun buku pedoman akademik telah dipublikasikan secara resmi dalam bentuk dokumen cetak maupun berkas PDF digital, pemahaman mahasiswa terhadap klausul-klausul regulasi tersebut kerap kali parsial dan terfragmentasi. Mahasiswa sering kali mengalami kesulitan dalam menafsirkan pasal-pasal yang bersifat kumulatif—yakni ketentuan yang mewajibkan seluruh prasyarat terpenuhi tanpa pengecualian—seperti larangan keberadaan nilai E pada riwayat transkrip untuk pendaftaran seminar proposal, atau perbedaan mendasar antara IPS semester sebelumnya dengan IPK kumulatif dalam menentukan jatah SKS semester aktif. Akibatnya, terjadi beban konsultasi yang berulang (*repetitive administrative overhead*) pada dosen pembimbing akademik dan staf divisi administrasi akademik. Lebih jauh lagi, salah tafsir terhadap aturan yang mengikat berpotensi menimbulkan konsekuensi fatal bagi mahasiswa, seperti pembatalan rencana studi secara sepihak, kedaluwarsanya masa bimbingan skripsi, hingga keterlambatan kelulusan yang merugikan secara moril dan materil.

Perkembangan pesat teknologi kecerdasan buatan, khususnya *Large Language Models* (LLM) seperti GPT-4 dan DeepSeek, telah membuka peluang signifikan bagi pengembangan asisten percakapan cerdas (*conversational AI*) yang mampu berdialog dalam bahasa alami. Namun demikian, penggunaan LLM secara langsung (*vanilla inference*) pada domain institusional akademik yang sarat hukum memiliki kerentanan kritis, yaitu fenomena halusinasi (*hallucination*) (Ji dkk., 2023; Huang dkk., 2023). LLM pada dasarnya bekerja berdasarkan probabilitas kemunculan token berikutnya (*stochastic next-token prediction*), bukan penalaran faktual deterministik. Ketika ditanya mengenai aturan spesifik kampus yang tidak tercakup secara mendalam dalam data pra-pelatihannya, LLM cenderung mengarang aturan fiktif yang terdengar sangat meyakinkan (*fluent yet factually ungrounded*), seperti mengizinkan beban studi melampaui batas legal atau mengarang prosedur dispensasi yang tidak pernah ada dalam peraturan rektor.

Untuk mengatasi kelemahan memori parametrik LLM, paradigma *Retrieval-Augmented Generation* (RAG) diperkenalkan sebagai standar baru dalam penjaminan faktualitas teks (Lewis dkk., 2020; Karpukhin dkk., 2020). Melalui pendekatan RAG, sistem terlebih dahulu menelusuri potongan dokumen resmi dari basis data pengetahuan eksternal (*non-parametric memory*) yang relevan dengan pertanyaan pengguna, lalu menyuntikkannya ke dalam konteks perintah (*prompt context*) sebelum LLM menyusun jawaban. Kendati demikian, implementasi RAG konvensional tahap tunggal (*Single-Stage RAG*) yang umum diterapkan masih memiliki empat kelemahan mendasar:
1. **Pergeseran Semantik pada Pencarian Vektor (*Semantic Drift in Dense ANN*)**: Pencarian kemiripan berbasis *Approximate Nearest Neighbor* (ANN) menggunakan *bi-encoder* bekerja dengan memproyeksikan teks ke dalam ruang vektor berdimensi tinggi. Pada teks regulasi hukum yang memiliki leksikal sangat mirip antar-pasal (seperti pengulangan kata "mahasiswa", "semester", "batas", dan "persyaratan"), pencarian vektor tunggal sering kali menarik pasal yang secara leksikal serupa namun secara konteks hukum tidak dapat diterapkan (*inapplicable clause*).
2. **Efek *Lost-in-the-Middle* Akibat Redundansi Konteks**: Apabila sejumlah besar potongan pasal (misalnya 20 kandidat dokumen) disodorkan langsung ke LLM tanpa penyaringan presisi, model sering kali mengalami disorientasi perhatian (*attention dilution*) dan melewatkan klausul inti yang berada di posisi tengah jendela konteks (Liu dkk., 2023).
3. **Risiko Privasi Data Akademik Mahasiswa**: Konsultasi akademik yang bersifat personal mengharuskan sistem untuk menganalisis dokumen privat seperti Kartu Hasil Studi (KHS) atau Kartu Rencana Studi (KRS). Pemuatan data privat mahasiswa ke dalam basis data vektor publik bersama (*communal vector database*) menimbulkan ancaman kebocoran privasi lintas mahasiswa yang melanggar prinsip kepatuhan data pribadi.
4. **Overhead Runtime dan Ketergantungan Kerangka Kerja Python**: Mayoritas purwarupa RAG akademik saat ini dibangun menggunakan kerangka kerja Python tingkat tinggi seperti LangChain atau LlamaIndex. Kerangka kerja ini sering kali memiliki jejak memori (*memory footprint*) yang besar, ketergantungan paket yang rentan rusak (*brittle dependencies*), serta latensi eksekusi yang tinggi ketika melayani permintaan konkuren di lingkungan komputasi peladen institusi yang terbatas.

Guna menjawab tantangan tersebut, penelitian ini merancang dan mengimplementasikan **AURA Core** (*Academic Universal Regulatory Assistant*), sebuah sistem asisten regulasi akademik berbasis arsitektur *Pure-Go Native Two-Stage Retrieval-Augmented Generation* yang dirancang khusus untuk lingkungan Fakultas Teknologi Industri, Universitas Islam Indonesia (FTI UII). Sistem ini mengintegrasikan tahap penelusuran awal (*Stage-1 Dense ANN Vector Retrieval*) menggunakan representasi vektor multibahasa 1024-dimensi untuk menjaring kandidat pasal secara luas (Top-20), yang kemudian dipadatkan secara presisi melalui tahap kedua (*Stage-2 Cross-Encoder Neural Reranking*) menjadi Top-8 klausul paling otoritatif dengan memanfaatkan komputasi *full cross-attention*. Selain itu, pemrosesan dokumen akademik privat mahasiswa (KHS/KRS) diisolasi secara ketat di dalam memori (*in-memory parsing*) per sesi transaksi tanpa pernah dipersistensikan ke basis data vektor publik.

Kinerja faktualitas dan ketahanan sistem terhadap halusinasi diuji secara komprehensif menggunakan tolok ukur terstandarisasi **UII-Bench-50** yang mencakup 50 skenario regulasi di 5 klaster institusional, dievaluasi melalui kerangka kerja otomatis *LLM-as-Judge RAG Triad* (*Context Relevance*, *Groundedness*, *Answer Relevance*), serta divalidasi melalui studi ablasi empiris dan pengujian 15 studi kasus siklus hidup mahasiswa.

---

## 1.2 Rumusan Masalah

Berdasarkan latar belakang yang telah dipaparkan, rumusan masalah dalam penelitian ini adalah sebagai berikut:
1. Bagaimana merancang dan membangun arsitektur *Two-Stage RAG* secara *native* menggunakan bahasa pemrograman Go (*Pure-Go*) tanpa ketergantungan kerangka kerja eksternal guna menghasilkan orkestrasi penelusuran regulasi yang efisien dan konkuren?
2. Bagaimana mekanisme isolasi data akademik mahasiswa (KHS/KRS) dirancang agar dapat dianalisis secara presisi dalam proses penalaran RAG tanpa melanggar prinsip privasi dan tanpa persistensi pada basis data vektor publik?
3. Seberapa besar signifikansi peningkatan akurasi faktualitas (*Groundedness*) yang diperoleh dari penambahan tahap *Cross-Encoder Neural Reranking* dibandingkan arsitektur *Single-Stage RAG* melalui studi ablasi empiris?
4. Bagaimana performa sistem AURA Core dalam menyelesaikan 50 skenario uji regulasi akademik institusional (*UII-Bench-50*) dan 15 studi kasus riil mahasiswa ditinjau dari metrik *RAG Triad* (*Context Relevance*, *Groundedness*, *Answer Relevance*) serta latensi komputasi *end-to-end*?

---

## 1.3 Batasan Masalah

Ruang lingkup dan batasan masalah dalam penelitian ini ditetapkan sebagai berikut:
1. **Domain Pengetahuan Regulasi**: Korpus dokumen otoritatif yang digunakan dibatasi pada regulasi resmi yang berlaku di lingkungan Fakultas Teknologi Industri Universitas Islam Indonesia, yang meliputi:
   - *Buku Pedoman Akademik Universitas Islam Indonesia* (Edisi 2024, 38 Pasal).
   - *Buku Pedoman Pelaksanaan Tugas Akhir dan Skripsi Fakultas Teknologi Industri UII* (Edisi 2024, 24 Pasal).
   - Ketentuan resmi terkait Beasiswa DPK dan DPPAI UII serta batas toleransi kemiripan karya ilmiah (*Turnitin similarity index*).
2. **Arsitektur Model dan Komponen Teknis**:
   - Model representasi vektor (*embedding*) menggunakan Cohere `embed-multilingual-v3.0` (1024 dimensi).
   - Basis data indeks vektor menggunakan Pinecone Serverless Vector Index dengan metrik kemiripan *Cosine Distance*.
   - Model penyortiran ulang (*reranker*) menggunakan Cohere `rerank-v3.5` berbasis arsitektur *Cross-Encoder*.
   - Model pembangkit teks (*generation engine*) menggunakan DeepSeek LLM (`deepseek-chat`) dengan parameter suhu (*temperature*) rendah $\tau = 0.2$ guna membatasi keluaran acak.
3. **Lingkungan Pengembangan dan Penerapan**:
   - Seluruh modul orkestrator, *gateway* HTTP/WebSocket, logika *parsing*, pemanggilan API hilir, dan kerangka evaluasi dibangun secara *native* menggunakan pustaka standar Go 1.24+ (*Go Standard Library*).
   - Sistem diuji dan dioperasikan pada peladen virtual *Ubuntu 24.04 LTS* berdaya tampung komputasi awan produksi institusional (`aura.imuii.id`).
4. **Metodologi Pengujian**:
   - Evaluasi kuantitatif dilakukan terhadap 50 skenario uji representatif (*UII-Bench-50*) yang terdistribusi ke dalam 5 klaster regulasi akademik institusi.
   - Penilaian metrik *Context Relevance*, *Groundedness*, dan *Answer Relevance* dilakukan secara deterministik menggunakan pendekatan *LLM-as-Judge* dengan parameter suhu $\tau = 0.0$.

---

## 1.4 Tujuan Penelitian

Tujuan yang ingin dicapai melalui penelitian ini adalah:
1. Mengembangkan arsitektur sistem *Two-Stage RAG* berkinerja tinggi yang diimplementasikan secara murni menggunakan bahasa pemrograman Go (*Pure-Go Native Orchestrator*) untuk domain regulasi akademik kampus.
2. Mengimplementasikan arsitektur pemrosesan dokumen KHS/KRS mahasiswa yang aman (*privacy-preserving in-memory context injection*) guna mencegah kebocoran data akademik privat ke repositori vektor publik.
3. Membuktikan secara empiris melalui studi ablasi (*ablation study*) kontribusi penambahan tahap *Cross-Encoder Neural Reranking* dalam mengeliminasi halusinasi dan meningkatkan nilai *Groundedness* sistem.
4. Mengukur dan mengevaluasi efektivitas sistem AURA Core berdasarkan metrik standar *RAG Triad* (*Context Relevance*, *Groundedness*, *Answer Relevance*, dan *RAG Triad Score*) pada dataset *UII-Bench-50* serta profil latensi komputasi *end-to-end*.

---

## 1.5 Manfaat Penelitian

Penelitian ini diharapkan memberikan kontribusi dan manfaat teoretis maupun praktis sebagai berikut:

### 1.5.1 Manfaat Teoretis
1. **Pengayaan Literatur AI Institusional**: Memberikan sumbangsih empiris terhadap penerapan arsitektur *Two-Stage RAG* pada teks hukum dan regulasi akademik yang memiliki tingkat ambiguitas leksikal tinggi.
2. **Validasi Kerangka Evaluasi *LLM-as-Judge***: Menjadi rujukan metodologis dalam pemanfaatan metrik *RAG Triad* untuk pengujian sistem kepatuhan (*compliance verification*) berbasis kecerdasan buatan.
3. **Pola Desain Sistem AI Konkuren**: Mendokumentasikan pola desain rekayasa perangkat lunak sistem RAG nir-kerangka-kerja (*frameworkless native implementation*) menggunakan paradigma konkurensi Go untuk menekan latensi dan jejak memori.

### 1.5.2 Manfaat Praktis
1. **Bagi Mahasiswa FTI UII**: Memberikan akses konsultasi regulasi akademik yang interaktif, responsif, dan akurat selama 24 jam sehari dengan rujukan pasal resmi, sehingga memitigasi risiko kesalahan administratif selama masa perkuliahan.
2. **Bagi Pengelola Akademik dan Dosen Pembimbing**: Mengurangi beban kerja repetitif staf administrasi akademik dan dosen dalam melayani pertanyaan administratif dasar, memungkinkan fokus kerja dialihkan pada bimbingan substansi akademik yang lebih esensial.
3. **Bagi Fakultas Teknologi Industri UII**: Mewujudkan tata kelola layanan informasi akademik berbasis teknologi terkini yang aman terhadap data privat mahasiswa serta berdaulat (*sovereign infrastructure*).

---

## 1.6 Sistematika Penulisan

Naskah skripsi ini disusun secara sistematis ke dalam lima bab dengan uraian sebagai berikut:

* **BAB I PENDAHULUAN**  
  Menguraikan latar belakang permasalahan regulasi akademik di perguruan tinggi, urgensi mitigasi halusinasi pada LLM, identifikasi rumusan masalah, batasan penelitian, tujuan yang hendak dicapai, manfaat teoretis dan praktis, serta sistematika penulisan laporan.

* **BAB II TINJAUAN PUSTAKA DAN LANDASAN TEORI**  
  Menyajikan kajian pustaka mengenai penelitian terdahulu yang relevan di bidang RAG dan sistem penjawab pertanyaan otomatis, diikuti landasan teori komprehensif mengenai *Large Language Models*, konsep *Dense Vector Retrieval*, arsitektur *Cross-Encoder Neural Reranking*, fenomena *Lost-in-the-Middle*, kerangka evaluasi *RAG Triad*, serta keunggulan bahasa Go dalam komputasi sistem cerdas.

* **BAB III METODOLOGI PENELITIAN DAN PERANCANGAN SISTEM**  
  Menjelaskan tahapan metodologi penelitian, akuisisi dan segmentasi korpus regulasi akademik FTI UII, perancangan arsitektur terperinci sistem AURA Core (*Stage-1 Retrieval, Stage-2 Reranking, In-Memory Privacy Parser, Stage-3 Grounded Reasoning*), serta perancangan dataset pengujian *UII-Bench-50* dan mekanisme otomatisasi *LLM-as-Judge*.

* **BAB IV HASIL DAN PEMBAHASAN**  
  Memaparkan implementasi sistem pada lingkungan komputasi nyata, penyajian data hasil evaluasi kuantitatif 50 skenario *UII-Bench-50* pada 5 klaster regulasi, pembahasan studi ablasi perbandingan *Single-Stage* vs *Two-Stage RAG*, analisis profil latensi eksekusi *end-to-end*, pembahasan 15 studi kasus riil mahasiswa, serta analisis mendalam terhadap kasus-kasus batas (*adversarial and edge cases*).

* **BAB V KESIMPULAN DAN SARAN**  
  Menyimpulkan hasil-hasil temuan penelitian yang menjawab seluruh rumusan masalah dan tujuan penelitian, serta menyajikan saran dan rekomendasi pengembangan teknis untuk penelitian lanjutan di masa mendatang.
