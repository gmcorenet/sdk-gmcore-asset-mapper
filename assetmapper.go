package gmcore_asset_mapper

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Registry struct {
	mu      sync.RWMutex
	paths   map[string]string
	baseDir string
}

var globalRegistry = &Registry{
	paths:   make(map[string]string),
	baseDir: "",
}

func RegisterPath(bundleName, publicDir string) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	globalRegistry.paths[bundleName] = publicDir
}

func SetPublicDir(dir string) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	globalRegistry.baseDir = dir
}

func GetPath(bundleName string) (string, bool) {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()
	path, ok := globalRegistry.paths[bundleName]
	return path, ok
}

func AllPaths() map[string]string {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()
	result := make(map[string]string, len(globalRegistry.paths))
	for k, v := range globalRegistry.paths {
		result[k] = v
	}
	return result
}

func ResolveAsset(assetPath string) string {
	return "/assets/" + strings.TrimPrefix(assetPath, "/")
}

func AssetHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assetPath := strings.TrimPrefix(r.URL.Path, "/assets/")
		if assetPath == "" || assetPath == r.URL.Path {
			http.NotFound(w, r)
			return
		}

		globalRegistry.mu.RLock()
		defer globalRegistry.mu.RUnlock()

		for bundleName, bundleDir := range globalRegistry.paths {
			prefix := "bundles/" + bundleName + "/"
			if strings.HasPrefix(assetPath, prefix) {
				relativePath := strings.TrimPrefix(assetPath, prefix)
				fullPath := filepath.Join(bundleDir, relativePath)

				if _, err := os.Stat(fullPath); err != nil {
					http.NotFound(w, r)
					return
				}

				http.ServeFile(w, r, fullPath)
				return
			}
		}

		if globalRegistry.baseDir != "" {
			fullPath := filepath.Join(globalRegistry.baseDir, assetPath)
			if _, err := os.Stat(fullPath); err == nil {
				http.ServeFile(w, r, fullPath)
				return
			}
		}

		http.NotFound(w, r)
	})
}

func InstallAssets(destPublicDir string) error {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	for bundleName, bundleDir := range globalRegistry.paths {
		bundlePublicDir := filepath.Join(bundleDir, "public")
		if _, err := os.Stat(bundlePublicDir); os.IsNotExist(err) {
			bundlePublicDir = bundleDir
		}

		destDir := filepath.Join(destPublicDir, "assets", "bundles", bundleName)
		if err := os.MkdirAll(destDir, 0755); err != nil {
			return fmt.Errorf("failed to create destination for %s: %w", bundleName, err)
		}

		if err := copyDirRecursive(bundlePublicDir, destDir); err != nil {
			return fmt.Errorf("failed to install assets for %s: %w", bundleName, err)
		}
	}

	return nil
}

func copyDirRecursive(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := os.MkdirAll(dstPath, 0755); err != nil {
				return err
			}
			if err := copyDirRecursive(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			data, err := os.ReadFile(srcPath)
			if err != nil {
				return err
			}
			if err := os.WriteFile(dstPath, data, 0644); err != nil {
				return err
			}
		}
	}

	return nil
}

func AssetFunc() func(string) string {
	return func(path string) string {
		return ResolveAsset(path)
	}
}
