package config

import (
	"bufio"
	"os"
	"strings"
)

// loadDotEnv, basit bir .env okuyucu: KEY=VALUE satırlarını os.Setenv ile yükler.
// Zaten set edilmiş env variable'ları EZMEZ (gerçek env her zaman öncelikli).
func loadDotEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		// satır içi yorumu at: VALUE  # açıklama
		if idx := strings.Index(val, " #"); idx >= 0 {
			val = val[:idx]
		}
		val = strings.TrimSpace(val)
		val = strings.Trim(val, `"'`)
		if os.Getenv(key) == "" {
			os.Setenv(key, val)
		}
	}
}
