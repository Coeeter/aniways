package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"strings"
	"time"

	"github.com/coeeter/aniways/internal/app"
	"github.com/go-chi/chi/v5"
)

var (
	addr      = flag.String("addr", ":1234", "Address to listen on")
	userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3"
	logger    = app.NewLogger("PROXY")
	client    = &http.Client{
		Timeout: 15 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
			DisableKeepAlives: true,
		},
	}
)

func main() {
	flag.Parse()

	r := chi.NewRouter()
	r.Get("/proxy/{server}/{headers}/{pEnc}", proxyHandler)
	r.Options("/*", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Allow-Methods", "*")
		w.WriteHeader(http.StatusOK)
	})

	srv := &http.Server{
		Addr:    *addr,
		Handler: r,
	}

	errChan := make(chan error, 1)
	go func() {
		logger.Info("AniWays Proxy listening", "on", *addr)

		if err := srv.ListenAndServe(); err != nil {
			errChan <- err
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	select {
	case err := <-errChan:
		logger.Error("Error starting server", "err", err)
		os.Exit(1)
	case sig := <-stop:
		logger.Info("Shutting down...", "signal", sig)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			logger.Error("graceful shutdown failed", "err", err)
			os.Exit(1)
		}
		logger.Info("server stopped gracefully")
		os.Exit(0)
	}
}

func getHeadersFromRequest(r *http.Request) http.Header {
	headers := http.Header{
		"Accept": []string{"*/*"},
	}

	headersEnc := chi.URLParam(r, "headers")
	headersBytes, err := base64.StdEncoding.DecodeString(headersEnc)
	if err != nil {
		logger.Error("error decoding headers", "err", err, "headersEnc", headersEnc)
		return headers
	}

	headersMap := map[string]string{}
	if err := json.Unmarshal(headersBytes, &headersMap); err != nil {
		logger.Error("error unmarshaling headers", "err", err)
		return headers
	}

	for k, v := range headersMap {
		headers.Set(k, v)
	}

	return headers
}

func proxyHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	serverName := chi.URLParam(r, "server")
	pEnc := chi.URLParam(r, "pEnc")
	targetURLBytes, err := base64.URLEncoding.DecodeString(pEnc)
	if err != nil {
		http.Error(w, "invalid URL encoding", http.StatusBadRequest)
		logger.Error("error decoding target URL", "err", err, "pEnc", pEnc)
		return
	}
	targetURL, err := url.Parse(string(targetURLBytes))
	if err != nil || (targetURL.Scheme != "http" && targetURL.Scheme != "https") {
		http.Error(w, "invalid target URL", http.StatusBadRequest)
		logger.Error("error parsing target URL", "err", err, "targetURL", string(targetURLBytes))
		return
	}

	isHianime := strings.HasPrefix(strings.ToLower(serverName), "hd")

	headers := getHeadersFromRequest(r)
	if len(headers) == 1 {
		if isHianime {
			headers.Set("Referer", "https://megacloud.blog/")
			headers.Set("Origin", "https://megacloud.blog")
		} else {
			headers.Set("Referer", "https://megaplay.buzz/")
			headers.Set("Origin", "https://megaplay.buzz")
			headers.Set("User-Agent", userAgent)
		}
	}

	req, err := http.NewRequestWithContext(ctx, r.Method, targetURL.String(), nil)
	if err != nil {
		logger.Error("error creating request", "err", err)
		http.Error(w, "bad target URL", http.StatusBadRequest)
		return
	}

	req.Header = headers

	resp, err := client.Do(req)
	if err != nil {
		logger.Error("error fetching upstream", "err", err)
		http.Error(w, "upstream fetch failed", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Methods", "*")

	ext := path.Ext(targetURL.Path)

	if ext == ".m3u8" || ext == ".vtt" {
		w.Header().Del("Content-Length")
		w.Header().Set("Cache-Control", "public, max-age=60")
	} else {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	}

	setContentType(w, ext)

	w.WriteHeader(resp.StatusCode)

	headersEnc := chi.URLParam(r, "headers")
	encodeProxyURL := func(next string) string {
		full := next
		if !strings.HasPrefix(next, "http") {
			// relative in playlist
			base := targetURL.String()[:strings.LastIndex(targetURL.String(), "/")+1]
			full = base + next
		}
		pEnc := base64.URLEncoding.EncodeToString([]byte(full))
		return fmt.Sprintf("/proxy/%s/%s/%s", serverName, headersEnc, pEnc)
	}

	if ext == ".m3u8" || ext == ".vtt" {
		scanner := bufio.NewScanner(resp.Body)
		scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024) // up to 10MB lines
		flusher, _ := w.(http.Flusher)

		for scanner.Scan() {
			line := scanner.Text()
			out := line

			if !strings.HasPrefix(line, "#") {
				parts := strings.Split(line, "#")
				// For .vtt thumbnail contents
				if len(parts) > 1 {
					out = encodeProxyURL(parts[0]) + "#" + parts[1]
				} else {
					out = encodeProxyURL(line)
				}
			}

			io.WriteString(w, out+"\n")
			if flusher != nil {
				flusher.Flush()
			}
		}
		if err := scanner.Err(); err != nil {
			logger.Error("error scanning playlist", "err", err)
		}
	} else {
		// static content (ts, images, etc.)—just pipe directly
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		io.Copy(w, resp.Body)
	}

	logger.Info("proxied", "remoteAddr", r.RemoteAddr, "server", serverName, "targetURL", targetURL, "headers", headers.Clone())
}

func setContentType(w http.ResponseWriter, ext string) {
	switch ext {
	case ".m3u8":
		w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
	case ".ts":
		w.Header().Set("Content-Type", "video/MP2T")
	case ".vtt":
		w.Header().Set("Content-Type", "text/vtt")
	case ".png":
		w.Header().Set("Content-Type", "image/png")
	case ".jpg", ".jpeg":
		w.Header().Set("Content-Type", "image/jpeg")
	case ".webp":
		w.Header().Set("Content-Type", "image/webp")
	case ".ico":
		w.Header().Set("Content-Type", "image/x-icon")
	case ".html":
		w.Header().Set("Content-Type", "text/html")
	case ".js":
		w.Header().Set("Content-Type", "application/javascript")
	case ".css":
		w.Header().Set("Content-Type", "text/css")
	case ".txt":
		w.Header().Set("Content-Type", "text/plain")
	default:
		w.Header().Set("Content-Type", "application/octet-stream")
	}
}
