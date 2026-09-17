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
	}
}
