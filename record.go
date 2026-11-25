package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

func recordInteractive(cfg Config, sketchPath string, rec RecordConfig, p5Version, p5Source string) {
	home, _ := os.UserHomeDir()
	pretty := func(p string) string {
		if home != "" && strings.HasPrefix(p, home) {
			return "~" + strings.ReplaceAll(p[len(home):], "\\", "/")
		}
		return strings.ReplaceAll(p, "\\", "/")
	}

	log.Println("Recording")
	log.Printf("  %s", pretty(sketchPath))

	if rec.Port == 0 {
		log.Fatal("record.port is not set in config.json")
	}

	args := []string{
		"--server", "--no-open",
		"--port", fmt.Sprint(rec.Port),
	}
	if p5Version != "" {
		args = append(args, "--p5-version", p5Version)
	}
	if p5Source != "" {
		args = append(args, "--p5-source", p5Source)
	}
	args = append(args, sketchPath)

	cmd := exec.Command(os.Args[0], args...)
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}
	defer func() {
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
	}()

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-done:
		log.Fatalf("recording server failed to start on port %d (port may already be in use)", rec.Port)
	case <-time.After(300 * time.Millisecond):
	}

	url := fmt.Sprintf("http://localhost:%d/", rec.Port)

	ready := false
	for i := 0; i < 40; i++ {
		resp, err := http.Get(url)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == 200 {
				ready = true
				break
			}
		}
		select {
		case <-done:
			log.Fatalf("recording server exited unexpectedly on port %d", rec.Port)
		default:
		}
		time.Sleep(250 * time.Millisecond)
	}
	if !ready {
		log.Fatalf("recording server on port %d did not become ready", rec.Port)
	}

	os.MkdirAll(cfg.OutDir, 0o755)
	ext := "webm"
	if strings.ToLower(rec.Codec) != "vp9" {
		ext = "mp4"
	}
	out := filepath.Join(cfg.OutDir, time.Now().Format("20060102_150405")+"."+ext)

	log.Printf("  → %s", pretty(out))

	if err := recordSketch(url, rec, out); err != nil {
		log.Printf("error: %v", err)
	} else {
		log.Printf("終わり: %s", pretty(out))
	}
}

