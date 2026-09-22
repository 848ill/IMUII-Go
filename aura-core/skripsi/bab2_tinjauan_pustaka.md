# BAB II
# TINJAUAN PUSTAKA DAN LANDASAN TEORI

## 2.1 Tinjauan Pustaka (*State of the Art*)

Penerapan kecerdasan buatan dalam penanganan dokumen institusional, khususnya sistem penjawab pertanyaan berbasis regulasi (*regulatory question answering*), telah mengalami perkembangan pesat seiring ditemukannya paradigma *Retrieval-Augmented Generation* (RAG). Subbab ini mengkaji penelitian-penelitian terdahulu yang menjadi fondasi dan pembanding bagi penelitian ini.

Lewis dkk. (2020) meletakkan fondasi arsitektur RAG hibrida yang menggabungkan model bahasa pra-terlatih (*parametric memory*) dengan penelusur teks padat (*dense retrieval*) berbasis *Dense Passage Retrieval* (DPR) (Karpukhin dkk., 2020). Pendekatan ini terbukti melampaui model generatif parametrik murni dalam tugas-tugas NLP intensif pengetahuan (*knowledge-intensive NLP tasks*), karena model dapat mengakses fakta terkini tanpa perlu melakukan pelatihan ulang (*fine-tuning*) pada miliaran parameter bobotnya. Namun demikian, penelitian Lewis dkk. menggunakan pendekatan pencarian satu tahap (*single-stage*) yang rentan terhadap ketidaktepatan dokumen saat dihadapkan pada korpus hukum yang memiliki kesamaan semantik tinggi antar-dokumen.

Keterbatasan penelusur padat berbasis representasi vektor tunggal (*bi-encoder*) diuraikan secara mendalam oleh Reimers dan Gurevych (2019) melalui *Sentence-BERT*. Meskipun *bi-encoder* sangat efisien dalam memetakan kalimat ke dalam ruang vektor berdimensi tetap untuk pencarian kemiripan kosinus (*cosine similarity*) secara cepat, arsitektur ini memisahkan representasi kueri dan dokumen sehingga kehilangan interaksi token tingkat kata (*cross-token interactions*). Untuk mengatasi hal ini, Nogueira dan Cho (2019) memperkenalkan *Passage Re-ranking with BERT* berbasis *cross-encoder*. Dengan menyuapkan pasangan kueri dan kandidat dokumen secara simultan ke dalam lapisan *self-attention* transformer, *cross-encoder* mampu menangkap keterkaitan semantik yang jauh lebih kaya dan presisi. Meskipun membutuhkan biaya komputasi yang tinggi, Nogueira dan Cho menunjukkan bahwa menerapkan *cross-encoder* hanya pada subset kandidat teratas hasil saringan awal menghasilkan efisiensi Pareto yang optimal.

Dalam tinjauan komprehensif, Gao dkk. (2023) mengklasifikasikan evolusi RAG menjadi tiga paradigma: *Naive RAG*, *Advanced RAG*, dan *Modular RAG*. Paradigma *Naive RAG* yang hanya mengandalkan penelusuran vektor sederhana dinilai tidak memadai untuk sistem produksi kritis karena tingginya tingkat *retrieval noise* dan risiko halusinasi. Sebaliknya, *Advanced RAG* memperkenalkan strategi pra-penelusuran (*pre-retrieval*) seperti *semantic chunking* dan pasca-penelusuran (*post-retrieval*) seperti *neural reranking*. Gao dkk. menegaskan bahwa penambahan modul *reranker* merupakan komponen pembeda paling krusial dalam mencegah distorsi konteks pada LLM generatif.

