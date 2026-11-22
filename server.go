package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
)

func runServer(cfg Config) {
	targets := flag.Args()
	if len(targets) == 0 {
		targets = []string{"."}
	}

	absTargets := make([]string, 0, len(targets))
	for _, t := range targets {
		abs, err := filepath.Abs(t)
		if err != nil {
			log.Fatal(err)
		}
		if _, err := os.Stat(abs); err != nil {
			log.Fatalf("not found: %s", t)
		}
		absTargets = append(absTargets, abs)
	}

	srv := NewServer(cfg, absTargets)
	go srv.Watch()

	url := fmt.Sprintf("http://localhost:%d/", cfg.Server.Port)
	log.Println(url)
	log.Printf("p5: %s (source=%s)", cfg.P5Version, cfg.P5Source)
	log.Println("press q or Ctrl+C to quit")

	if cfg.Server.OpenBrowser && !*noOpenFlag {
		if err := openBrowser(url, cfg.Server.Browser.Path, cfg.Server.Browser.Args); err != nil {
			log.Printf("failed to open browser: %v", err)
		}
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	go func() {
		buf := make([]byte, 1)
		for {
			if n, _ := os.Stdin.Read(buf); n > 0 && (buf[0] == 'q' || buf[0] == 3) {
				sig <- syscall.SIGINT
				return
			}
		}
	}()
	<-sig
	log.Printf("stopping server (PID %d) on port %d", os.Getpid(), cfg.Server.Port)
	srv.Close()
}

type Server struct {
	cfg         Config
	targets     []string
	clients     sync.Map
	http        *http.Server
	root        string
	localP5Path string
	p5URL       string
}

func NewServer(cfg Config, targets []string) *Server {
	root := targets[0]
	if info, err := os.Stat(root); err != nil || !info.IsDir() {
		root = filepath.Dir(root)
	}

	localP5 := cfg.P5LocalPath
	if _, err := os.Stat(localP5); err != nil {
		localP5 = ""
	}

	s := &Server{
		cfg:         cfg,
		targets:     targets,
		root:        root,
		localP5Path: localP5,
	}

	s.p5URL = s.resolveP5URL()

	addr := fmt.Sprintf("127.0.0.1:%d", cfg.Server.Port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("port %d is already in use", cfg.Server.Port)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/events", s.handleSSE)
	mux.HandleFunc("/__p5.js", s.handleLocalP5)
	mux.HandleFunc("/", s.handleRequest)
	s.http = &http.Server{Addr: addr, Handler: mux}

	go s.http.Serve(ln)
	return s
}

func (s *Server) resolveP5URL() string {
	source := strings.ToLower(strings.TrimSpace(s.cfg.P5Source))
	if source == "" {
		source = "auto"
	}

	cdn := fmt.Sprintf("https://cdn.jsdelivr.net/npm/p5@%s/lib/p5.min.js", s.cfg.P5Version)

	switch source {
	case "local":
		if s.localP5Path != "" {
			return "/__p5.js"
		}
		log.Println("p5Source=local but p5LocalPath is invalid or not set")
		return "/__p5.js"

	case "cdn":
		return cdn

	case "auto":
		fallthrough
	default:
		if source != "auto" {
			log.Printf("unknown p5Source %q, using auto", source)
		}
		if s.localP5Path == "" {
			return cdn
		}
		client := &http.Client{Timeout: 2 * time.Second}
		resp, err := client.Head(cdn)
		ok := err == nil && resp != nil && resp.StatusCode == 200
		if resp != nil {
			resp.Body.Close()
		}
		if ok {
			return cdn
		}
		log.Println("CDN unreachable, using local p5")
		return "/__p5.js"
	}
}

func (s *Server) Close() { s.http.Close() }

func (s *Server) handleLocalP5(w http.ResponseWriter, r *http.Request) {
	if s.localP5Path == "" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	http.ServeFile(w, r, s.localP5Path)
}

func (s *Server) handleSSE(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	id := fmt.Sprintf("%p", w)
	s.clients.Store(id, w)
	defer s.clients.Delete(id)
	fmt.Fprint(w, ": connected\n\n")
	flusher.Flush()
	<-r.Context().Done()
}

func (s *Server) broadcastReload() {
	s.clients.Range(func(_, v any) bool {
		w := v.(http.ResponseWriter)
		fmt.Fprint(w, "data: reload\n\n")
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		return true
	})
}

func (s *Server) handleRequest(w http.ResponseWriter, r *http.Request) {
	upath := r.URL.Path
	// ?raw が付いている場合はラッピングせず生のファイルを返す
	_, hasRaw := r.URL.Query()["raw"]

	if upath == "/" {
		for _, t := range s.targets {
			if strings.HasSuffix(t, ".js") {
				s.serveWrapped(w, t)
				return
			}
		}
	}

	if strings.HasSuffix(upath, ".js") && !hasRaw {
		// ラッピングして HTML を返す
		base := filepath.Base(upath)
		for _, t := range s.targets {
			if filepath.Base(t) == base {
				s.serveWrapped(w, t)
				return
			}
		}
	}

	filePath := filepath.Join(s.root, filepath.Clean(strings.TrimPrefix(upath, "/")))
	if !strings.HasPrefix(filepath.Clean(filePath), filepath.Clean(s.root)) {
		http.NotFound(w, r)
		return
	}

	if strings.HasSuffix(filePath, ".js") {
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
	}
	http.ServeFile(w, r, filePath)
}

func (s *Server) serveWrapped(w http.ResponseWriter, jsPath string) {
	// ?raw を付けることで handleRequest 側でラッピングを回避し、生の JS を返す
	// （<script> で読み込むときに再度 HTML でラップされるのを防ぐため）
	src := "/" + filepath.Base(jsPath) + "?raw&t=" + fmt.Sprint(time.Now().UnixNano())

	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<link rel="icon" href="data:,">
<style>
html, body {
  margin: 0;
  padding: 0;
  width: 100%%;
  height: 100%%;
  background: #030303;
  overflow: hidden;
}
body, main {
  display: flex;
  justify-content: center;
  align-items: center;
  width: 100%%;
  height: 100%%;
}
canvas {
  display: block;
  width: 100%% !important;
  height: 100%% !important;
  max-width: calc(100%% - 40px) !important;
  max-height: calc(100%% - 40px) !important;
  object-fit: contain;
}
</style>
<script>
Object.defineProperty(navigator, 'language', { get: () => 'en' });
Object.defineProperty(navigator, 'languages', { get: () => ['en'] });
</script>
<script src="%s"></script>
<script src="%s"></script>
</head>
<body>
<script>
const es = new EventSource('/events');
es.onmessage = () => location.reload();
</script>
</body>
</html>`, s.p5URL, src)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Write([]byte(html))
}

func (s *Server) Watch() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Printf("failed to create watcher: %v", err)
		return
	}
	defer watcher.Close()

	for _, t := range s.targets {
		if err := watcher.Add(t); err != nil {
			log.Printf("watcher add %s: %v", t, err)
			continue
		}
		if info, err := os.Stat(t); err == nil && info.IsDir() {
			filepath.Walk(t, func(p string, i os.FileInfo, e error) error {
				if e != nil {
					return nil
				}
				if i.IsDir() {
					watcher.Add(p)
				}
				return nil
			})
		}
	}

	var timer *time.Timer
	for {
		select {
		case ev, ok := <-watcher.Events:
			if !ok {
				return
			}
			if ev.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Rename|fsnotify.Remove) != 0 {
				if timer != nil {
					timer.Stop()
				}
				timer = time.AfterFunc(50*time.Millisecond, func() {
					log.Println("reload:", filepath.Base(ev.Name))
					s.broadcastReload()
				})

				if ev.Op&(fsnotify.Rename|fsnotify.Remove) != 0 {
					time.Sleep(10 * time.Millisecond)
					if err := watcher.Add(ev.Name); err != nil {
						// log.Printf("re-add watcher: %v", err)
					}
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Printf("watcher error: %v", err)
		}
	}
}

func openBrowser(rawURL, path string, args []string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return fmt.Errorf("invalid or unsupported URL: %s", rawURL)
	}

	var cmd *exec.Cmd
	if path != "" && fileExists(path) {
		fullArgs := make([]string, 0, len(args)+1)
		fullArgs = append(fullArgs, args...)
		fullArgs = append(fullArgs, rawURL)
		cmd = exec.Command(path, fullArgs...)
	} else {
		switch runtime.GOOS {
		case "windows":
			escapedURL := strings.ReplaceAll(rawURL, "&", "^&")
			cmd = exec.Command("cmd", "/c", "start", "", escapedURL)
		case "darwin":
			cmd = exec.Command("open", rawURL)
		default:
			cmd = exec.Command("xdg-open", rawURL)
		}
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start browser: %w", err)
	}

	go func() {
		_ = cmd.Wait()
	}()

	return nil
}
