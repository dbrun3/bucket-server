package handler

import (
	"bucket-serve/cache"
	"bucket-serve/s3"
	"context"
	"errors"
	"mime"
	"net/http"
	"path/filepath"
	"slices"
	"strings"
)

type Handler struct {
	indexPath             string
	contentFileExt        string
	cacheFileExtInBrowser []string
	cache                 *cache.Cache
	s3                    *s3.Client
}

func NewHandler(cfg Config) *Handler {
	return &Handler{
		indexPath:             cfg.IndexPath,
		contentFileExt:        cfg.ContentFileExt,
		cacheFileExtInBrowser: cfg.CacheFileExtInBrowser,
		cache:                 cache.NewCache(cfg.CacheConfig),
		s3:                    s3.NewClient(),
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	host := r.Host
	path := r.URL.Path
	ctx := r.Context()

	page, err := h.getPage(ctx, host, path)
	if err != nil {
		if errors.Is(err, s3.ErrNoBucket) {
			http.Error(w, "Bucket not found", http.StatusForbidden)
		} else if errors.Is(err, s3.ErrNoKey) {
			http.Error(w, "Page Not Found", http.StatusNotFound)
		} else {
			http.Error(w, "Error while contacting s3", http.StatusInternalServerError)
		}
	}
	ext := filepath.Ext(path)
	contentType := mime.TypeByExtension(ext)
	w.Header().Set("Content-Type", contentType)
	if slices.Contains(h.cacheFileExtInBrowser, ext) {
		w.Header().Set("Cache-Control", "public, max-age=31536000")
	}
	w.Write(page)

}

func (h *Handler) getPage(ctx context.Context, host, path string) ([]byte, error) {

	page, state := h.cache.GetPage(host, path)
	ext := filepath.Ext(path)

	if ext == "" {
		path = strings.TrimRight(path, "/")
		path += h.contentFileExt
	}

	switch state {

	case cache.HIT:
		return page, nil

	case cache.STALE:
		go func() {
			updatedPage, err := h.s3.DownloadFile(context.Background(), host, path)
			if err != nil {
				h.cache.ClearPage(host, path)
				return
			}
			h.cache.UpdatePage(host, path, updatedPage)
		}()
		return page, nil

	default:
		newPage, err := h.s3.DownloadFile(ctx, host, path)
		if err != nil {
			if errors.Is(err, s3.ErrNoKey) &&
				ext == "" &&
				path != h.indexPath {
				// rewrite directory routes with index when true (used in SPA's) SPA assumed, todo: make configurable from bucket
				index, err := h.getPage(ctx, host, h.indexPath)
				if err == nil {
					go h.cache.CachePage(host, path, index)
				}
				return index, err
			}
			return nil, err
		}
		go h.cache.CachePage(host, path, newPage)
		return newPage, nil
	}
}
