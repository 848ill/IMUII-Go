package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"aurauii/aura-core/internal/config"
	"aurauii/aura-core/internal/models"
	"aurauii/aura-core/pkg/chunker"
	"aurauii/aura-core/pkg/cohere"
	"aurauii/aura-core/pkg/pinecone"
)

type CanonicalDoc struct {
	ID       string
	Title    string
	FileName string
	URL      string
	Content  string
}

func main() {
	log.Println("==================================================")
	log.Println("  AURA Core: Seeding Canonical UII Academic Corpus")
	log.Println("==================================================")

	cfg := config.Load()
	ctx := context.Background()

	cohereClient := cohere.NewClient(cfg.CohereAPIKey, "", cfg.CohereEmbedModel, cfg.CohereRerankModel)
	pineconeClient := pinecone.NewClient(cfg.PineconeAPIKey, cfg.PineconeIndex, cfg.PineconeHost)
	splitter := chunker.NewDefaultSplitter()

	docs := getCanonicalCorpus()

	totalChunks := 0
	for _, doc := range docs {
		log.Printf("[Seeder] Processing document: %s (%s)...", doc.Title, doc.FileName)

		rawChunks := splitter.SplitText(doc.Content)
		log.Printf("[Seeder] -> Generated %d chunks", len(rawChunks))

		batchSize := 48
		chunks := make([]models.Chunk, len(rawChunks))

		for i := 0; i < len(rawChunks); i += batchSize {
			end := i + batchSize
			if end > len(rawChunks) {
				end = len(rawChunks)
			}

			subBatch := rawChunks[i:end]
			embeddings, err := cohereClient.Embed(ctx, subBatch, "search_document")
			if err != nil {
				log.Fatalf("[Seeder] Cohere embedding error: %v", err)
			}

			for j, emb := range embeddings {
				idx := i + j
				chunkID := fmt.Sprintf("%s_c%d", doc.ID, idx)
				chunks[idx] = models.Chunk{
					ID:     chunkID,
					DocID:  doc.ID,
					Index:  idx,
					Text:   subBatch[j],
					Vector: emb,
					Metadata: map[string]interface{}{
						"title":        doc.Title,
						"file_name":    doc.FileName,
						"url":          doc.URL,
						"chunk_index":  idx,
						"total_chunks": len(rawChunks),
					},
				}
			}
		}

		if err := pineconeClient.Upsert(ctx, chunks); err != nil {
			log.Fatalf("[Seeder] Pinecone upsert error: %v", err)
		}

		totalChunks += len(chunks)
		log.Printf("[Seeder] Successfully seeded %s into Pinecone index '%s'!\n", doc.Title, cfg.PineconeIndex)
		time.Sleep(500 * time.Millisecond)
	}

	log.Printf("[Seeder] FINISHED! Total %d chunks successfully indexed into Pinecone.\n", totalChunks)
}

