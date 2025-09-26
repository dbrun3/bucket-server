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
	IsSPA                 bool
	contentFileExt        string
	cacheFileExtInBrowser []string
	cache                 *cache.Cache
	s3                    *s3.Client
}

func NewHandler(cfg Config) *Handler {
	return &Handler{
		indexPath:             cfg.IndexPath,
		IsSPA:                 cfg.IsSPA,
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

	// append ext (.html) to paths if missing (non-SPA)
	if !h.IsSPA && ext == "" {
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
				h.IsSPA &&
				ext == "" &&
				path != h.indexPath {
				// redirect directory routes to index when true (used in SPA's)
				return h.getPage(ctx, host, h.indexPath)
			}
			return nil, err
		}
		go func() {
			h.cache.CachePage(host, path, newPage)
		}()
		return newPage, nil
	}
}
