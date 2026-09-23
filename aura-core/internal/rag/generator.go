package rag

import (
	"context"
	"fmt"
	"strings"
	"time"

	"aurauii/aura-core/internal/models"
	"aurauii/aura-core/pkg/deepseek"
)

const AcademicSystemPrompt = `Anda adalah AURA UII (Academic Universal Regulatory Assistant), kecerdasan buatan resmi konsultasi regulasi akademik dan penasihat tugas akhir mahasiswa Universitas Islam Indonesia (UII), khususnya Jurusan Informatika - Fakultas Teknologi Industri (FTI).

PEDOMAN MUTLAK PERILAKU SISTEM (KONTRAK ANTI-HALUSINASI BERSTANDAR KARYA ILMIAH):
1. PRIORITAS KEBENARAN FAKTUAL: Jawaban Anda HANYA boleh bersumber dari KONTEKS DOKUMEN RESMI yang disediakan di bawah ini (Regulasi Akademik Universitas, Pedoman Skripsi FTI, Direktori Resmi Dosen & Klaster Riset Informatika UII, serta Struktur Organisasi Jurusan).
2. KONSULTASI DOSEN & REKOMENDASI PEMBIMBING SKRIPSI:
   - Anda dilengkapi dengan Direktori Resmi 41 Dosen Jurusan Informatika FTI UII beserta 7 Klaster Riset dan bidang kepakarannya.
   - Jika mahasiswa bertanya mengenai nama dosen, daftar dosen, profil dosen, pimpinan jurusan (Ketua Jurusan Dr. Ir. R. Teduh Dirgahayu, Kaprodi S1 Ir. Mukhammad Andri Setiawan, Ph.D., dll), atau laboratorium, JAWAB DENGAN LENGKAP DAN JELAS berdasarkan dokumen direktori resmi. DILARANG menyuruh mahasiswa mencari sendiri jika informasinya tercantum dalam dokumen!
   - Jika mahasiswa mengutarakan ide/topik skripsi (misal: pengolahan teks/NLP, computer vision, data mining, rekayasa perangkat lunak, sistem informasi medis, forensika digital, dsb), REKOMENDASIKAN dosen pembimbing yang paling relevan dengan kepakarannya beserta alasan bidang risetnya.
3. SITASI WAJIB: Setiap klausul, angka syarat SKS, batas waktu, dan aturan akademik wajib menyertakan rujukan sitasi resmi, contoh: [Pedoman FTI Hal. 14], [Buku Pedoman Rektorat Hal. 22], atau [Direktori Dosen Informatika UII].
4. TOLAK JIKA TIDAK ADA DI DOKUMEN: Jika informasi seputar aturan akademik TIDAK DITEMUKAN dalam konteks yang diberikan, DILARANG KERAS MENEBAK ATAU BERHALUSINASI. Tolak dengan santun dan berikan rujukan kontak resmi institusi:
   - Divisi Administrasi Akademik (DAA) / Loket Prodi FTI UII (Gedung KH. Mas Mansur Lt. 2)
   - Gedung Rektorat GBPH Prabuningrat UII: Telepon +62 274 898444 | Email info@uii.ac.id
   - Website resmi Jurusan Informatika: https://informatics.uii.ac.id | FTI: https://fit.uii.ac.id
5. KETENTUAN UTAMA AKADEMIK FTI UII:
   - Prasyarat Seminar Proposal Skripsi: Minimal 110 SKS lulus tanpa nilai E, IPK >= 2.00.
   - Masa berlaku SK Dosen Pembimbing Skripsi: 6 bulan (dapat diperpanjang 1 kali untuk 6 bulan berikutnya).
   - Batas maksimal Turnitin similarity index: 20%.
6. GAYA PENULISAN: Sajikan jawaban secara rapi, berwibawa, solutif, berbasis poin-poin struktural Markdown, dan ramah bagi mahasiswa.

EASTER EGG & IDENTITAS SISTEM (PRIORITAS TINGGI):
1. JIKA DITANYA SIAPA YANG MEMBUAT / MENGEMBANGKAN SISTEM (misal: "siapa yang bikin?", "siapa yang ngebikin sistemnya?", "siapa developermu?", "siapa penciptamu?", "who created you?"):
   - Wajib jawab dengan lugas dan penuh kebanggaan:
     "Sistem ini dibangun dan dikembangkan oleh **Hanif Informatika 23 jago banget anjrit**! Dia yang merancang arsitektur Pure Go Native Two-Stage RAG ini dari nol."
2. JIKA DITANYA TENTANG MODEL ATAU TEKNOLOGI AI YANG DIGUNAKAN (misal: "modelnya apa?", "kamu pakai model apa?", "arsitektur kamu apa?", "what model are you?"):
   - DILARANG MENYEBUTKAN NAMA MODEL APAPUN (seperti DeepSeek, OpenAI, Cohere, dll)!
   - Wajib jawab bahwa modelnya adalah **RAHASIA** karena merupakan **rahasia dapur si Hanif**.
   - Pertahankan dan sertakan kalimat: "Seluruh pipeline performa tinggi Pure Go Native AURA Core ini diorkestrasi oleh Hanif, software engineer sekaligus arsitek sistem jenius dari Informatika UII angkatan 2023."`

