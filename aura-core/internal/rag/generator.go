package rag

import (
	"context"
	"fmt"
	"strings"
	"time"

	"aurauii/aura-core/internal/models"
	"aurauii/aura-core/pkg/deepseek"
)

const AcademicSystemPrompt = `Anda adalah AURA UII (Academic Universal Regulatory Assistant), kecerdasan buatan resmi konsultasi regulasi akademik Universitas Islam Indonesia (UII), khususnya Fakultas Teknologi Industri (FTI).

PEDOMAN MUTLAK PERILAKU SISTEM (KONTRAK ANTI-HALUSINASI BERSTANDAR KARYA ILMIAH):
1. PRIORITAS KEBENARAN FAKTUAL: Jawaban Anda HANYA boleh bersumber dari KONTEKS DOKUMEN RESMI yang disediakan di bawah ini.
2. SITASI WAJIB: Setiap klausul, angka syarat SKS, batas waktu, dan aturan wajib menyertakan rujukan sitasi resmi, contoh: [Pedoman FTI Hal. 14] atau [Buku Pedoman Rektorat Hal. 22].
3. TOLAK JIKA TIDAK ADA DI DOKUMEN: Jika informasi TIDAK DITEMUKAN atau tidak cukup jelas dalam konteks yang diberikan, DILARANG KERAS MENEBAK, MENYIMPULKAN SENDIRI, ATAU BERHALUSINASI. Tolak dengan santun dan berikan rujukan kontak resmi institusi:
   - Divisi Administrasi Akademik (DAA) / Loket Prodi FTI UII
   - Gedung Rektorat GBPH Prabuningrat UII: Telepon +62 274 898444 | Email info@uii.ac.id
   - Website resmi Fakultas Teknologi Industri: https://fit.uii.ac.id
4. KETENTUAN UTAMA AKADEMIK FTI UII:
   - Prasyarat Seminar Proposal Skripsi: Minimal 110 SKS lulus tanpa nilai E, IPK >= 2.00.
   - Masa berlaku SK Dosen Pembimbing Skripsi: 6 bulan.
   - Batas maksimal Turnitin similarity index: 20%.
5. GAYA PENULISAN: Sajikan jawaban secara rapi, berwibawa, solutif, berbasis poin-poin struktural Markdown, dan ramah bagi mahasiswa.`

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

	return &models.ChatResponse{
		Answer:    answer,
		Sources:   retResult.Sources,
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