func recordSketch(url string, rec RecordConfig, outPath string) error {
	chrome := rec.Browser.Path
	if !fileExists(chrome) {
		return fmt.Errorf("chrome not found: %s", chrome)
	}

	var ctx context.Context
	var cancel context.CancelFunc
	if rec.Timeout > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), time.Duration(rec.Timeout)*time.Minute)
	} else {
		ctx, cancel = context.WithCancel(context.Background())
	}
	defer cancel()

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ExecPath(chrome),
		chromedp.Flag("headless", "new"),
		chromedp.Flag("hide-scrollbars", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("force-device-scale-factor", "1"),
		chromedp.WindowSize(400, 400),
	)

	allocCtx, cancelAlloc := chromedp.NewExecAllocator(ctx, opts...)
	defer cancelAlloc()

	taskCtx, cancelTask := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancelTask()

	if err := chromedp.Do(taskCtx, chromedp.Navigate(url)); err != nil {
		return fmt.Errorf("navigate: %w", err)
	}

	var hasCanvas bool
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		exists, err := chromedp.Run(taskCtx, chromedp.Evaluate[bool](`!!document.querySelector("canvas")`))
		if err == nil && exists {
			hasCanvas = true
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	if !hasCanvas {
		return fmt.Errorf("canvas not found")
	}

	if err := chromedp.Do(taskCtx, chromedp.Sleep(300*time.Millisecond)); err != nil {
		return err
	}

	type Size struct {
		Width  int `json:"width"`
		Height int `json:"height"`
	}
	size, err := chromedp.Run(taskCtx, chromedp.Evaluate[Size](fmt.Sprintf(`
        (() => {
            const p = window.p5Instance || window;
            if (typeof p.pixelDensity === "function") p.pixelDensity(%v);
            if (typeof p.noLoop === "function") p.noLoop();
            const c = document.querySelector("canvas");
            if (!c) return {width:0, height:0};
            c.style.width = "auto";
            c.style.height = "auto";
            c.style.maxWidth = "none";
            c.style.maxHeight = "none";
            return {width: c.width, height: c.height};
        })()
    `, rec.PixelDensity)))
	if err != nil {
		return fmt.Errorf("size: %w", err)
	}
	if size.Width == 0 || size.Height == 0 {
		return fmt.Errorf("canvas size zero")
	}

	log.Printf("  %d frames @ %gfps", rec.Total, rec.FPS)
	log.Printf("  %s | CRF %d", strings.ToUpper(rec.Codec), rec.CRF)
	log.Printf("  %dx%d (pixelDensity %g)", size.Width, size.Height, rec.PixelDensity)
	if rec.Timeout > 0 {
		log.Printf("  timeout: %d min", rec.Timeout)
	} else {
		log.Printf("  timeout: none")
	}
	log.Printf("  maxFail: %d", rec.MaxFail)

	if err := chromedp.Do(taskCtx, chromedp.EmulateViewport(int64(size.Width), int64(size.Height))); err != nil {
		return fmt.Errorf("emulate viewport: %w", err)
	}

	if rec.Start > 0 {
		fmt.Printf("  advancing 0/%d frames...", rec.Start)
		for i := 0; i < rec.Start; i++ {
			if err := chromedp.Do(taskCtx, chromedp.Evaluate[chromedp.Void](`
                (() => {
                    const p = window.p5Instance || window;
                    if (typeof p.redraw === "function") p.redraw();
                })()
            `)); err != nil {
				return fmt.Errorf("warmup redraw: %w", err)
			}

			if err := chromedp.Do(taskCtx, chromedp.Evaluate[chromedp.Void](`
                new Promise(r => requestAnimationFrame(() => requestAnimationFrame(r)))
            `, chromedp.EvalAwaitPromise)); err != nil {
				return fmt.Errorf("warmup raf: %w", err)
			}

			if (i+1)%30 == 0 || i+1 == rec.Start {
				fmt.Printf("\r  advancing %d/%d frames...", i+1, rec.Start)
			}
		}
		fmt.Println()
	}

	ff := exec.Command("ffmpeg", buildFfmpegArgs(rec, outPath)...)
	stdin, err := ff.StdinPipe()
	if err != nil {
		return err
	}

	var ffmpegOut bytes.Buffer
	ff.Stdout = &ffmpegOut
	ff.Stderr = &ffmpegOut

	if err := ff.Start(); err != nil {
		return fmt.Errorf("start ffmpeg: %w", err)
	}
	defer stdin.Close()

	failCount := 0
	maxFail := rec.MaxFail
	if maxFail <= 0 {
		maxFail = 10
	}

	for i := 0; i < rec.Total; i++ {
		// redraw
		if err := chromedp.Do(taskCtx, chromedp.Evaluate[chromedp.Void](`
            (() => {
                const p = window.p5Instance || window;
                if (typeof p.redraw === "function") p.redraw();
            })()
        `)); err != nil {
			if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
				return fmt.Errorf("timeout or canceled: %w", err)
			}
			failCount++
			if failCount >= maxFail {
				return fmt.Errorf("too many consecutive frame failures (%d): %w", maxFail, err)
			}
			continue
		}

		// raf
		if err := chromedp.Do(taskCtx, chromedp.Evaluate[chromedp.Void](`
            new Promise(r => requestAnimationFrame(() => requestAnimationFrame(r)))
        `, chromedp.EvalAwaitPromise)); err != nil {
			if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
				return fmt.Errorf("timeout or canceled: %w", err)
			}
			failCount++
			if failCount >= maxFail {
				return fmt.Errorf("too many consecutive frame failures (%d): %w", maxFail, err)
			}
			continue
		}

		// dataURL
		dataURL, err := chromedp.Run(taskCtx, chromedp.Evaluate[string](`
            (async () => {
                const canvas = document.querySelector("canvas");
                if (!canvas) return "";
                const blob = await new Promise(res => canvas.toBlob(res, "image/png"));
                if (!blob) return "";
                return await new Promise(resolve => {
                    const reader = new FileReader();
                    reader.onloadend = () => resolve(reader.result);
                    reader.readAsDataURL(blob);
                });
            })()
        `, chromedp.EvalAwaitPromise))
		if err != nil || !strings.HasPrefix(dataURL, "data:image/png;base64,") {
			if err != nil && (errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled)) {
				return fmt.Errorf("timeout or canceled: %w", err)
			}
			failCount++
			if failCount >= maxFail {
				return fmt.Errorf("too many consecutive frame failures (%d): %w", maxFail, err)
			}
			continue
		}

		// decode
		b64 := strings.TrimPrefix(dataURL, "data:image/png;base64,")
		png, err := base64.StdEncoding.DecodeString(b64)
		if err != nil || len(png) < 100 {
			failCount++
			if failCount >= maxFail {
				return fmt.Errorf("too many consecutive frame failures (%d): %w", maxFail, err)
			}
			continue
		}

		failCount = 0

		if _, err := stdin.Write(png); err != nil {
			return fmt.Errorf("write frame: %w", err)
		}

		if i%30 == 0 || i == rec.Total-1 {
			log.Printf("Frame %d/%d", i+1, rec.Total)
		}
	}

	stdin.Close()

	err = ff.Wait()
	if err != nil {
		log.Printf("ffmpeg output:\n%s", ffmpegOut.String())
		return fmt.Errorf("ffmpeg: %w", err)
	}

	return nil
}

func buildFfmpegArgs(r RecordConfig, out string) []string {
	args := []string{
		"-y", "-hide_banner", "-loglevel", "error", "-stats",
		"-f", "image2pipe", "-framerate", fmt.Sprintf("%g", r.FPS),
		"-i", "pipe:0",
	}
	if strings.ToLower(r.Codec) == "vp9" {
		args = append(args,
			"-c:v", "libvpx-vp9",
			"-crf", fmt.Sprint(r.CRF),
			"-b:v", "0",
			"-deadline", "good",
			"-g", fmt.Sprint(r.G),
			"-cpu-used", r.CPUUsed,
			"-row-mt", "1",
			"-pix_fmt", "yuv420p",
		)
	} else {
		args = append(args,
			"-c:v", "libx264",
			"-preset", r.Preset,
			"-crf", fmt.Sprint(r.CRF),
			"-pix_fmt", "yuv420p",
			"-movflags", "+faststart",
		)
	}
	return append(args, out)
}
