package handler

import (
	"bucket-serve/cache"
	"bucket-serve/s3"
	"context"
	"errors"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type Handler struct {
	indexPath             string
	contentFileExt        string
	cacheFileExtInBrowser []string
	skipCacheExt          []string
	disableCache          bool
	cache                 *cache.Cache
	s3                    *s3.Client
}

func NewHandler(cfg Config) *Handler {
	return &Handler{
		indexPath:             cfg.IndexPath,
		contentFileExt:        cfg.ContentFileExt,
		cacheFileExtInBrowser: cfg.CacheFileExtInBrowser,
		skipCacheExt:          cfg.SkipCacheExt,
		disableCache:          cfg.CacheConfig.DisableCache,
		cache:                 cache.NewCache(cfg.CacheConfig),
		s3:                    s3.NewClient(),
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	host := r.Host
	path := r.URL.Path
	ctx := r.Context()

	// dev mode: override host with DEV_HOST env var
	if devHost := os.Getenv("DEV_HOST"); devHost != "" {
		host = devHost
	}

	page, etag, err := h.getPage(ctx, host, path)
	if err != nil {
		if errors.Is(err, s3.ErrNoBucket) {
			http.Error(w, "Bucket not found", http.StatusForbidden)
		} else if errors.Is(err, s3.ErrNoKey) {
			http.Error(w, "Page Not Found", http.StatusNotFound)
		} else {
			http.Error(w, "Error while contacting s3", http.StatusInternalServerError)
		}
		return
	}

	// 304 on matching etag
	if r.Header.Get("If-None-Match") == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	ext := filepath.Ext(path)
	contentType := mime.TypeByExtension(ext)
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("ETag", etag)
	if slices.Contains(h.cacheFileExtInBrowser, ext) {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	} else {
		w.Header().Set("Cache-Control", "no-cache")
	}
	w.Write(page)

}

func (h *Handler) getPage(ctx context.Context, host, path string) (page []byte, etag string, err error) {

	if filepath.Ext(path) == "" {
		path = strings.TrimRight(path, "/")
		path += h.contentFileExt
	}

	if h.disableCache || slices.Contains(h.skipCacheExt, filepath.Ext(path)) {
		return h.s3.DownloadFile(ctx, host, path)
	}

	var state cache.PageState
	page, etag, state = h.cache.GetPage(host, path)

	switch state {

	case cache.HIT:
		return

	case cache.STALE:
		go func() {
			updatedPage, updatedTag, err := h.s3.DownloadFile(context.Background(), host, path)
			if err != nil {
				h.cache.ClearPage(host, path)
				return
			}
			h.cache.UpdatePage(host, path, updatedPage, updatedTag)
		}()
		return

	default:
		page, etag, err = h.s3.DownloadFile(ctx, host, path)
		if err != nil {
			if errors.Is(err, s3.ErrNoKey) &&
				filepath.Ext(path) == h.contentFileExt &&
				path != h.indexPath {
				// rewrite missing content routes with index (used in SPA's) SPA assumed, todo: make configurable from bucket
				page, etag, err = h.getPage(ctx, host, h.indexPath)
				if err == nil {
					go h.cache.CachePage(host, path, page, etag)
				}
				return
			}
			return
		}
		go h.cache.CachePage(host, path, page, etag)
		return
	}
}