Urgensi pemadatan konteks diperkuat oleh temuan empiris Liu dkk. (2023) dalam publikasi ilmiah mereka mengenai fenomena *Lost in the Middle*. Penelitian tersebut membuktikan bahwa performa LLM dalam mengekstraksi informasi faktual mengalami penurunan drastis ketika informasi kunci berada di posisi tengah dokumen konteks yang panjang. LLM cenderung memiliki bias perhatian pada awal (*primacy effect*) dan akhir konteks (*recency effect*). Temuan ini menjadi landasan ilmiah yang kuat bahwa menyajikan 20 dokumen mentah hasil pencarian vektor ke LLM justru kontraproduktif, sehingga diperlukan penyaringan ketat menjadi sedikit dokumen yang sangat relevan (misalnya Top-8) melalui *cross-encoder*.

Dari aspek evaluasi, pengukuran keandalan sistem RAG tidak dapat lagi mengandalkan metrik leksikal tradisional seperti BLEU atau ROUGE yang tidak mampu mengukur faktualitas semantik. Es dkk. (2023) melalui kerangka kerja *RAGAS* dan Saad-Falcon dkk. (2023) melalui kerangka kerja *TruLens* memformalisasikan konsep *RAG Triad* berbasis pendekatan *LLM-as-Judge*. Tiga pilar utama dalam *RAG Triad*—yaitu *Context Relevance* (relevansi konteks terhadap pertanyaan), *Groundedness* (kesetiaan klaim jawaban terhadap konteks rujukan), dan *Answer Relevance* (ketepatan jawaban menyelesaikan masalah pengguna)—telah diakui secara luas dalam komunitas ilmiah sebagai standar de facto untuk mengukur keterbebasan sistem dari halusinasi (Zheng dkk., 2023).

Terkait aspek rekayasa perangkat lunak dan performa peladen, mayoritas penelitian sistem RAG sebelumnya mengandalkan kerangka kerja Python (seperti LangChain atau LlamaIndex). Namun, penelitian Cox dkk. (2020) dan Pike (2012) menunjukkan bahwa untuk layanan sistem mikro berkonkurensi tinggi dengan latensi rendah, bahasa Go menawarkan keunggulan komparatif yang signifikan. Go memiliki paradigma konkurensi berbasis *Goroutine* yang sangat ringan (alokasi memori awal hanya ~2 KB per goroutine dibandingkan ~1-2 MB pada thread OS tradisional), ketiadaan overhead *interpreter*, serta kompilasi biner statis mandiri (*single static binary*) yang tahan terhadap degradasi performa di lingkungan peladen produksi terbatas.

Tabel 2.1 menyajikan pemetaan posisi penelitian ini (*research positioning matrix*) dibandingkan dengan penelitian-penelitian terdahulu yang relevan.

**Tabel 2.1** Matriks Pemetaan Posisi Penelitian AURA Core terhadap Penelitian Terdahulu
| Peneliti & Tahun | Domain Aplikasi | Arsitektur Retrieval | Model Reranking | Isolasi Privasi Pengguna | Bahasa Pemrograman / Runtime | Metrik Evaluasi |
|---|---|---|---|---|---|---|
| Lewis dkk. (2020) | *Open-Domain QA* (Wikipedia) | *Single-Stage* (DPR Bi-Encoder) | Tanpa Reranker | Tidak Ada (Publik) | Python (PyTorch) | Akurasi Leksikal / Exact Match |
| Nogueira & Cho (2019) | *Information Retrieval* (MS MARCO) | *Two-Stage Cascade* | BERT Cross-Encoder | Tidak Ada (Publik) | Python (TensorFlow) | MRR@10, NDCG@10 |
| Liu dkk. (2023) | Analisis Panjang Konteks LLM | Multi-passage QA | Tanpa Reranker | Tidak Ada | Python | Accuracy vs Document Position |
| Es dkk. (2023) (RAGAS) | Evaluasi Kerangka RAG Umum | *Single-Stage* / Modular | Bergantung Implementasi | Tidak Ada | Python | Faithfulness, Answer Relevance |
| Asai dkk. (2024) (Self-RAG) | Teks Umum & Biologi | *Adaptive Retrieval* | LLM Self-Reflection Token | Tidak Ada | Python | Factual Accuracy |
| **AURA Core (Penelitian Ini, 2026)** | **Regulasi Akademik Perguruan Tinggi (FTI UII)** | **Two-Stage RAG (Dense ANN Top-20 + Cross-Encoder Top-8)** | **Cohere Rerank v3.5 (Cross-Attention)** | **Strict In-Memory Zero-Persistence (KHS/KRS di RAM Sesi)** | **Pure-Go Native (Go 1.24+ Stdlib, Tanpa Framework Eksternal)** | **Automated RAG Triad (CR, G, AR, RTS) via UII-Bench-50** |

