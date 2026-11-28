package main

import (
	"bufio"
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

func sheetInteractive(cfg Config) {
	if _, err := exec.LookPath("magick"); err != nil {
		log.Println("magick (ImageMagick) not found")
		return
	}

	targetMode := fzfSelect(`all      : すべて
one      : ひとつ
multi    : えらぶ
quit     : キャンセル`, []string{
		"--prompt", "対象> ",
		"--layout", "reverse",
		"--style", "minimal",
		"--height", "10",
		"--border",
	}, "")

	if targetMode == "" || strings.HasPrefix(targetMode, "quit") {
		return
	}

	var sketches []string
	switch {
	case strings.HasPrefix(targetMode, "one"):
		path := selectFile(cfg.SketchDir)
		if path == "" {
			return
		}
		sketches = []string{path}

	case strings.HasPrefix(targetMode, "multi"):
		sketches = selectFiles(cfg.SketchDir)
		if len(sketches) == 0 {
			return
		}

	default: // all
		sketches = listSketches(cfg.SketchDir)
		if len(sketches) == 0 {
			log.Println("no sketches found")
			return
		}
		log.Printf("%d sketches found (sketches in .ignore are excluded)", len(sketches))

		var buf strings.Builder
		for _, s := range sketches {
			rel, err := filepath.Rel(cfg.SketchDir, s)
			if err != nil {
				rel = s
			}
			buf.WriteString(filepath.ToSlash(rel) + "\n")
		}

		less := exec.Command("less", "-P", "Press 'q' to continue")
		less.Stdin = strings.NewReader(buf.String())
		less.Stdout = os.Stdout
		less.Stderr = os.Stderr
		if err := less.Run(); err != nil {
			fmt.Print(buf.String())
			log.Println("(install less for pager support)")
		}
	}

	fmt.Println("Opening options...")
	opts := selectSheetOptions(len(sketches), cfg.Record.PixelDensity)
	if opts == nil {
		return
	}

	frames := parseFrames(opts.frames)
	if len(frames) == 0 {
		log.Println("no targets")
		return
	}

	pd := opts.pixelDensity
	geometry := opts.geometry
	tile := opts.columns
	if tile == "" {
		cols := int(math.Ceil(math.Sqrt(float64(len(sketches) * len(frames)))))
		if cols < 1 {
			cols = 1
		}
		tile = fmt.Sprintf("%dx", cols)
	} else if !strings.Contains(tile, "x") {
		tile += "x"
	}

	tmpDir, err := os.MkdirTemp("", "sketch-sheet-")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	expected := len(sketches) * len(frames)
	var allPNGs []string
	var failedSketches int
	var captureErrors int

	for i, sketch := range sketches {
		rel, err := filepath.Rel(cfg.SketchDir, sketch)
		if err != nil {
			rel = filepath.Base(sketch)
		}
		rel = filepath.ToSlash(rel)

		fmt.Printf("[%d/%d] %s\n", i+1, len(sketches), rel)
		os.Stdout.Sync()

		pngs, err := captureSketchFrames(cfg, sketch, frames, tmpDir, pd)
		if err != nil {
			fmt.Printf("\r\033[K  error: %v\n", err)
			failedSketches++
			continue
		}

		if len(pngs) < len(frames) {
			captureErrors += len(frames) - len(pngs)
		}
		allPNGs = append(allPNGs, pngs...)

		fmt.Print("\r\033[K")
	}

	fmt.Println()
	if len(allPNGs) == 0 {
		log.Println("no images captured")
		return
	}

	fmt.Printf("captured: %d / %d images", len(allPNGs), expected)
	if failedSketches > 0 || captureErrors > 0 {
		fmt.Printf("  (sketches failed: %d, frames missed: %d)", failedSketches, captureErrors)
	}
	fmt.Println()

	if len(allPNGs) < expected {
		log.Println("some captures failed (contact sheet will be incomplete)")
	}

	sheetDir := filepath.Join(cfg.SketchDir, "sheet")
	os.MkdirAll(sheetDir, 0o755)
	out := filepath.Join(sheetDir, "sheet-"+time.Now().Format("20060102_150405")+".png")
	if err := makeContactSheet(allPNGs, out, tile, geometry, opts.background, opts.fill, opts.showLabel); err != nil {
		log.Printf("montage error: %v", err)
		return
	}
	log.Printf("終わり: %s", displayPath(out))
}

func listSketches(sketchDir string) []string {
	args := []string{
		"--type", "f",
		"--extension", "js",
		"--hidden",
		"--follow",
		"--exclude", ".git",
		"--no-require-git",
		".",
	}
	cmd := exec.Command("fd", args...)
	cmd.Dir = sketchDir
	out, err := cmd.Output()
	if err != nil {
		return nil
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var paths []string
	for _, line := range lines {
		if line != "" {
			paths = append(paths, filepath.Join(sketchDir, line))
		}
	}
	return paths
}

func parseFrames(s string) []int {
	s = strings.TrimSpace(s)
	if s == "" {
		return []int{0}
	}

	fields := strings.FieldsFunc(s, func(r rune) bool {
		return r == ',' || r == ' ' || r == '\t'
	})

	var frames []int
	seen := map[int]bool{}

	for _, part := range fields {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		nums := expandRange(part)
		for _, n := range nums {
			if n >= 0 && n <= 10000 && !seen[n] {
				frames = append(frames, n)
				seen[n] = true
			}
		}
	}

	if len(frames) == 0 {
		return []int{0}
	}
	sort.Ints(frames)
	return frames
}

// 30           → [30]
// 0..10        → [0 1 2 3 4 5 6 7 8 9]
// 1..5         → [1 2 3 4]
// 0..10..2     → [0 2 4 6 8]
// 10..0..-2    → [10 8 6 4 2]
// 0..100/5     → [0 20 40 60 80]
// 0..100/6     → [0 16 32 48 64 80]
func expandRange(s string) []int {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}

	if strings.Contains(s, "/") {
		parts := strings.SplitN(s, "/", 2)
		if len(parts) != 2 {
			return nil
		}
		rangePart := strings.TrimSpace(parts[0])
		n, err := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err != nil || n <= 0 {
			return nil
		}

		rp := strings.Split(rangePart, "..")
		if len(rp) != 2 {
			return nil
		}
		start, err1 := strconv.Atoi(strings.TrimSpace(rp[0]))
		end, err2 := strconv.Atoi(strings.TrimSpace(rp[1]))
		if err1 != nil || err2 != nil || start < 0 || end < 0 || start >= end {
			return nil
		}

		var result []int
		span := end - start
		for i := 0; i < n; i++ {
			v := start + (span*i)/n
			result = append(result, v)
		}
		return result
	}

	// 通常の範囲 or 単一値
	if !strings.Contains(s, "..") {
		if i, err := strconv.Atoi(s); err == nil && i >= 0 {
			return []int{i}
		}
		return nil
	}

	parts := strings.Split(s, "..")
	if len(parts) < 2 || len(parts) > 3 {
		return nil
	}

	start, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
	end, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err1 != nil || err2 != nil || start < 0 || end < 0 {
		return nil
	}

	step := 1
	if len(parts) == 3 {
		st, err := strconv.Atoi(strings.TrimSpace(parts[2]))
		if err != nil || st == 0 {
			return nil
		}
		step = st
	}

	var result []int
	if step > 0 {
		for i := start; i < end; i += step {
			result = append(result, i)
		}
	} else {
		for i := start; i > end; i += step {
			result = append(result, i)
		}
	}
	return result
}

func makeContactSheet(pngs []string, outPath, tile, geometry, background, fill string, showLabel bool) error {
	sx, sy := extractSpacing(geometry)
	border := sx
	if sx != sy {
		border = sx + "x" + sy
	}

	montageArgs := []string{
		"montage",
		"-background", background,
		"-geometry", geometry,
		"-tile", tile,
	}

	if showLabel {
		montageArgs = append(montageArgs,
			"-pointsize", "11",
			"-fill", fill,
			"-label", "%f",
		)
	}

	montageArgs = append(montageArgs, pngs...)
	montageArgs = append(montageArgs, "miff:-")

	borderArgs := []string{
		"-",
		"-bordercolor", background,
		"-border", border,
		outPath,
	}

	cmdMontage := exec.Command("magick", montageArgs...)
	cmdBorder := exec.Command("magick", borderArgs...)

	pipe, err := cmdMontage.StdoutPipe()
	if err != nil {
		return err
	}
	cmdBorder.Stdin = pipe
	cmdBorder.Stdout = os.Stdout
	cmdBorder.Stderr = os.Stderr
	cmdMontage.Stderr = os.Stderr

	if err := cmdBorder.Start(); err != nil {
		return err
	}
	if err := cmdMontage.Start(); err != nil {
		return err
	}
	if err := cmdMontage.Wait(); err != nil {
		return err
	}
	return cmdBorder.Wait()
}

func extractSpacing(geometry string) (x, y string) {
	x, y = "0", "0"

	parts := strings.Split(geometry, "+")
	if len(parts) < 2 {
		return
	}

	var nums []string
	for i := len(parts) - 1; i >= 0 && len(nums) < 2; i-- {
		s := strings.TrimSpace(parts[i])
		if s == "" {
			continue
		}
		if _, err := strconv.Atoi(s); err == nil {
			nums = append([]string{s}, nums...)
		}
	}

	if len(nums) >= 1 {
		x = nums[0]
	}
	if len(nums) >= 2 {
		y = nums[1]
	} else if len(nums) == 1 {
		y = nums[0]
	}
	return
}

func captureSketchFrames(cfg Config, sketchPath string, frames []int, outDir string, pixelDensity float64) ([]string, error) {
	if len(frames) > 0 {
		fmt.Printf("\r1/%d", len(frames))
		os.Stdout.Sync()
	}

	rec := cfg.Record
	if rec.Port == 0 {
		return nil, fmt.Errorf("record.port is not set")
	}

	args := []string{
		"--server", "--no-open",
		"--port", fmt.Sprint(rec.Port),
	}
	if cfg.P5Version != "" {
		args = append(args, "--p5-version", cfg.P5Version)
	}
	if cfg.P5Source != "" {
		args = append(args, "--p5-source", cfg.P5Source)
	}
	args = append(args, sketchPath)

	cmd := exec.Command(os.Args[0], args...)
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	defer func() {
		if cmd.Process != nil {
			cmd.Process.Kill()
			<-done
		}
	}()

	select {
	case <-done:
		return nil, fmt.Errorf("server failed to start")
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
			return nil, fmt.Errorf("server exited")
		default:
		}
		time.Sleep(250 * time.Millisecond)
	}
	if !ready {
		return nil, fmt.Errorf("server not ready")
	}

	chrome := rec.Browser.Path
	if !fileExists(chrome) {
		return nil, fmt.Errorf("chrome not found: %s", chrome)
	}

	ctx, cancel := context.WithCancel(context.Background())
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

	taskCtx, cancelTask := chromedp.NewContext(allocCtx)
	defer cancelTask()

	if err := chromedp.Do(taskCtx, chromedp.Navigate(url)); err != nil {
		return nil, fmt.Errorf("navigate: %w", err)
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
		return nil, fmt.Errorf("canvas not found")
	}

	if err := chromedp.Do(taskCtx, chromedp.Sleep(300*time.Millisecond)); err != nil {
		return nil, err
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
    `, pixelDensity)))
	if err != nil || size.Width == 0 || size.Height == 0 {
		return nil, fmt.Errorf("canvas size")
	}

	if err := chromedp.Do(taskCtx, chromedp.EmulateViewport(int64(size.Width), int64(size.Height))); err != nil {
		return nil, fmt.Errorf("viewport: %w", err)
	}

	baseName := strings.TrimSuffix(filepath.Base(sketchPath), filepath.Ext(sketchPath))
	if strings.HasSuffix(baseName, ".mod") {
		baseName = strings.TrimSuffix(baseName, ".mod")
	}

	var captured []string
	current := -1

	for idx, target := range frames {
		fmt.Printf("\r%d/%d", idx+1, len(frames))
		os.Stdout.Sync()

		for current < target {
			if err := chromedp.Do(taskCtx, chromedp.Evaluate[chromedp.Void](`
                (() => {
                    const p = window.p5Instance || window;
                    if (typeof p.redraw === "function") p.redraw();
                })()
            `)); err != nil {
				return captured, err
			}
			if err := chromedp.Do(taskCtx, chromedp.Evaluate[chromedp.Void](`
                new Promise(r => requestAnimationFrame(() => requestAnimationFrame(r)))
            `, chromedp.EvalAwaitPromise)); err != nil {
				return captured, err
			}
			current++
		}

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
			continue
		}

		b64 := strings.TrimPrefix(dataURL, "data:image/png;base64,")
		png, err := base64.StdEncoding.DecodeString(b64)
		if err != nil || len(png) < 100 {
			continue
		}

		name := fmt.Sprintf("%s_f%d.png", baseName, target)
		path := filepath.Join(outDir, name)
		if err := os.WriteFile(path, png, 0o644); err != nil {
			continue
		}
		captured = append(captured, path)
	}

	return captured, nil
}

type sheetOpts struct {
	frames       string
	pixelDensity float64
	geometry     string
	columns      string
	showLabel    bool
	background   string
	fill         string
}

func selectSheetOptions(sketchCount int, defaultPD float64) *sheetOpts {
	opts := &sheetOpts{
		frames:       "0",
		pixelDensity: defaultPD,
		geometry:     "200x200+6+6",
		columns:      "",
		showLabel:    true,
		background:   "white",
		fill:         "black",
	}

	for {
		frames := parseFrames(opts.frames)
		total := sketchCount * len(frames)

		var framesPreview string
		switch {
		case len(frames) == 0:
			framesPreview = "none"
		case len(frames) <= 8:
			framesPreview = strings.Trim(strings.Join(strings.Fields(fmt.Sprint(frames)), ", "), "[]")
		default:
			framesPreview = fmt.Sprintf("%d, %d, %d, ..., %d", frames[0], frames[1], frames[2], frames[len(frames)-1])
		}
		framesPreview = fmt.Sprintf("%s", framesPreview)

		var cols int
		if opts.columns == "" {
			cols = int(math.Ceil(math.Sqrt(float64(total))))
		} else {
			clean := strings.TrimSuffix(opts.columns, "x")
			if n, err := strconv.Atoi(clean); err == nil && n > 0 {
				cols = n
			} else {
				cols = int(math.Ceil(math.Sqrt(float64(total))))
			}
		}
		if cols < 1 {
			cols = 1
		}
		rows := int(math.Ceil(float64(total) / float64(cols)))
		if rows < 1 {
			rows = 1
		}
		layout := fmt.Sprintf("%dx%d", cols, rows)

		columnsDisplay := opts.columns
		if columnsDisplay == "" {
			columnsDisplay = fmt.Sprintf("%d", cols)
		}

		labelDisplay := "on"
		if !opts.showLabel {
			labelDisplay = "off"
		}

		// infoLines := []string{
		// 	fmt.Sprintf("target → %d sketches × %d frames = %d images", sketchCount, len(frames), total),
		// 	fmt.Sprintf("frames → %s", framesPreview),
		// 	fmt.Sprintf("layout → %s", layout),
		// 	"",
		// }
		infoLines := []string{
			fmt.Sprintf("total  → %d", total),
			fmt.Sprintf("frames → %d (%s)", len(frames), framesPreview),
			fmt.Sprintf("layout → %s", layout),
			"",
		}

		menuLines := []string{
			fmt.Sprintf("frames      : %s", opts.frames),
			fmt.Sprintf("pixelDensity: %g", opts.pixelDensity),
			fmt.Sprintf("geometry    : %s", opts.geometry),
			fmt.Sprintf("columns     : %s", columnsDisplay),
			fmt.Sprintf("label       : %s", labelDisplay),
			fmt.Sprintf("background  : %s", opts.background),
			fmt.Sprintf("fill        : %s", opts.fill),
			"help        : framesの書き方",
			"done        : 始める",
			"quit        : キャンセル",
		}

		input := strings.Join(append(infoLines, menuLines...), "\n")

		selected := fzfSelect(input, []string{
			"--prompt", "sheet> ",
			"--layout", "reverse",
			"--style", "minimal",
			"--margin", "1,7",
			"--height", "20",
			"--border",
			"--header-lines", strconv.Itoa(len(infoLines)),
			"--color", "header:-1",
		}, "")

		if selected == "" || strings.HasPrefix(selected, "quit") {
			return nil
		}

		key := strings.TrimSpace(strings.SplitN(selected, ":", 2)[0])
		key = strings.TrimSpace(strings.SplitN(key, "→", 2)[0])

		if strings.Contains(selected, "→") || key == "" || key == "target" || key == "layout" {
			continue
		}

		switch key {
		case "frames":
			if v := prompt(fmt.Sprintf("frames (current: %s): ", opts.frames)); v != "" {
				if isValidFrames(v) {
					opts.frames = v
				} else {
					fmt.Println("invalid frames syntax (e.g. 0, 1..10, 0..100/5)")
				}
			}
		case "pixelDensity":
			if v := prompt(fmt.Sprintf("pixelDensity (current: %g): ", opts.pixelDensity)); v != "" {
				f, err := strconv.ParseFloat(v, 64)
				if err != nil || f <= 0 || f > 4 {
					fmt.Println("pixelDensity must be > 0 and <= 4")
				} else {
					opts.pixelDensity = f
				}
			}
		case "geometry":
			if v := prompt(fmt.Sprintf("geometry (current: %s): ", opts.geometry)); v != "" {
				if isValidGeometry(v) {
					opts.geometry = v
				} else {
					fmt.Println("invalid geometry (e.g. 200x200+6+6)")
				}
			}
		case "columns":
			if v := prompt("columns: "); v != "" {
				clean := strings.TrimSuffix(strings.TrimSpace(v), "x")
				if n, err := strconv.Atoi(clean); err == nil && n > 0 {
					opts.columns = clean
				} else {
					fmt.Println("invalid columns (positive integer only)")
				}
			}
		case "label":
			opts.showLabel = !opts.showLabel
		case "background":
			if v := prompt("background (e.g. white, #111, gray): "); v != "" {
				if isValidColor(v) {
					opts.background = v
				} else {
					fmt.Println("invalid color")
				}
			}
		case "fill":
			if v := prompt("fill (label color, e.g. black, #eee): "); v != "" {
				if isValidColor(v) {
					opts.fill = v
				} else {
					fmt.Println("invalid color")
				}
			}
		case "help":
			fmt.Println()
			fmt.Println("frames syntax:")
			fmt.Println("  30              → frame 30")
			fmt.Println("  1 5 10          → frames 1, 5, 10")
			fmt.Println("  1..5            → frames 1,2,3,4,5")
			fmt.Println("  0..10..2        → frames 0,2,4,6,8,10")
			fmt.Println("  0..100/5        → 5 equal parts of 0..100")
			fmt.Println()
			fmt.Println("range = start..<end..step")
			fmt.Println("/n   = n equal parts")
			fmt.Println()
			fmt.Print("Press Enter to return...")
			bufio.NewReader(os.Stdin).ReadString('\n')
		case "done":
			return opts
		}
	}
}

func isValidFrames(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return true
	}

	fields := strings.FieldsFunc(s, func(r rune) bool {
		return r == ',' || r == ' ' || r == '\t'
	})

	if len(fields) == 0 {
		return false
	}

	for _, part := range fields {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if expandRange(part) == nil {
			return false
		}
	}
	return true
}

func isValidGeometry(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}

	parts := strings.Split(s, "+")
	base := parts[0]
	if !strings.Contains(base, "x") {
		return false
	}

	dims := strings.Split(base, "x")
	if len(dims) != 2 {
		return false
	}
	for _, d := range dims {
		n, err := strconv.Atoi(strings.TrimSpace(d))
		if err != nil || n <= 0 {
			return false
		}
	}

	for _, p := range parts[1:] {
		n, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil || n < 0 {
			return false
		}
	}
	return true
}

func isValidColor(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	if strings.HasPrefix(s, "#") {
		hex := s[1:]
		if len(hex) != 3 && len(hex) != 6 {
			return false
		}
		for _, c := range hex {
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
				return false
			}
		}
		return true
	}
	for _, c := range s {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_') {
			return false
		}
	}
	return true
}