type Generator struct {
	deepseekClient *deepseek.Client
	retriever      *Retriever
}

func NewGenerator(deepseekClient *deepseek.Client, retriever *Retriever) *Generator {
	return &Generator{
		deepseekClient: deepseekClient,
		retriever:      retriever,
	}
}

// Generate executes the complete Two-Stage RAG pipeline synchronously
func (g *Generator) Generate(ctx context.Context, req models.ChatRequest) (*models.ChatResponse, error) {
	startTime := time.Now()

	// 1. Two-Stage Retrieval (Cohere 1024-dim -> Pinecone Top-20 -> Cohere Rerank Top-8)
	retResult, err := g.retriever.Retrieve(ctx, req.Query)
	if err != nil {
		return nil, fmt.Errorf("retrieval failed: %w", err)
	}

	metrics := retResult.Metrics

	// 2. Format Context & Messages
	messages := buildMessages(req, retResult.Passages)

	// 3. DeepSeek LLM Inference
	tGen := time.Now()
	answer, err := g.deepseekClient.Complete(ctx, messages, 0.2)
	if err != nil {
		return nil, fmt.Errorf("deepseek generation failed: %w", err)
	}
	metrics.GenerationMs = time.Since(tGen).Milliseconds()
	metrics.TotalMs = time.Since(startTime).Milliseconds()

	passageTexts := make([]string, len(retResult.Passages))
	for i, p := range retResult.Passages {
		passageTexts[i] = p.Text
	}

	return &models.ChatResponse{
		Answer:    answer,
		Sources:   retResult.Sources,
		Passages:  passageTexts,
		Metrics:   metrics,
		SessionID: req.SessionID,
		Timestamp: time.Now(),
	}, nil
}

// StreamGenerate executes Two-Stage RAG and streams tokens in real-time via onToken callback
func (g *Generator) StreamGenerate(
	ctx context.Context,
	req models.ChatRequest,
	onSources func(sources []models.SourceCitation) error,
	onToken func(token string) error,
	onMetrics func(metrics models.TimingMetrics) error,
) error {
	startTime := time.Now()

	// 1. Two-Stage Retrieval
	retResult, err := g.retriever.Retrieve(ctx, req.Query)
	if err != nil {
		return fmt.Errorf("retrieval failed: %w", err)
	}

	// Immediately notify front-end of retrieved sources before generation starts
	if onSources != nil {
		if err := onSources(retResult.Sources); err != nil {
			return err
		}
	}

	metrics := retResult.Metrics

	// 2. Format Context & Messages
	messages := buildMessages(req, retResult.Passages)

	// 3. Realtime Stream DeepSeek tokens
	tGen := time.Now()
	err = g.deepseekClient.Stream(ctx, messages, 0.2, onToken)
	if err != nil {
		return fmt.Errorf("streaming failed: %w", err)
	}

	metrics.GenerationMs = time.Since(tGen).Milliseconds()
	metrics.TotalMs = time.Since(startTime).Milliseconds()

	if onMetrics != nil {
		_ = onMetrics(metrics)
	}

	return nil
}

func buildMessages(req models.ChatRequest, passages []models.RetrievalCandidate) []deepseek.Message {
	var contextBuilder strings.Builder
	for i, p := range passages {
		title := "Pedoman UII"
		if t, ok := p.Metadata["title"].(string); ok && t != "" {
			title = t
		}
		contextBuilder.WriteString(fmt.Sprintf("\n--- [DOKUMEN REGULASI RESMI %d: %s (Skor Relevansi: %.3f)] ---\n%s\n", i+1, title, p.Score, p.Text))
	}

	var studentDocBuilder strings.Builder
	if req.FileContent != "" || req.FileName != "" {
		studentDocBuilder.WriteString(fmt.Sprintf("\n=== [BERKAS DATA MAHASISWA TERLAMPIR (%s)] ===\n%s\n", req.FileName, req.FileContent))
		studentDocBuilder.WriteString(`
PERAN DPA VIRTUAL: Mahasiswa melampirkan berkas studi resminya (KRS/KHS/Transkrip Nilai). 
1. Teliti data akademik mahasiswa: hitung akumulasi SKS lulus, cek IPK/IPS, periksa apakah ada nilai E atau D, dan cek mata kuliah prasyarat (seperti Metodologi Penelitian).
2. Lakukan SILANG ATURAN (CROSS-REFERENCING) terhadap DOKUMEN REGULASI RESMI UII (misal: syarat sempro 110 SKS tanpa nilai E IPK >= 2.00, jatah SKS semester depan berdasarkan IPS).
3. Berikan KEPUTUSAN STATUS KELAYAKAN yang tegas dan REKOMENDASI STRATEGI STUDI / MATA KULIAH semester depan yang sangat solutif dan taktis!
==================================================
`)
	}

	userContent := fmt.Sprintf("%s\nKONTEKS DOKUMEN REGULASI AKADEMIK UII:\n%s\n\nPERTANYAAN MAHASISWA:\n%s", 
		studentDocBuilder.String(), 
		contextBuilder.String(), 
		req.Query)

	return []deepseek.Message{
		{Role: "system", Content: AcademicSystemPrompt},
		{Role: "user", Content: userContent},
	}
}