---

## 2.2 Landasan Teori

### 2.2.1 *Large Language Models* (LLM) dan Fenomena Halusinasi
*Large Language Models* (LLM) adalah model komputasi probabilistik yang dibangun di atas arsitektur *Transformer Decoder-only* berskala besar. Model dilatih menggunakan korpus teks berukuran terabita untuk memprediksi probabilitas bersyarat kemunculan token berikutnya $w_t$ berdasarkan urutan token sebelumnya $w_1, w_2, \dots, w_{t-1}$:

$$P(W) = \prod_{t=1}^{T} P(w_t \mid w_1, w_2, \dots, w_{t-1}; \Theta)$$

di mana $\Theta$ merepresentasikan himpunan parameter bobot model yang dioptimalkan selama fase pra-pelatihan.

Meskipun menunjukkan kapabilitas luar biasa dalam pemrosesan bahasa alami, LLM memiliki keterbatasan teoretis inheren berupa halusinasi (Ji dkk., 2023). Halusinasi terbagi menjadi dua kategori utama:
1. **Halusinasi Intrinsik (*Intrinsic Hallucination*)**: Keluaran teks yang dihasilkan bertentangan secara langsung dengan informasi yang diberikan dalam konteks masukan.
2. **Halusinasi Ekstrinsik (*Extrinsic Hallucination*)**: Keluaran teks memuat klaim atau fakta baru yang tidak dapat diverifikasi dari konteks masukan maupun dari fakta dunia nyata, namun dirangkai dengan gaya bahasa meyakinkan.

Dalam konteks hukum dan regulasi akademik, halusinasi ekstrinsik sangat berbahaya karena dapat melahirkan "aturan bayangan" fiktif yang menyesatkan mahasiswa.

### 2.2.2 Konsep Dasar *Retrieval-Augmented Generation* (RAG)
Paradigma RAG mengatasi keterbatasan memori statis LLM dengan memisahkan komponen penyimpan pengetahuan (*retriever*) dari komponen penalaran bahasa (*generator*). Proses kerja RAG secara formal dapat dinyatakan sebagai berikut: diberikan kueri pengguna $q$, tahap *retriever* menelusuri himpunan potongan dokumen pendukung $\mathcal{D} = \{d_1, d_2, \dots, d_k\}$ dari korpus eksternal $\mathcal{C}$. Selanjutnya, model generatif menghasilkan respons $a$ dengan mengondisikan distribusi probabilitasnya pada kueri $q$ dan dokumen terambil $\mathcal{D}$:

$$P(a \mid q) = \sum_{d \in \mathcal{D}} P(a \mid q, d) P(d \mid q)$$

Pada implementasi praktis modern, dokumen pendukung $\mathcal{D}$ digabungkan secara langsung ke dalam jendela perintah (*in-context prompt*) model generatif:

$$a \sim \text{LLM}(q, \mathcal{D}; \tau)$$

di mana $\tau$ adalah parameter suhu (*temperature*) yang mengontrol keacakan distribusi pengambilan sampel token.

### 2.2.3 Representasi Vektor Semantik (*Dense Embedding*) dan Pencarian ANN
Untuk menemukan dokumen yang relevan dari korpus yang besar, teks ditransformasikan menjadi representasi vektor numerik padat berkepadatan tinggi (*dense semantic vectors*) menggunakan model *embedding* $f_{\text{embed}}: \mathcal{X} \to \mathbb{R}^d$. Kueri $q$ dipetakan menjadi vektor $\mathbf{v}_q = f_{\text{embed}}(q)$, dan setiap potongan dokumen $p_i$ dipetakan menjadi vektor $\mathbf{v}_{p_i} = f_{\text{embed}}(p_i)$.