func getCanonicalCorpus() []CanonicalDoc {
	return []CanonicalDoc{
		{
			ID:       "pedoman_skripsi_fti",
			Title:    "Buku Pedoman Pelaksanaan Tugas Akhir & Skripsi FTI UII",
			FileName: "Pedoman_Skripsi_FTI_UII_2024.pdf",
			URL:      "https://fit.uii.ac.id/pedoman-tugas-akhir",
			Content: `# BUKU PEDOMAN TUGAS AKHIR DAN SKRIPSI
FAKULTAS TEKNOLOGI INDUSTRI - UNIVERSITAS ISLAM INDONESIA

BAB I: KETENTUAN UMUM DAN PRASYARAT SEMINAR PROPOSAL
Pasal 1: Definisi Tugas Akhir
Tugas Akhir/Skripsi adalah karya ilmiah mandiri yang wajib diselesaikan oleh mahasiswa program sarjana (S1) FTI UII sebagai salah satu prasyarat kelulusan untuk memperoleh gelar Sarjana Teknik (S.T.) atau Sarjana Komputer (S.Kom.).

Pasal 2: Prasyarat Pendaftaran Seminar Proposal (Sempro) Skripsi
Mahasiswa berhak mengajukan pendaftaran Seminar Proposal Skripsi apabila telah memenuhi ketentuan kumulatif sebagai berikut:
1. Telah menyelesaikan beban studi minimal 110 SKS (Satuan Kredit Semester) yang telah dinyatakan lulus.
2. Tidak memiliki nilai E pada mata kuliah yang telah ditempuh dalam transkrip nilai resmi.
3. Memiliki Indeks Prestasi Kumulatif (IPK) sekurang-kurangnya 2.00 (dua koma nol nol).
4. Telah lulus mata kuliah Metodologi Penelitian dengan nilai minimal C.
5. Telah menyelesaikan seluruh kewajiban administrasi akademik dan keuangan semester berjalan.
6. Mengunggah draf proposal yang telah disetujui dan ditandatangani oleh Dosen Pembimbing Skripsi.

BAB II: DOSEN PEMBIMBING DAN MASA BERLAKU SK
Pasal 7: Masa Berlaku Surat Keputusan (SK) Pembimbingan
1. Surat Keputusan (SK) Penugasan Dosen Pembimbing Skripsi berlaku selama 6 (enam) bulan terhitung sejak tanggal diterbitkan oleh Dekanat FTI UII.
2. Apabila dalam kurun waktu 6 bulan mahasiswa belum dapat menyelesaikan ujian pendadaran/sidang skripsi, mahasiswa wajib mengajukan permohonan perpanjangan SK Pembimbing ke Divisi Administrasi Akademik (DAA) FTI dengan melampirkan lembar kemajuan bimbingan (logbook).
3. Perpanjangan SK hanya dapat diberikan maksimal 1 (satu) kali untuk durasi 6 bulan berikutnya.
4. Apabila perpanjangan telah habis dan skripsi belum selesai, Ketua Program Studi berhak mengevaluasi dan mengganti topik atau Dosen Pembimbing.

BAB III: KETENTUAN PLAGIARISME DAN TURNITIN
Pasal 12: Ambang Batas Uji Kesamaan (Similarity Index)
1. Setiap naskah Tugas Akhir/Skripsi yang akan didaftarkan untuk Ujian Pendadaran wajib melalui pemeriksaan kesamaan naskah (anti-plagiarisme) menggunakan perangkat lunak resmi Turnitin yang dikelola oleh Pengelola Jurnal/Perpustakaan FTI UII.
2. Batas toleransi maksimal Similarity Index Turnitin adalah 20% (dua puluh persen), dengan ketentuan mengecualikan daftar pustaka (exclude bibliography) dan kutipan langsung yang sah (exclude quotes).
3. Naskah yang memiliki indeks kesamaan lebih dari 20% wajib direvisi dan diajukan uji ulang setelah mendapatkan bimbingan perbaikan sitasi.

BAB IV: PRASYARAT UJIAN PENDADARAN DAN YUDISIUM
Pasal 18: Prasyarat Yudisium Sarjana
Kelulusan program sarjana ditetapkan dalam Rapat Yudisium Fakultas dengan syarat:
1. Telah lulus seluruh mata kuliah wajib dan pilihan dengan total minimal 144 SKS.
2. Indeks Prestasi Kumulatif (IPK) minimal 2.00.
3. Nilai D tidak melebihi 10% dari total SKS kumulatif.
4. Tidak memiliki nilai E pada seluruh mata kuliah.
5. Telah lulus Ujian Skripsi/Pendadaran dan mengunggah naskah final yang telah direvisi.
6. Memiliki sertifikat uji kemahiran Bahasa Inggris (CEPT/TOEFL) resmi dari CILACS UII dengan skor minimal yang ditetapkan prodi (minimal 450 atau setara).
7. Telah lulus Ujian Hafalan Al-Qur'an dan sertifikasi keislaman sesuai ketentuan DPPAI UII.`,
		},
		{
			ID:       "buku_pedoman_akademik_uii",
			Title:    "Buku Pedoman Akademik Universitas Islam Indonesia",
			FileName: "Buku_Pedoman_Akademik_UII.pdf",
			URL:      "https://academic.uii.ac.id/pedoman-akademik",
			Content: `# BUKU PEDOMAN AKADEMIK UNIVERSITAS ISLAM INDONESIA
BAB III: BEBAN STUDI DAN RENCANA STUDI MAHASISWA

Pasal 10: Beban SKS Semester Berdasarkan IPK
Beban kredit studi (SKS) yang dapat diambil mahasiswa pada semester reguler ditentukan berdasarkan Indeks Prestasi Semester (IPS) yang diperoleh pada semester sebelumnya, dengan rincian:
1. IPS >= 3.00: Mahasiswa berhak mengambil beban studi maksimal hingga 24 SKS.
2. IPS 2.50 - 2.99: Mahasiswa berhak mengambil beban studi maksimal hingga 21 SKS.
3. IPS 2.00 - 2.49: Mahasiswa berhak mengambil beban studi maksimal hingga 18 SKS.
4. IPS 1.50 - 1.99: Mahasiswa berhak mengambil beban studi maksimal hingga 15 SKS.
5. IPS < 1.50: Mahasiswa hanya berhak mengambil beban studi maksimal hingga 12 SKS.

Pasal 15: Prosedur Cuti Akademik
1. Mahasiswa berhak mengajukan cuti akademik setelah menempuh studi sekurang-kurangnya 2 (dua) semester berturut-turut pada program sarjana.
2. Cuti akademik diberikan maksimal 2 (dua) semester, baik berturut-turut maupun terpisah, selama masa studi.
3. Masa cuti akademik resmi tidak diperhitungkan dalam evaluasi batas masa studi maksimal mahasiswa.
4. Permohonan cuti diajukan secara daring melalui sistem informasi akademik sebelum batas akhir masa KRS semester berjalan ditutup.

Pasal 22: Evaluasi Batas Masa Studi dan Sanksi Drop Out (DO)
1. Evaluasi Tahap I (Akhir Tahun Kedua / Semester 4): Mahasiswa wajib memperoleh minimal 35 SKS lulus dengan IPK minimal 2.00.
2. Evaluasi Tahap II (Akhir Tahun Keempat / Semester 8): Mahasiswa wajib memperoleh minimal 80 SKS lulus dengan IPK minimal 2.00.
3. Batas maksimal masa studi program sarjana (S1) di UII adalah 14 (empat belas) semester atau 7 tahun akademik. Mahasiswa yang melampaui batas tersebut dan belum memenuhi syarat kelulusan akan diberhentikan status kemahasiswaannya (Drop Out).`,
		},
		{
			ID:       "beasiswa_layanan_uii",
			Title:    "Panduan Program Beasiswa & Bantuan Biaya Studi UII",
			FileName: "Panduan_Beasiswa_DPK_DPPAI_UII.pdf",
			URL:      "https://kemahasiswaan.uii.ac.id/beasiswa",
			Content: `# PANDUAN PROGRAM BEASISWA UNIVERSITAS ISLAM INDONESIA
DIREKTORAT PEMBINAAN KEMAHASISWAAN (DPK) & DPPAI UII

BAB I: MACAM-MACAM PROGRAM BEASISWA
1. Beasiswa Santri Unggulan UII (BSU):
   - Diperuntukkan bagi santri berprestasi dari pondok pesantren di seluruh Indonesia.
   - Komponen pembiayaan: Bebas biaya kuliah SPP, dana catur dharma, biaya hidup bulanan, dan asrama mahasiswa di Pondok Pesantren UII.
   - Prasyarat: Memiliki hafalan Al-Qur'an minimal 10 Juz, rekomendasi pengasuh pesantren, dan lolos seleksi wawancara keagamaan.

2. Beasiswa Hafiz Al-Qur'an 30 Juz:
   - Diperuntukkan bagi calon mahasiswa yang telah menyelesaikan hafalan 30 Juz Al-Qur'an mutqin.
   - Memperoleh pembebasan SPP dan biaya kuliah 100% selama 8 semester dengan syarat menjaga hafalan dan mempertahankan IPK minimal 3.25 setiap semester.

3. Beasiswa Bantuan Keuangan DPK & Keringanan Biaya Studi:
   - Bagi mahasiswa aktif yang mengalami kendala ekonomi mendesak di tengah masa studi.
   - Mahasiswa dapat mengajukan permohonan penundaan pembayaran angsuran SPP atau pemotongan dana bantuan melalui verifikasi berkas Surat Keterangan Tidak Mampu (SKTM) dan slip gaji orang tua kepada DPK UII.`,
		},
		{
			ID:       "uii_kontak_akreditasi",
			Title:    "Informasi Kontak Resmi & Akreditasi Universitas Islam Indonesia",
			FileName: "Informasi_Kontak_Akreditasi_UII.pdf",
			URL:      "https://www.uii.ac.id/kontak",
			Content: `# INFORMASI KONTAK RESMI DAN STATUS AKREDITASI INSTITUSI
UNIVERSITAS ISLAM INDONESIA (UII)

1. Alamat Kampus Terpadu:
   Gedung GBPH Prabuningrat (Kantor Rektorat UII)
   Jl. Kaliurang km. 14,5 Sleman, D.I. Yogyakarta 55584 Indonesia
   Telepon Resmi: +62 274 898444 (Hunting)
   Faks: +62 274 898459
   Email Layanan Informasi: info@uii.ac.id
   Website Resmi Universitas: https://www.uii.ac.id

2. Kontak Fakultas Teknologi Industri (FTI UII):
   Gedung KH. Mas Mansur, Kampus Terpadu UII
   Jl. Kaliurang km. 14,5 Sleman, Yogyakarta
   Telepon: +62 274 895287 | Email: fti@uii.ac.id
   Website Resmi Fakultas: https://fit.uii.ac.id

3. Divisi Administrasi Akademik (DAA UII):
   Layanan permohonan surat keterangan aktif kuliah, legalisir ijazah, transkrip nilai, dan administrasi yudisium.
   Email: akademik@uii.ac.id | Jam Layanan: Senin - Jumat 08.00 - 15.30 WIB

4. Status Akreditasi Institusi:
   Universitas Islam Indonesia telah meraih predikat AKREDITASI INSTITUSI UNGGUL dari Badan Akreditasi Nasional Perguruan Tinggi (BAN-PT) berdasarkan Surat Keputusan No. 192/SK/BAN-PT/Ak/PT/III/2022. Mayoritas program studi sarjana di lingkungan FTI UII juga telah terakreditasi Unggul dan beberapa telah meraih akreditasi internasional IABEE.`,
		},
		{
			ID:       "direktori_dosen_informatika_uii",
			Title:    "Direktori Resmi Staf Pengajar & Profil Dosen Jurusan Informatika FTI UII",
			FileName: "Direktori_Dosen_Informatika_FTI_UII.pdf",
			URL:      "https://informatics.uii.ac.id/profil/dosen_if/",
			Content: `# DIREKTORI RESMI DOSEN & KLASTER RISET JURUSAN INFORMATIKA
FAKULTAS TEKNOLOGI INDUSTRI - UNIVERSITAS ISLAM INDONESIA (FTI UII)

Sumber Resmi: https://informatics.uii.ac.id/profil/dosen_if/ dan https://informatics.uii.ac.id/dosen-berdasarkan-klaster/

BAB I: PEMBAGIAN KLASTER RISET & KEPASANGAN TOPIK SKRIPSI

1. Klaster Sains Data & Kecerdasan Buatan (Data Science & Artificial Intelligence):
   Fokus Riset: Natural Language Processing (NLP), Text Mining, Deep Learning, Big Data Analytics, Predictive Modeling, Causal Modeling, Machine Learning Terapan, Data Mining.
   Rekomendasi Dosen Pembimbing Skripsi:
   - Dr. Syarif Hidayat, S.Kom., M.I.T. (Bidang: Data Mining, Kecerdasan Buatan, Embedded System)
   - Ahmad Fathan Hidayatullah, S.T., M.Cs., Ph.D. (Bidang: Natural Language Processing, Sains Data, Text Mining)
   - Ir. Dhomas Hatta Fudholi, S.T., M.Eng., Ph.D., IPM., ASEAN Eng. (Bidang: Big Data, Deep Learning, NLP, Ontologi, Sains Data)
   - Dr. Feri Wijayanto, S.T., M.T. (Bidang: Machine Learning, Model Probabilistik, Pemodelan Causal, Psikometrik, Sains Data)
   - Lizda Iswari, S.T., M.Sc. (Bidang: Data Profiling, Data Clustering, Visualisasi Data)
   - Septia Rani, S.T., M.Cs. (Bidang: Sains Data, Kecerdasan Buatan, Information Hiding)

2. Klaster Informatika Medis (Medical Informatics):
   Fokus Riset: Sistem Informasi Kesehatan, Clinical Decision Support System (CDSS), Pemrosesan Citra Medis, Bioinformatika, Rekam Medis Elektronik.
   Rekomendasi Dosen Pembimbing Skripsi:
   - Prof. Dr. Sri Kusumadewi, S.Si., M.T. (Guru Besar Sistem Cerdas & Informatika Medis)
   - Ir. Izzati Muhimmah, S.T., M.Sc., Ph.D. (Bidang: Informatika Medis, Pencitraan Medis, Visi Komputer)
   - Aridhanyati Arifin, S.T., M.Cs. (Bidang: Informatika Medis, Sistem Pendukung Keputusan)
   - Elyza Gustri Wahyuni, S.T., M.Cs. (Bidang: Informatika Medis, Sistem Pendukung Keputusan)
   - Chanifah Indah Ratnasari, S.Kom., M.Kom. (Bidang: Informatika Medis, Ekstraksi Informasi, NLP)
   - Rahadian Kurniawan, S.Kom., M.Kom. (Bidang: Sistem Informasi Kesehatan, Pemrosesan Citra Medis, Gim Serius Medis)

3. Klaster Rekayasa Perangkat Lunak (Software Engineering):
   Fokus Riset: Arsitektur Perangkat Lunak, Software Testing & Quality Assurance, Requirement Engineering, DevOps, Microservices, Service Computing, Metodologi Agile/Scrum.
   Rekomendasi Dosen Pembimbing Skripsi:
   - Dr. Ir. Raden Teduh Dirgahayu, S.T., M.Sc. (Ketua Jurusan Informatika, Bidang: Rekayasa Perangkat Lunak, Service Computing, Rekayasa Enterprise)
   - Beni Suranto, S.T., M.SoftEng. (Bidang: Rekayasa Perangkat Lunak, Software Design)
   - Andhik Budi Cahyono, S.T., M.T. (Bidang: Rekayasa Perangkat Lunak, Web & Mobile Development)
   - Hari Setiaji, S.Kom., M.Eng. (Bidang: Rekayasa Perangkat Lunak, Sistem Informasi, Teknologi Basis Data)
   - Dr. Novi Setiani, S.T., M.T. (Bidang: Software Testing, Requirement Engineering, Computer Science Education)
   - Hanson Prihantoro Putro, S.T., M.T. (Bidang: Software Testing, Arsitektur Enterprise, Pemrograman Kompetitif)

4. Klaster Forensika Digital & Keamanan Siber (Digital Forensics & Cybersecurity - PUSFID):
   Fokus Riset: Analisis Bukti Digital, Incident Response, Network Security, Malware Analysis, Ethical Hacking, Hukum Siber, Steganografi & Watermarking, Cloud Forensics.
   Rekomendasi Dosen Pembimbing Skripsi:
   - Dr. Yudi Prayudi, S.Si., M.Kom. (Kepala Pusat Studi Forensika Digital / PUSFID, Bidang: Bukti Digital, Forensika Digital, Hukum Siber, Steganografi, Watermarking)
   - Dr. Ahmad Luthfi, S.Kom., M.Kom. (Bidang: Forensika Digital, Jaringan Komputer, Keamanan Jaringan, Open Government Data)
   - Fietyata Yudha, S.Kom., M.Kom., Ph.D. (Bidang: Ethical Hacking, Forensika Digital, Keamanan Jaringan, Keamanan Siber)
   - Erika Ramadhani, S.T., M.Eng. (Bidang: Forensika Digital, Keamanan Komputer)
   - Fayruz Rahma, S.T., M.Eng. (Bidang: Forensika Jaringan, Jaringan Komputer, Keamanan Siber)

5. Klaster Sistem Cerdas, Visi Komputer & Robotika (Intelligent Systems & Computer Vision):
   Fokus Riset: Computer Vision, Object Detection, Autonomous Systems, Robotika, Reinforcement Learning, Neural Networks, Soft Computing, Algoritma Optimasi.
   Rekomendasi Dosen Pembimbing Skripsi:
   - Ir. Chandra Kusuma Dewa, S.Kom., M.Kom., Ph.D. (Bidang: Machine Learning, Multimedia, Reinforcement Learning, Robotika)
   - Arrie Kurniawardhani, S.Si., M.Kom. (Bidang: Machine Learning, Visi Komputer, Pengenalan Pola)
   - Rian Adam Rajagede, S.Kom., M.Cs. (Bidang: Jaringan Syaraf Tiruan, Machine Learning, Deep Learning)
   - Taufiq Hidayat, S.T., M.Sc., Ph.D. (Bidang: Sistem Cerdas, Logika, Soft Computing, Representasi Pengetahuan)
   - Sri Mulyati, S.Kom., M.Kom. (Bidang: Sistem Cerdas, Informatika Teori)
   - Zainudin Zukhri, S.T., MIT. (Bidang: Kecerdasan Buatan, Optimasi, Pengenalan Pola)

6. Klaster Sistem Informasi Enterprise & Tata Kelola IT (Enterprise Systems & IT Governance):
   Fokus Riset: IT Governance (COBIT), Enterprise Architecture (TOGAF), Business Process Management (BPM), ERP, Business Intelligence, Linked Data, Semantic Web, e-Government.
   Rekomendasi Dosen Pembimbing Skripsi:
   - Prof. Fathul Wahid, S.T., M.Sc., Ph.D. (Rektor UII, Bidang: e-Government, e-Participation, ICT4D, Sistem Enterprise)
   - Dr. Hendrik, S.T., M.Eng. (Bidang: Business Intelligence, Linked Data, Semantic Web, Sistem Informasi, Teknologi Pembelajaran)
   - Kholid Haryono, S.T., M.Kom. (Bidang: IT Governance, Audit dan Kontrol Sistem Informasi, TRIZ / Inventive Problem Solving, Enterprise Information Systems)
   - Ir. Mukhammad Andri Setiawan, S.T., M.Sc., Ph.D. (Kaprodi S1 Informatika, Bidang: Manajemen Proses Bisnis, Keamanan Informasi, Cloud Infrastructure)
   - Ari Sujarwo, S.Kom., MIT. (Hons) (Bidang: Sistem Informasi, Kebijakan Publik, Internet of Things)
   - Moh. Idris, S.Kom., M.Kom. (Bidang: Sistem Informasi, Jaringan Komputer)
   - Dr. Nur Wijayaning Rahayu, S.Kom., M.Cs. (Bidang: Sistem Informasi, Basis Data, Teknologi Pendidikan)

7. Klaster Game Technology, Interaksi Manusia-Komputer & Multimedia (Game, HCI & Edu-Tech):
   Fokus Riset: Gamifikasi, Game-based Learning, Augmented Reality/Virtual Reality (AR/VR), UI/UX Design, Human-Computer Interaction, Adaptive E-Learning.
   Rekomendasi Dosen Pembimbing Skripsi:
   - Galang Prihadi Mahardhika, S.Kom., M.Kom. (Bidang: Game Based Learning, Gamifikasi, Interaksi Game)
   - Almed Hamzah, S.T., M.Eng. (Bidang: Interaksi Manusia dan Komputer / HCI, Aplikasi Web Adaptif, M-Learning, UI/UX)
   - Sheila Nurul Huda, S.Kom., M.Cs. (Bidang: Ilmu Komputer, Multimedia Interaktif)

8. Klaster Jaringan Komputer, Komunikasi Nirkabel & IoT (Computer Networks & IoT):
   Fokus Riset: Wireless Sensor Networks, Cognitive Radio, Internet of Things (IoT), Network Coding, Routing Protocols, Smart Campus Infrastructure.
   Rekomendasi Dosen Pembimbing Skripsi:
   - Ir. Kurniawan Dwi Irianto, S.T., M.Sc. (Bidang: Cognitive Radio Networks, Internet of Things, Komunikasi Nirkabel, Network Coding)
   - Ir. Irving Vitra Paputungan, S.T., M.Sc., Ph.D. (Bidang: Internet of Things, Algoritma Genetika, Optimisasi, Basis Data)

BAB II: DAFTAR LENGKAP 41 DOSEN JURUSAN INFORMATIKA FTI UII (URUT ABJAD)
1. Ahmad Fathan Hidayatullah, S.T., M.Cs., Ph.D. | Kepakaran: NLP, Sains Data, Text Mining
2. Dr. Ahmad Luthfi, S.Kom., M.Kom. | Kepakaran: Forensika Digital, Jaringan Komputer, Open Government Data
3. Aridhanyati Arifin, S.T., M.Cs. | Kepakaran: Informatika Medis, Sistem Pendukung Keputusan
4. Andhik Budi Cahyono, S.T., M.T. | Kepakaran: Rekayasa Perangkat Lunak, Web/Mobile Development
5. Ari Sujarwo, S.Kom., MIT. (Hons) | Kepakaran: IoT, Kebijakan Publik, Sistem Informasi
6. Arrie Kurniawardhani, S.Si., M.Kom. | Kepakaran: Machine Learning, Computer Vision
7. Beni Suranto, S.T., M.SoftEng. | Kepakaran: Rekayasa Perangkat Lunak, Software Architecture
8. Ir. Chandra Kusuma Dewa, S.Kom., M.Kom., Ph.D. | Kepakaran: Machine Learning, Reinforcement Learning, Robotika, Multimedia
9. Chanifah Indah Ratnasari, S.Kom., M.Kom. | Kepakaran: Ekstraksi Informasi, Informatika Medis, NLP, Sistem Informasi
10. Ir. Dhomas Hatta Fudholi, S.T., M.Eng., Ph.D., IPM., ASEAN Eng. | Kepakaran: Big Data, Deep Learning, NLP, Ontologi, Sains Data
11. Elyza Gustri Wahyuni, S.T., M.Cs. | Kepakaran: Informatika Medis, Sistem Pendukung Keputusan
12. Erika Ramadhani, S.T., M.Eng. | Kepakaran: Forensika Digital, Keamanan Komputer
13. Dr. Feri Wijayanto, S.T., M.T. | Kepakaran: Machine Learning, Model Probabilistik, Pemodelan Causal, Sains Data
14. Galang Prihadi Mahardhika, S.Kom., M.Kom. | Kepakaran: Game Based Learning, Gamifikasi
15. Hari Setiaji, S.Kom., M.Eng. | Kepakaran: Rekayasa Perangkat Lunak, Sistem Informasi, Teknologi Basis Data
16. Dr. Hendrik, S.T., M.Eng. | Kepakaran: Business Intelligence, Linked Data, Semantic Web, Sistem Informasi
17. Ir. Izzati Muhimmah, S.T., M.Sc., Ph.D. | Kepakaran: Informatika Medis, Pencitraan Medis, Visi Komputer
18. Kholid Haryono, S.T., M.Kom. | Kepakaran: Audit & Kontrol, IT Governance, Sistem Informasi Enterprise, TRIZ
19. Ir. Kurniawan Dwi Irianto, S.T., M.Sc. | Kepakaran: Cognitive Radio Networks, IoT, Komunikasi Nirkabel
20. Prof. Dr. Sri Kusumadewi, S.Si., M.T. | Kepakaran: Informatika Medis, Sistem Pakar, Fuzzy Logic
21. Sheila Nurul Huda, S.Kom., M.Cs. | Kepakaran: Ilmu Komputer, Algoritma Pemrograman
22. Dr. Novi Setiani, S.T., M.T. | Kepakaran: Software Testing, Requirement Engineering, Computer Science Education
23. Moh. Idris, S.Kom., M.Kom. | Kepakaran: Jaringan Komputer, Sistem Informasi
24. Dr. Syarif Hidayat, S.Kom., M.I.T. | Kepakaran: Data Mining, Kecerdasan Buatan, Embedded System
25. Zainudin Zukhri, S.T., MIT. | Kepakaran: Kecerdasan Buatan, Optimasi, Pengenalan Pola
26. Dr. Yudi Prayudi, S.Si., M.Kom. | Kepakaran: Forensika Digital, Bukti Digital, Hukum Siber, Steganografi
27. Sri Mulyati, S.Kom., M.Kom. | Kepakaran: Informatika Teori, Sistem Cerdas
28. Ir. Irving Vitra Paputungan, S.T., M.Sc., Ph.D. | Kepakaran: IoT, Algoritma Genetika, Optimisasi, Basis Data
29. Hanson Prihantoro Putro, S.T., M.T. | Kepakaran: Arsitektur Enterprise, Pemrograman Kompetitif, Pengujian Perangkat Lunak
30. Lizda Iswari, S.T., M.Sc. | Kepakaran: Data Profiling, Data Clustering, Visualisasi Data
31. Fietyata Yudha, S.Kom., M.Kom., Ph.D. | Kepakaran: Ethical Hacking, Forensika Digital, Keamanan Jaringan
32. Prof. Fathul Wahid, S.T., M.Sc., Ph.D. | Kepakaran: e-Government, e-Participation, ICT4D, Enterprise Systems
33. Ir. Mukhammad Andri Setiawan, S.T., M.Sc., Ph.D. | Kepakaran: Manajemen Proses Bisnis (BPM), Keamanan Informasi, Cloud
34. Fayruz Rahma, S.T., M.Eng. | Kepakaran: Forensika Jaringan, Jaringan Komputer, Keamanan Siber
35. Dr. Nur Wijayaning Rahayu, S.Kom., M.Cs. | Kepakaran: Basis Data, Pendidikan Komputer, Sistem Informasi
36. Dr. Ir. Raden Teduh Dirgahayu, S.T., M.Sc. | Kepakaran: Rekayasa Enterprise, Rekayasa Perangkat Lunak, Service Computing
37. Almed Hamzah, S.T., M.Eng. | Kepakaran: Aplikasi Web Adaptif, HCI, M-Learning, UI/UX
38. Rahadian Kurniawan, S.Kom., M.Kom. | Kepakaran: Gim Serius, Pemrosesan Citra Medis, Sistem Informasi Kesehatan
39. Rian Adam Rajagede, S.Kom., M.Cs. | Kepakaran: Jaringan Syaraf Tiruan, Machine Learning, Deep Learning
40. Septia Rani, S.T., M.Cs. | Kepakaran: Information Hiding, Kecerdasan Buatan, Sains Data
41. Taufiq Hidayat, S.T., M.Sc., Ph.D. | Kepakaran: Akuisisi Pengetahuan, Logika, Machine Learning, Sistem Cerdas, Soft Computing`,
		},
		{
			ID:       "struktur_organisasi_dan_laboratorium_informatika_uii",
			Title:    "Struktur Organisasi, Pimpinan, dan Laboratorium Jurusan Informatika FTI UII",
			FileName: "Struktur_Organisasi_Laboratorium_Informatika_UII.pdf",
			URL:      "https://informatics.uii.ac.id/profil/struktur-organisasi/",
			Content: `# STRUKTUR ORGANISASI, LABORATORIUM, DAN LAYANAN JURUSAN INFORMATIKA FTI UII

Sumber Resmi: https://informatics.uii.ac.id/profil/struktur-organisasi/ dan https://informatics.uii.ac.id/profil/laboratorium/

BAB I: STRUKTUR PIMPINAN JURUSAN & PROGRAM STUDI INFORMATIKA FTI UII
1. Pimpinan Jurusan:
   - Ketua Jurusan Informatika: Dr. Ir. Raden Teduh Dirgahayu, S.T., M.Sc.
   - Sekretaris Jurusan Informatika: Dr. Hendrik, S.T., M.Eng.
   - Lokasi Kantor Jurusan: Gedung KH. Mas Mansur Lantai 2, Fakultas Teknologi Industri (FTI), Kampus Terpadu UII, Jl. Kaliurang KM. 14,5 Sleman, Yogyakarta.

2. Program Studi Sarjana (S1):
   - Ketua Program Studi Sarjana Informatika (S1 Reguler & International Program): Ir. Mukhammad Andri Setiawan, S.T., M.Sc., Ph.D.
   - Sekretaris Program Studi Sarjana Informatika (S1 Reguler & IP): Chanifah Indah Ratnasari, S.Kom., M.Kom.
   - Ketua Program Studi Informatika Program Sarjana PJJ (Pendidikan Jarak Jauh): Fietyata Yudha, S.Kom., M.Kom., Ph.D.

3. Program Studi Pascasarjana (S2 & S3):
   - Ketua Program Studi Magister Informatika (S2): Dr. Ahmad Luthfi, S.Kom., M.Kom.
   - Sekretaris Program Studi Magister Informatika (S2): Ahmad Fathan Hidayatullah, S.T., M.Cs., Ph.D.
   - Ketua Program Studi Doktor Informatika (S3): Dr. Novi Setiani, S.T., M.T.

BAB II: LABORATORIUM DAN PUSAT STUDI
1. Laboratorium Komputer Jurusan Informatika FTI UII:
   - Laboratorium Basis Data & Rekayasa Perangkat Lunak (Database & Software Engineering Lab)
   - Laboratorium Sistem Cerdas & Visi Komputer (Intelligent Systems Lab)
   - Laboratorium Komputasi Terdistribusi, Jaringan, dan IoT (Networking & IoT Lab)
   - Laboratorium Multimedia & Game Technology
   - Laboratorium Forensika Digital & Keamanan Komputer (Cybersecurity & Digital Evidence Lab)
   - Laboratorium Pemrograman Dasar & Algoritma

2. Pusat Studi Riset (Research Centers):
   - Pusat Studi Forensika Digital (PUSFID UII) - Website: https://forensics.uii.ac.id (Kepala: Dr. Yudi Prayudi, S.Si., M.Kom.)
   - Pusat Studi Informatika Medis (PSIM UII) (Dipimpin oleh Prof. Dr. Sri Kusumadewi & Ir. Izzati Muhimmah, Ph.D.)
   - Pusat Studi Sistem Informasi Enterprise (PS-SIE)
   - Pusat Studi Sains Data (Data Science Center)

BAB III: PANDUAN PENGAJUAN SKRIPSI DAN PEMILIHAN DOSEN PEMBIMBING
1. Penentuan Topik Skripsi:
   - Mahasiswa memilih topik skripsi yang sesuai minat dan termasuk dalam salah satu klaster riset di Jurusan Informatika UII.
   - Mahasiswa dianjurkan membaca publikasi ilmiah terbaru calon dosen pembimbing sebelum mengajukan proposal.
2. Konsultasi Proposal & Pengajuan SK Pembimbing:
   - Mahasiswa menyusun draf proposal skripsi (Bab 1-3).
   - Mengajukan permohonan calon dosen pembimbing ke Program Studi Informatika FTI UII melalui Divisi Administrasi Akademik (DAA).
   - Setelah SK Dekan FTI terbit, SK Pembimbing berlaku selama 6 (enam) bulan dan dapat diperpanjang maksimal 1 kali (6 bulan berikutnya).`,
		},
	}
}
