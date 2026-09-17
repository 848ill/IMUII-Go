package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"aurauii/aura-core/internal/models"
)

type UploadResponse struct {
	Success          bool   `json:"success"`
	Type             string `json:"type"` // "student_doc" or "regulation"
	FileName         string `json:"fileName"`
	FileURL          string `json:"fileUrl"`
	ExtractedContent string `json:"extractedContent,omitempty"`
	Summary          string `json:"summary,omitempty"`
	Message          string `json:"message"`
}

// Upload handles both Student Documents (KRS/KHS/Transkrip) and Admin Regulation Ingestion
func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 25MB max upload
	if err := r.ParseMultipartForm(25 << 20); err != nil {
		http.Error(w, "File too large (max 25MB)", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to read uploaded file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	uploadType := r.FormValue("type")
	if uploadType == "" {
		uploadType = "student_doc" // default: student document
	}

	sessionID := r.FormValue("sessionId")
	if sessionID == "" {
		sessionID = fmt.Sprintf("session-%d", time.Now().Unix())
	}

	fileData, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read file content: "+err.Error(), http.StatusInternalServerError)
		return
	}

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	cleanFileName := strings.ReplaceAll(header.Filename, " ", "_")
	storagePath := fmt.Sprintf("%s/%d_%s", uploadType, time.Now().Unix(), cleanFileName)

	var publicURL string
	if h.supabaseClient != nil {
		pURL, err := h.supabaseClient.UploadFile(r.Context(), "chat-files", storagePath, contentType, fileData)
		if err != nil {
			log.Printf("[Upload] Warning: Supabase storage upload failed (%v), proceeding with in-memory parsing", err)
		} else {
			publicURL = pURL
		}
	}

	// Case 1: Student Document (KRS, KHS, Transkrip)
	if uploadType == "student_doc" {
		var extractedText string

		// If public URL is available and file is PDF, use Jina Reader for high-accuracy layout & table extraction
		if publicURL != "" && (ext == ".pdf" || ext == ".html") && h.jinaClient != nil {
			readerResp, err := h.jinaClient.ReadURL(r.Context(), publicURL)
			if err == nil && len(readerResp.Data.Content) > 0 {
				extractedText = readerResp.Data.Content
			}
		}

		// Fallback for text/csv files or if Jina fails
		if extractedText == "" && (ext == ".txt" || ext == ".csv" || ext == ".md") {
			extractedText = string(fileData)
		}

		if extractedText == "" {
			extractedText = fmt.Sprintf("Berkas akademik %s berhasil diunggah (%d bytes). Mohon sebutkan rincian nilai atau SKS yang ingin Anda konsultasikan jika teks tabel tidak otomatis terbaca.", header.Filename, len(fileData))
		}

		summary := generateStudentDocSummary(header.Filename, extractedText)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(UploadResponse{
			Success:          true,
			Type:             "student_doc",
			FileName:         header.Filename,
			FileURL:          publicURL,
			ExtractedContent: extractedText,
			Summary:          summary,
			Message:          "Dokumen akademik mahasiswa berhasil diunggah dan dianalisis!",
		})
		return
	}

	// Case 2: Regulation Document (Admin Knowledge Base Ingestion)
	if uploadType == "regulation" {
		if publicURL == "" {
			http.Error(w, "Supabase storage URL required for regulation ingestion", http.StatusInternalServerError)
			return
		}

		docTitle := r.FormValue("title")
		if docTitle == "" {
			docTitle = strings.TrimSuffix(header.Filename, ext)
		}

		doc := models.Document{
			ID:       fmt.Sprintf("reg_%d", time.Now().Unix()),
			Title:    docTitle,
			FileName: header.Filename,
			URL:      publicURL,
		}

		// Asynchronously or synchronously ingest into Pinecone
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
			defer cancel()
			if err := h.ingestor.IngestDocument(ctx, doc); err != nil {
				log.Printf("[Upload] Ingestion failed for %s: %v", header.Filename, err)
			} else {
				log.Printf("[Upload] Regulation %s successfully ingested into Pinecone!", header.Filename)
			}
		}()

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(UploadResponse{
			Success:  true,
			Type:     "regulation",
			FileName: header.Filename,
			FileURL:  publicURL,
			Message:  fmt.Sprintf("Dokumen regulasi '%s' sedang diindeks ke Pinecone secara otomatis.", docTitle),
		})
		return
	}

	http.Error(w, "Invalid upload type", http.StatusBadRequest)
}

func generateStudentDocSummary(filename, text string) string {
	lower := strings.ToLower(text)
	docType := "Dokumen Akademik"
	if strings.Contains(lower, "rencana studi") || strings.Contains(lower, "krs") {
		docType = "Kartu Rencana Studi (KRS)"
	} else if strings.Contains(lower, "hasil studi") || strings.Contains(lower, "khs") {
		docType = "Kartu Hasil Studi (KHS)"
	} else if strings.Contains(lower, "transkrip") {
		docType = "Transkrip Nilai Akademik"
	}

	return fmt.Sprintf("%s terdeteksi (%s). Siap dikonsultasikan dengan DPA Virtual.", docType, filename)
}