Tingkat kemiripan semantik antara kueri dan dokumen dihitung menggunakan metrik *Cosine Similarity*:

$$\text{Sim}_{\cos}(\mathbf{v}_q, \mathbf{v}_{p_i}) = \frac{\mathbf{v}_q \cdot \mathbf{v}_{p_i}}{\|\mathbf{v}_q\|_2 \|\mathbf{v}_{p_i}\|_2} = \frac{\sum_{j=1}^{d} v_{q,j} v_{p_i,j}}{\sqrt{\sum_{j=1}^{d} (v_{q,j})^2} \sqrt{\sum_{j=1}^{d} (v_{p_i,j})^2}}$$

Nilai kemiripan berada pada rentang $[-1, 1]$, di mana nilai mendekati $1$ mengindikasikan kedekatan semantik yang sangat tinggi dalam ruang laten (*latent semantic space*). 

Karena perhitungan kosinus secara *brute-force* terhadap jutaan dokumen membutuhkan waktu $\mathcal{O}(N \cdot d)$, indeks *Approximate Nearest Neighbor* (ANN)—seperti *Hierarchical Navigable Small World* (HNSW)—digunakan untuk menyusutkan kompleksitas penelusuran menjadi $\mathcal{O}(\log N)$, memungkinkan penarikan Top-$K$ kandidat dalam hitungan milidetik.

Namun, model *bi-encoder* menghitung representasi vektor kueri dan dokumen secara independen tanpa *cross-attention*, sehingga rentan terhadap fenomena *semantic drift* ketika menghadapi frasa hukum dengan negasi atau syarat kumulatif yang rumit.

### 2.2.4 *Cross-Encoder Neural Reranking*
*Cross-Encoder* menyelesaikan kelemahan *bi-encoder* dengan melewatkan kueri $q$ dan kandidat dokumen $p_i$ secara bersamaan ke dalam model transformer sebagai satu urutan masukan:

$$\mathbf{x} = \text{[CLS]} \circ q \circ \text{[SEP]} \circ p_i \circ \text{[SEP]}$$

Pada setiap lapisan *self-attention*, setiap token pada kueri berinteraksi secara langsung dengan setiap token pada dokumen kandidat melalui komputasi *Scaled Dot-Product Attention*:

$$\text{Attention}(Q, K, V) = \text{softmax}\left(\frac{QK^T}{\sqrt{d_k}}\right)V$$

Skor relevansi final $s_i \in [0, 1]$ diperoleh dengan melewatkan representasi tersembunyi dari token `[CLS]` pada lapisan terakhir ke sebuah kepala klasifikasi (*classification head*):

$$s_i = \sigma(\mathbf{W} \cdot \mathbf{h}_{\text{[CLS]}} + b)$$

Mekanisme ini memungkinkan model mengevaluasi dependensi sintaksis yang sangat halus, seperti membedakan antara *"Mahasiswa dengan nilai E tidak diperbolehkan sempro"* dan *"Mahasiswa boleh menempuh sempro jika nilai E telah diperbaiki"*.

### 2.2.5 Fenomena *Lost-in-the-Middle*
Penelitian Liu dkk. (2023) membuktikan bahwa kurva performa retrieval-ke-jawaban pada LLM menyerupai bentuk huruf "U". Informasi yang ditempatkan pada posisi awal konteks (persentil 0–20%) dan akhir konteks (persentil 80–100%) memiliki probabilitas pemanggilan (*retrieval recall*) di atas 85%, sedangkan informasi yang berada di tengah konteks (persentil 40–60%) mengalami penurunan probabilitas pemanggilan hingga di bawah 55%.

Oleh karena itu, arsitektur *Two-Stage RAG* berfungsi ganda:
1. Membuang dokumen non-kritis (*noise filtering*).
2. Memadatkan urutan dokumen sehingga klausul paling esensial terkonsentrasi dalam jendela perhatian optimal LLM.

