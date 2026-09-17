package chunker

import (
	"strings"
	"testing"
)

func TestSplitText(t *testing.T) {
	splitter := NewDefaultSplitter()
	splitter.ChunkSize = 100
	splitter.ChunkOverlap = 20

	text := `Pedoman Akademik FTI UII mengatur beban studi mahasiswa secara sistematis. 
Setiap mahasiswa wajib menempuh minimal 144 SKS untuk kelulusan program sarjana.
Untuk mendaftar Seminar Proposal Skripsi, mahasiswa wajib menyelesaikan minimal 110 SKS tanpa nilai E.`

	chunks := splitter.SplitText(text)
	if len(chunks) == 0 {
		t.Fatalf("expected chunks, got 0")
	}

	for i, chunk := range chunks {
		if len(chunk) > 150 { // allow small margin if separator is long
			t.Errorf("chunk %d exceeded size: %d", i, len(chunk))
		}
		if strings.TrimSpace(chunk) == "" {
			t.Errorf("chunk %d is empty", i)
		}
	}
}
