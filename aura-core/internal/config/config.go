package config

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Config struct {
	Port int
	Env  string

	// DeepSeek
	DeepSeekAPIKey  string
	DeepSeekBaseURL string
	DeepSeekModel   string

	// Cohere
	CohereAPIKey      string
	CohereEmbedModel  string
	CohereRerankModel string

	// Jina
	JinaAPIKey        string
	JinaReaderBaseURL string

	// Pinecone
	PineconeAPIKey string
	PineconeIndex  string
	PineconeHost   string

	// Supabase
	SupabaseURL     string
	SupabaseAnonKey string
}

// Load reads .env if present and environment variables into Config
func Load() *Config {
	loadDotEnv(".env")

	cfg := &Config{
		Port: getEnvInt("PORT", 8090),
		Env:  getEnv("ENV", "development"),

		DeepSeekAPIKey:  getEnv("DEEPSEEK_API_KEY", ""),
		DeepSeekBaseURL: getEnv("DEEPSEEK_BASE_URL", "https://api.deepseek.com"),
		DeepSeekModel:   getEnv("DEEPSEEK_MODEL", "deepseek-chat"),

		CohereAPIKey:      getEnv("COHERE_API_KEY", ""),
		CohereEmbedModel:  getEnv("COHERE_EMBED_MODEL", "embed-multilingual-v3.0"),
		CohereRerankModel: getEnv("COHERE_RERANK_MODEL", "rerank-v3.5"),

		JinaAPIKey:        getEnv("JINA_API_KEY", ""),
		JinaReaderBaseURL: getEnv("JINA_READER_BASE_URL", "https://r.jina.ai"),

		PineconeAPIKey: getEnv("PINECONE_API_KEY", ""),
		PineconeIndex:  getEnv("PINECONE_INDEX", "imuiirags2"),
		PineconeHost:   getEnv("PINECONE_HOST", ""),

		SupabaseURL:     getEnv("SUPABASE_URL", ""),
		SupabaseAnonKey: getEnv("SUPABASE_ANON_KEY", ""),
	}

	return cfg
}

func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok && strings.TrimSpace(val) != "" {
		return strings.TrimSpace(val)
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val, ok := os.LookupEnv(key); ok && strings.TrimSpace(val) != "" {
		if n, err := strconv.Atoi(strings.TrimSpace(val)); err == nil {
			return n
		}
	}
	return defaultVal
}

// loadDotEnv parses a .env file without external dependencies
func loadDotEnv(filenames ...string) {
	for _, filename := range filenames {
		path := filename
		if !filepath.IsAbs(path) {
			// Also check current directory
			if _, err := os.Stat(path); os.IsNotExist(err) {
				continue
			}
		}

		file, err := os.Open(path)
		if err != nil {
			continue
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				val := strings.TrimSpace(parts[1])
				// Strip quotes if present
				val = strings.Trim(val, `"'`)
				if _, exists := os.LookupEnv(key); !exists {
					_ = os.Setenv(key, val)
				}
			}
		}
	}
}