### 2.2.6 Kerangka Evaluasi *RAG Triad* Berbasis *LLM-as-Judge*
Kerangka kerja *RAG Triad* memformalkan evaluasi kualitas RAG menjadi tiga dimensi independen (Saad-Falcon dkk., 2023; Es dkk., 2023):

1. **Context Relevance (CR)**: Mengukur ketepatan tahap *retrieval* dalam mengambil dokumen yang hanya berisi informasi yang relevan dengan pertanyaan kueri $q$:
   $$\text{CR}(q, \mathcal{P}) = \frac{|\{p \in \mathcal{P} : \text{Relevan}(q, p)\}|}{|\mathcal{P}|} \in [0.0, 1.0]$$

2. **Groundedness / Faithfulness (G)**: Mengukur kesetiaan klaim faktual yang dinyatakan dalam jawaban $a$ terhadap konteks dokumen rujukan $\mathcal{P}$. Setiap proposisi klaim $c_i \in a$ diverifikasi apakah didukung secara logis (*entailed*) oleh $\mathcal{P}$:
   $$\text{G}(\mathcal{P}, a) = \frac{\sum_{i=1}^{|a|} \mathbb{I}(\mathcal{P} \models c_i)}{|a|} \in [0.0, 1.0]$$
   di mana $\mathbb{I}(\cdot)$ adalah fungsi indikator biner bernilai $1$ jika klaim terbukti didukung teks rujukan, dan $0$ jika klaim merupakan halusinasi.

3. **Answer Relevance (AR)**: Mengukur seberapa tepat dan langsung jawaban yang dihasilkan $a$ menjawab maksud informasi yang diminta pada kueri $q$:
   $$\text{AR}(q, a) \in [0.0, 1.0]$$

4. **RAG Triad Score (RTS)**: Skor komposit yang dihitung menggunakan rata-rata harmonik (*harmonic mean*) dari ketiga metrik di atas:
   $$\text{RTS} = \frac{3}{\frac{1}{\text{CR}} + \frac{1}{\text{G}} + \frac{1}{\text{AR}}}$$
   Rata-rata harmonik dipilih karena memberikan penalti yang sangat berat (*strong penalty*) apabila salah satu metrik mengalami kegagalan fatal (misalnya terjadi halusinasi yang membuat nilai $G$ anjlok).

### 2.2.7 Karakteristik Bahasa Pemrograman Go untuk Infrastruktur Cerdas
Bahasa pemrograman Go (sering disebut Golang) dirancang oleh Google untuk mengatasi inefisiensi rekayasa pada sistem terdistribusi berdaya tampung besar (Pike, 2012). Beberapa keunggulan komparatif Go yang melandasi pemilihan bahasa dalam penelitian ini meliputi:
1. **Goroutine Concurrency**: Go menggunakan model konkurensi berbasis CSP (*Communicating Sequential Processes*) dengan benang kerja pengguna (*user-space green threads*) yang disebut *goroutine*. Biaya pembuatan goroutine hanya membutuhkan ruang tumpukan (*stack space*) sebesar ~2 KB, dibandingkan benang kerja sistem operasi (*OS threads*) pada Python atau Java yang membutuhkan 1–2 MB. Hal ini memungkinkan peladen AURA Core melayani ribuan koneksi WebSocket secara simultan dengan alokasi memori minimal.
2. **Kompilasi Biner Statis Mandiri (*Single Static Binary*)**: Seluruh kode sumber, dependensi pustaka standar, dan *runtime engine* dikompilasi menjadi satu berkas biner statis tanpa memerlukan instalasi *interpreter* eksternal pada peladen target, meminimalkan potensi kegagalan akibat *dependency drift*.
3. **Pengelolaan Memori Deterministik**: Pengumpul sampah (*Garbage Collector*) pada Go dioptimalkan untuk latensi sub-milidetik, mencegah terjadinya jeda henti layanan (*stop-the-world pauses*) yang umum terjadi pada lingkungan pemrosesan paralel berskala besar.
