package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"log"
	"math"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var (
	serverMode    = flag.Bool("server", false, "run as server")
	noOpenFlag    = flag.Bool("no-open", false, "do not open browser")
	portFlag      = flag.Int("port", 0, "override server port")
	p5VersionFlag = flag.String("p5-version", "", "override p5Version")
	p5SourceFlag  = flag.String("p5-source", "", "override p5Source")
)

func main() {
	log.SetFlags(0)
	flag.Parse()
	cfg := loadConfig()
	if *portFlag > 0 {
		cfg.Server.Port = *portFlag
	}
	if *p5VersionFlag != "" {
		cfg.P5Version = *p5VersionFlag
	}
	if *p5SourceFlag != "" {
		cfg.P5Source = *p5SourceFlag
	}

	if flag.NArg() > 0 && flag.Arg(0) == "build" {
		runBuild(cfg, flag.Args()[1:])
		return
	}

	if *serverMode || flag.NArg() > 0 {
		runServer(cfg)
		return
	}
	runInteractive(cfg)
}

func runInteractive(cfg Config) {
	os.MkdirAll(cfg.SketchDir, 0o755)

	modes := `today     : 今日のスケッチを開く
fuzzy     : 内容のあいまい検索
exact     : 内容の完全一致検索
file      : ファイル名検索
add       : スケッチ作成
record    : 録画する
build     : ビルドする
sheet     : コンタクトシートを作る
config    : 設定ファイルを開く`

	selected := fzfSelect(modes, []string{
		"--prompt", "モード> ",
		"--layout", "reverse",
		"--style", "minimal",
		"--margin", "3,7",
		"--height", "19",
		"--border",
	}, "")
	if selected == "" {
		return
	}
	mode := strings.TrimSpace(strings.Split(selected, ":")[0])

	switch mode {
	case "today":
		today := time.Now().Format("2006-01-02")
		jsPath := filepath.Join(cfg.SketchDir, today+".js")
		modPath := filepath.Join(cfg.SketchDir, today+".mod.js")

		jsExists := fileExists(jsPath)
		modExists := fileExists(modPath)

		var path string
		var isMod bool

		switch {
		case jsExists && !modExists:
			path = jsPath
			isMod = false
		case modExists && !jsExists:
			path = modPath
			isMod = true
		case jsExists && modExists:
			sel := fzfSelect(`global   : `+today+`.js
instance : `+today+`.mod.js`, []string{
				"--prompt", "どちらを開く？> ",
				"--layout", "reverse",
				"--style", "minimal",
				"--height", "10",
				"--border",
			}, "")
			if sel == "" {
				return
			}
			if strings.Contains(sel, "instance") {
				path = modPath
				isMod = true
			} else {
				path = jsPath
				isMod = false
			}
		default:
			sel := fzfSelect(`global   : グローバルモード (.js)
instance : インスタンスモード (.mod.js)`, []string{
				"--prompt", "新規作成> ",
				"--layout", "reverse",
				"--style", "minimal",
				"--height", "10",
				"--border",
			}, "")
			if sel == "" {
				return
			}
			isMod = strings.HasPrefix(strings.TrimSpace(sel), "instance")
			if isMod {
				path = modPath
			} else {
				path = jsPath
			}
			ensureTemplate(path, isMod)
		}

		startServerAndEdit(cfg, path, "")

	case "add":
		fmt.Print("新しいスケッチ名: ")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		name := strings.TrimSpace(scanner.Text())
		if name == "" {
			return
		}

		var isMod *bool
		base := name
		lower := strings.ToLower(name)

		switch {
		case strings.HasSuffix(lower, ".mod.js"):
			b := true
			isMod = &b
			base = name[:len(name)-len(".mod.js")]
		case strings.HasSuffix(lower, ".js"):
			b := false
			isMod = &b
			base = name[:len(name)-len(".js")]
		}

		if strings.TrimSpace(base) == "" {
			fmt.Println("無効な名前です")
			return
		}

		if isMod == nil {
			sel := fzfSelect(`global   : グローバルモード (.js)
instance : インスタンスモード (.mod.js)`, []string{
				"--prompt", "モード> ",
				"--layout", "reverse",
				"--style", "minimal",
				"--height", "10",
				"--border",
			}, "")
			if sel == "" {
				return
			}
			mod := strings.HasPrefix(strings.TrimSpace(sel), "instance")
			isMod = &mod
		}

		ext := ".js"
		if *isMod {
			ext = ".mod.js"
		}

		path := filepath.Join(cfg.SketchDir, base+ext)
		ensureTemplate(path, *isMod)
		startServerAndEdit(cfg, path, "")

	case "file":
		if path := selectFile(cfg.SketchDir); path != "" {
			startServerAndEdit(cfg, path, "")
		}
	case "fuzzy", "exact":
		if path, line := selectByContent(cfg.SketchDir, mode == "exact"); path != "" {
			startServerAndEdit(cfg, path, line)
		}
	// case "server":
	// 	startServerOnly(cfg.SketchDir)
	case "record":
		path := selectFile(cfg.SketchDir)
		if path == "" {
			return
		}
		rec, p5Version, p5Source := selectRecordOptions(cfg.Record, cfg.P5Version, cfg.P5Source)
		if rec == nil {
			return
		}
		recordInteractive(cfg, path, *rec, p5Version, p5Source)
	case "build":
		path := selectFile(cfg.SketchDir)
		if path == "" {
			return
		}
		runBuildInteractive(cfg, path)
	case "sheet":
		sheetInteractive(cfg)
	case "config":
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatal(err)
		}
		configPath := filepath.Join(home, ".config", "sketch", "config.json")
		openEditor(resolveEditor(cfg.Editor), configPath, "")
	default:
		log.Println("未対応のモードです")
	}
}

func runBuildInteractive(cfg Config, jsPath string) {
	options := `separate : 分離版（index.html + bundle.js）
single   : 単一ファイル版（index.html のみ）
quit     : キャンセル`

	selected := fzfSelect(options, []string{
		"--prompt", "ビルド形式> ",
		"--layout", "reverse",
		"--style", "minimal",
		"--margin", "3,7",
		"--height", "15",
		"--border",
	}, "")
	if selected == "" {
		return
	}

	key := strings.TrimSpace(strings.Split(selected, ":")[0])
	if key == "quit" {
		return
	}

	single := key == "single"

	defaultOut := filepath.Join(cfg.SketchDir, "dist")
	defaultOutDisplay := displayPath(defaultOut)

	outdir := prompt(fmt.Sprintf("出力ディレクトリ (空enterで %s): ", defaultOutDisplay))
	if outdir == "" {
		outdir = defaultOut
	} else {
		outdir = expandPath(outdir)
	}

	args := []string{}
	if single {
		args = append(args, "--single")
	}
	args = append(args, "--outdir", outdir)
	args = append(args, jsPath)

	runBuild(cfg, args)
}

func displayPath(path string) string {
	home, err := os.UserHomeDir()
	if err == nil && strings.HasPrefix(path, home) {
		path = "~" + path[len(home):]
	}
	return filepath.ToSlash(path)
}

func selectRecordOptions(base RecordConfig, p5Version, p5Source string) (*RecordConfig, string, string) {
	rec := base
	for {
		options := []string{
			fmt.Sprintf("fps         : %g", rec.FPS),
			fmt.Sprintf("total       : %d", rec.Total),
			fmt.Sprintf("start       : %d", rec.Start),
			fmt.Sprintf("codec       : %s", rec.Codec),
			fmt.Sprintf("crf         : %d", rec.CRF),
			fmt.Sprintf("cpuUsed     : %s", rec.CPUUsed),
			fmt.Sprintf("g           : %d", rec.G),
			fmt.Sprintf("pixelDensity: %g", rec.PixelDensity),
			fmt.Sprintf("p5Version   : %s", p5Version),
			fmt.Sprintf("p5Source    : %s", p5Source),
			fmt.Sprintf("timeout     : %d min (0=none)", rec.Timeout),
			fmt.Sprintf("maxFail     : %d", rec.MaxFail),
			"help        : 使い方",
			"done        : 録画を開始",
			"quit        : キャンセル",
		}
		selected := fzfSelect(strings.Join(options, "\n"), []string{
			"--prompt", "録画オプション> ",
			"--layout", "reverse",
			"--style", "minimal",
			"--margin", "3,7",
			"--height", "25",
			"--border",
		}, "")
		if selected == "" {
			return nil, "", ""
		}
		key := strings.TrimSpace(strings.Split(selected, ":")[0])
		switch key {
		case "fps":
			if v := prompt(fmt.Sprintf("fps (現在: %g): ", rec.FPS)); v != "" {
				if f, err := strconv.ParseFloat(v, 64); err == nil {
					rec.FPS = f
				}
			}
		case "total":
			if v := prompt(fmt.Sprintf("total (現在: %d): ", rec.Total)); v != "" {
				if i, err := evalIntExpr(v); err == nil {
					rec.Total = i
				} else {
					fmt.Printf("  invalid: %v\n", err)
				}
			}
		case "start":
			if v := prompt(fmt.Sprintf("start (現在: %d): ", rec.Start)); v != "" {
				if i, err := strconv.Atoi(v); err == nil && i >= 0 {
					rec.Start = i
				} else {
					fmt.Println("  invalid: must be an integer >= 0")
				}
			}
		case "codec":
			if v := prompt(fmt.Sprintf("codec (現在: %s / vp9 or h264): ", rec.Codec)); v != "" {
				rec.Codec = v
			}
		case "crf":
			if v := prompt(fmt.Sprintf("crf (現在: %d): ", rec.CRF)); v != "" {
				if i, err := strconv.Atoi(v); err == nil {
					rec.CRF = i
				}
			}
		case "cpuUsed":
			if v := prompt(fmt.Sprintf("cpuUsed (現在: %s): ", rec.CPUUsed)); v != "" {
				rec.CPUUsed = v
			}
		case "g":
			if v := prompt(fmt.Sprintf("g (現在: %d): ", rec.G)); v != "" {
				if i, err := strconv.Atoi(v); err == nil {
					rec.G = i
				}
			}
		case "pixelDensity":
			if v := prompt(fmt.Sprintf("pixelDensity (現在: %g): ", rec.PixelDensity)); v != "" {
				if f, err := strconv.ParseFloat(v, 64); err == nil {
					rec.PixelDensity = f
				}
			}
		case "p5Version":
			if v := prompt(fmt.Sprintf("p5Version (現在: %s): ", p5Version)); v != "" {
				p5Version = v
			}
		case "p5Source":
			if v := prompt(fmt.Sprintf("p5Source (現在: %s / cdn, local, auto): ", p5Source)); v != "" {
				v = strings.ToLower(strings.TrimSpace(v))
				switch v {
				case "cdn", "local", "auto":
					p5Source = v
				default:
					fmt.Println("  → cdn / local / auto のいずれかを指定してください")
				}
			}
		case "timeout":
			if v := prompt(fmt.Sprintf("timeout minutes (現在: %d, 0=none): ", rec.Timeout)); v != "" {
				if i, err := strconv.Atoi(v); err == nil {
					rec.Timeout = i
				}
			}
		case "maxFail":
			if v := prompt(fmt.Sprintf("maxFail (現在: %d): ", rec.MaxFail)); v != "" {
				if i, err := strconv.Atoi(v); err == nil && i > 0 {
					rec.MaxFail = i
				}
			}
		case "help":
			fmt.Println()
			fmt.Println("record options:")
			fmt.Println()
			fmt.Println("  fps           frames per second")
			fmt.Println("  total         frames to record (number or formula like 60*30+140)")
			fmt.Println("  start         starting frame (frames to skip from 0)")
			fmt.Println("  codec         vp9 (webm) / h264 (mp4)")
			fmt.Println("  crf           15-35 (vp9) / 18-28 (h264)")
			fmt.Println("  cpuUsed       vp9 speed/quality tradeoff (0-5)")
			fmt.Println("  g             keyframe interval")
			fmt.Println("  pixelDensity  resolution multiplier")
			fmt.Println("  p5Version     p5.js version")
			fmt.Println("  p5Source      p5.js source (cdn / local / auto)")
			fmt.Println("  timeout       timeout in minutes (0 = none)")
			fmt.Println("  maxFail       max consecutive failures")
			fmt.Println()
			fmt.Print("Press Enter to return...")
			bufio.NewReader(os.Stdin).ReadString('\n')
		case "done":
			return &rec, p5Version, p5Source
		case "quit":
			return nil, "", ""
		}
	}
}

func prompt(msg string) string {
	fmt.Print(msg)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	return strings.TrimSpace(scanner.Text())
}

func fzfSelect(input string, extraArgs []string, dir string) string {
	cmd := exec.Command("fzf", extraArgs...)
	if dir != "" {
		cmd.Dir = dir
	}
	cmd.Stdin = strings.NewReader(input)
	cmd.Stderr = os.Stderr

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return ""
	}

	if err := cmd.Start(); err != nil {
		return ""
	}

	out, err := io.ReadAll(stdout)
	cmd.Wait()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func selectFile(sketchDir string) string {
	args := []string{
		"--type", "f",
		"--extension", "js",
		"--hidden",
		"--follow",
		"--exclude", ".git",
		"--no-require-git",
	}
	args = append(args, ".")
	fd := exec.Command("fd", args...)
	fd.Dir = sketchDir
	out, err := fd.Output()
	if err != nil {
		return ""
	}
	selected := fzfSelect(string(out), []string{
		"--layout", "reverse",
		"--style", "minimal",
		"--margin", "4,7",
		"--preview", "bat --style=plain --color=always --wrap=never {}",
		"--preview-window", "hidden,border,right,60%",
		"--bind", "ctrl-t:toggle-preview,J:preview-down,K:preview-up,ctrl-q:abort",
		"--cycle",
	}, sketchDir)
	if selected == "" {
		return ""
	}
	return filepath.Join(sketchDir, selected)
}

func selectFiles(sketchDir string) []string {
	args := []string{
		"--type", "f",
		"--extension", "js",
		"--hidden",
		"--follow",
		"--exclude", ".git",
		"--no-require-git",
		".",
	}
	fd := exec.Command("fd", args...)
	fd.Dir = sketchDir
	out, err := fd.Output()
	if err != nil {
		return nil
	}

	selected := fzfSelect(string(out), []string{
		"--multi",
		"--layout", "reverse",
		"--style", "minimal",
		"--margin", "4,7",
		"--preview", "bat --style=plain --color=always --wrap=never {}",
		"--preview-window", "hidden,border,right,60%",
		"--bind", "ctrl-t:toggle-preview,J:preview-down,K:preview-up,ctrl-q:abort,ctrl-a:select-all,ctrl-d:deselect-all",
		"--cycle",
		"--prompt", "select (Tab/Shift-Tab)> ",
	}, sketchDir)

	if selected == "" {
		return nil
	}

	var paths []string
	for _, line := range strings.Split(selected, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			paths = append(paths, filepath.Join(sketchDir, line))
		}
	}
	return paths
}

func selectByContent(sketchDir string, exact bool) (string, string) {
	args := []string{
		"--line-number",
		"--no-heading",
		"--color=never",
		"--smart-case",
		"--hidden",
		"--follow",
		"--no-require-git",
		"--glob", "!.git/**",
		"", // 全行を出す
		".",
	}

	rg := exec.Command("rg", args...)
	rg.Dir = sketchDir
	out, err := rg.Output()
	if err != nil {
		return "", ""
	}

	fzfArgs := []string{
		"--delimiter", ":",
		"--nth", "3..",
		"--style", "minimal",
		"--layout", "reverse",
		"--margin", "4,7",
		"--preview", "bat --style=plain --color=always --wrap=never --highlight-line {2} {1}",
		"--preview-window", "hidden,border,right,60%,+{2}+3/3",
		"--bind", "ctrl-t:toggle-preview,J:preview-down,K:preview-up,ctrl-q:abort",
		"--cycle",
	}

	if exact {
		fzfArgs = append(fzfArgs, "--exact")
	}

	selected := fzfSelect(string(out), fzfArgs, sketchDir)

	if selected == "" {
		return "", ""
	}
	parts := strings.SplitN(selected, ":", 3)
	if len(parts) < 2 {
		return "", ""
	}
	return filepath.Join(sketchDir, parts[0]), parts[1]
}

func ensureTemplate(path string, isMod bool) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(filepath.Dir(path), 0o755)
		tmpl := sketchTemplate
		if isMod {
			tmpl = sketchModTemplate
		}
		os.WriteFile(path, []byte(tmpl), 0o644)
	}
}

const sketchTemplate = `p5.disableFriendlyErrors = true;

function setup() {
  createCanvas(400, 400);
}

function draw() {
  background(frameCount % 256);
}`

const sketchModTemplate = `const sketch = (p) => {
  p.setup = () => {
    p.createCanvas(400, 400);
  };

  p.draw = () => {
    p.background(p.frameCount % 256);
  };
};

window.p5Instance = new p5(sketch);`

func startServerAndEdit(cfg Config, target, line string) {
	var cmd *exec.Cmd
	if isPortInUse(cfg.Server.Port) {
		log.Printf("port %d is already in use", cfg.Server.Port)
		fmt.Printf("server not started. open in %s anyway? [y/N]: ", cfg.Editor)
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		answer := strings.ToLower(strings.TrimSpace(scanner.Text()))
		if answer != "y" && answer != "yes" {
			return
		}
	} else {
		cmd = exec.Command(os.Args[0], "--server", target)
		if err := cmd.Start(); err != nil {
			log.Fatal(err)
		}
		defer func() {
			if cmd.Process != nil {
				cmd.Process.Kill()
			}
		}()
	}

	openEditor(resolveEditor(cfg.Editor), target, line)

	if cmd != nil && cmd.Process != nil {
		log.Printf("ポート %d は閉じられました。", cfg.Server.Port)
	}
}

func resolveEditor(cfgEditor string) string {
	termProgram := os.Getenv("TERM_PROGRAM")

	if os.Getenv("ZED_TERM") == "true" || termProgram == "zed" {
		switch cfgEditor {
		case "zed", "zeditor":
			return cfgEditor
		default:
			return "zed"
		}
	}

	if termProgram == "vscode" {
		switch cfgEditor {
		case "code", "code-insiders":
			return cfgEditor
		default:
			return "code"
		}
	}

	if cfgEditor == "" {
		return "nvim"
	}
	return cfgEditor
}

func openEditor(editor, target, line string) {
	if editor == "" {
		editor = "nvim"
	}

	var args []string
	switch editor {
	case "nvim", "vim", "vi", "neovide", "nano", "fresh", "micro", "edit":
		if line != "" {
			args = append(args, "+"+line)
		}
		args = append(args, target)
	case "hx", "helix":
		if line != "" {
			args = append(args, target+":"+line)
		} else {
			args = append(args, target)
		}
	case "zed", "zeditor":
		args = append(args, "--wait")
		if line != "" {
			args = append(args, target+":"+line)
		} else {
			args = append(args, target)
		}
	case "code", "code-insiders":
		args = append(args, "--wait")
		if line != "" {
			args = append(args, "-g", target+":"+line)
		} else {
			args = append(args, target)
		}
	default:
		if line != "" {
			args = append(args, "+"+line)
		}
		args = append(args, target)
	}

	editorCmd := exec.Command(editor, args...)
	editorCmd.Stdin, editorCmd.Stdout, editorCmd.Stderr = os.Stdin, os.Stdout, os.Stderr

	if err := editorCmd.Run(); err != nil {
		log.Printf("failed to start editor %q: %v", editor, err)
	}
}

// func startServerOnly(dir string) {
// 	cmd := exec.Command(os.Args[0], "--server", dir)
// 	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
// 	cmd.Run()
// }

func isPortInUse(port int) bool {
	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return true
	}
	ln.Close()
	return false
}

type BrowserConfig struct {
	Path string   `json:"path"`
	Args []string `json:"args"`
}

type ServerConfig struct {
	Port        int           `json:"port"`
	OpenBrowser bool          `json:"openBrowser"`
	Browser     BrowserConfig `json:"browser"`
}

type RecordConfig struct {
	Port         int           `json:"port"`
	FPS          float64       `json:"fps"`
	Total        int           `json:"total"`
	Start        int           `json:"start"`
	CRF          int           `json:"crf"`
	Codec        string        `json:"codec"`
	Preset       string        `json:"preset"`
	CPUUsed      string        `json:"cpuUsed"`
	G            int           `json:"g"`
	PixelDensity float64       `json:"pixelDensity"`
	Timeout      int           `json:"timeout"` // minutes. 0 = no timeout
	MaxFail      int           `json:"maxFail"` // consecutive failures
	Browser      BrowserConfig `json:"browser"`
}

type Config struct {
	Server      ServerConfig `json:"server"`
	Record      RecordConfig `json:"record"`
	SketchDir   string       `json:"sketchDir"`
	OutDir      string       `json:"outDir"`
	Editor      string       `json:"editor"`
	P5Version   string       `json:"p5Version"`
	P5Source    string       `json:"p5Source"`
	P5LocalPath string       `json:"p5LocalPath"`
}

func expandPath(p string) string {
	if strings.HasPrefix(p, "~") {
		if home, err := os.UserHomeDir(); err == nil {
			rest := strings.TrimPrefix(p, "~")
			rest = strings.TrimPrefix(rest, "/")
			rest = strings.TrimPrefix(rest, "\\")
			p = filepath.Join(home, rest)
		}
	}
	return os.ExpandEnv(p)
}

func loadConfig() Config {
	var cfg Config

	cfg.Record.Timeout = 60
	cfg.Record.MaxFail = 10

	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("failed to get home directory")
	}
	configPath := filepath.Join(home, ".config", "sketch", "config.json")

	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalf("config.json not found: %s", prettyPath(configPath))
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		log.Fatal(err)
	}

	if cfg.Record.MaxFail <= 0 {
		cfg.Record.MaxFail = 10
	}

	if cfg.Editor == "" {
		cfg.Editor = "nvim"
	}

	cfg.SketchDir = expandPath(cfg.SketchDir)
	cfg.OutDir = expandPath(cfg.OutDir)
	if cfg.OutDir != "" && !filepath.IsAbs(cfg.OutDir) {
		cfg.OutDir = filepath.Join(cfg.SketchDir, cfg.OutDir)
	}
	if cfg.OutDir == "" {
		cfg.OutDir = filepath.Join(cfg.SketchDir, "record")
	}
	cfg.P5LocalPath = expandPath(cfg.P5LocalPath)
	cfg.Server.Browser.Path = expandPath(cfg.Server.Browser.Path)
	cfg.Record.Browser.Path = expandPath(cfg.Record.Browser.Path)

	// log.Println("今日は")
	// log.Printf("loaded config: %s", prettyPath(configPath))
	return cfg
}

func prettyPath(p string) string {
	home, _ := os.UserHomeDir()
	if home != "" && strings.HasPrefix(p, home) {
		return "~" + strings.ReplaceAll(p[len(home):], "\\", "/")
	}
	return strings.ReplaceAll(p, "\\", "/")
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func evalIntExpr(expr string) (int, error) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return 0, fmt.Errorf("empty")
	}

	node, err := parser.ParseExpr(expr)
	if err != nil {
		return 0, err
	}

	val, err := evalAST(node)
	if err != nil {
		return 0, err
	}

	if val < 1 || val > math.MaxInt {
		return 0, fmt.Errorf("out of range (must be >= 1)")
	}
	return val, nil
}

func evalAST(node ast.Expr) (int, error) {
	switch n := node.(type) {
	case *ast.BasicLit:
		switch n.Kind {
		case token.INT:
			v, err := strconv.ParseInt(n.Value, 0, 64)
			if err != nil {
				return 0, err
			}
			return int(v), nil

		case token.FLOAT:
			fv, err := strconv.ParseFloat(n.Value, 64)
			if err != nil {
				return 0, err
			}
			if math.IsNaN(fv) || math.IsInf(fv, 0) {
				return 0, fmt.Errorf("invalid number")
			}
			// 整数値のみ通す（1e4, 1.0, .0 など）
			if fv != math.Trunc(fv) {
				return 0, fmt.Errorf("not an integer value")
			}
			if fv < 1 || fv > math.MaxInt {
				return 0, fmt.Errorf("out of range (must be >= 1)")
			}
			return int(fv), nil

		default:
			return 0, fmt.Errorf("not an integer value")
		}

	case *ast.ParenExpr:
		return evalAST(n.X)

	case *ast.UnaryExpr:
		x, err := evalAST(n.X)
		if err != nil {
			return 0, err
		}
		switch n.Op {
		case token.ADD:
			return x, nil
		case token.SUB:
			return -x, nil
		default:
			return 0, fmt.Errorf("unsupported unary operator: %s", n.Op)
		}

	case *ast.BinaryExpr:
		x, err := evalAST(n.X)
		if err != nil {
			return 0, err
		}
		y, err := evalAST(n.Y)
		if err != nil {
			return 0, err
		}
		switch n.Op {
		case token.ADD:
			return x + y, nil
		case token.SUB:
			return x - y, nil
		case token.MUL:
			return x * y, nil
		case token.QUO:
			if y == 0 {
				return 0, fmt.Errorf("division by zero")
			}
			return x / y, nil
		case token.REM:
			if y == 0 {
				return 0, fmt.Errorf("division by zero")
			}
			return x % y, nil
		default:
			return 0, fmt.Errorf("unsupported binary operator: %s", n.Op)
		}

	default:
		return 0, fmt.Errorf("unsupported expression")
	}
}

func runBuild(cfg Config, args []string) {
	fs := flag.NewFlagSet("build", flag.ExitOnError)
	outdir := fs.String("outdir", "dist", "output directory for build")
	single := fs.Bool("single", false, "output single-file HTML only (index.html)")

	if err := fs.Parse(args); err != nil {
		log.Fatal(err)
	}

	remaining := fs.Args()
	if len(remaining) == 0 {
		log.Fatal("usage: sketch build [options] <file.js>")
	}

	jsPath := remaining[0]
	abs, err := filepath.Abs(jsPath)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := os.Stat(abs); err != nil {
		log.Fatalf("not found: %s", jsPath)
	}

	if err := os.MkdirAll(*outdir, 0o755); err != nil {
		log.Fatal(err)
	}

	if _, err := exec.LookPath("bun"); err != nil {
		log.Fatal("bun not found in PATH")
	}

	isModule := strings.HasSuffix(filepath.Base(abs), ".mod.js")

	format := "iife"
	if isModule {
		format = "esm"
	}

	bundlePath := filepath.Join(*outdir, "bundle.js")
	cmd := exec.Command("bun", "build",
		"--bundle",
		"--minify",
		"--target=browser",
		"--format="+format,
		abs,
		"--outfile", bundlePath,
	)
	cmd.Dir = filepath.Dir(abs)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	log.Printf("bundling %s...", filepath.Base(abs))
	if err := cmd.Run(); err != nil {
		log.Fatalf("bun build failed: %v", err)
	}

	p5Src, err := resolveP5ForBuild(cfg, *outdir)
	if err != nil {
		log.Fatal(err)
	}

	if *single {
		bundleContent, err := os.ReadFile(bundlePath)
		if err != nil {
			log.Fatal(err)
		}
		html := generateStaticHTMLContentSingle(p5Src, string(bundleContent), isModule)
		htmlPath := filepath.Join(*outdir, "index.html")
		if err := os.WriteFile(htmlPath, []byte(html), 0o644); err != nil {
			log.Fatal(err)
		}
		os.Remove(bundlePath)
		log.Printf("done → %s (single file)", displayPath(htmlPath))
	} else {
		html := generateStaticHTMLContentSeparate(p5Src, "bundle.js", isModule)
		htmlPath := filepath.Join(*outdir, "index.html")
		if err := os.WriteFile(htmlPath, []byte(html), 0o644); err != nil {
			log.Fatal(err)
		}
		log.Printf("done → %s/ (index.html + bundle.js)", displayPath(*outdir))
	}
}

func generateStaticHTMLContentSeparate(p5Src, bundleSrc string, isModule bool) string {
	scriptTag := `<script src="./%s"></script>`
	if isModule {
		scriptTag = `<script type="module" src="./%s"></script>`
	}

	return fmt.Sprintf(`<!DOCTYPE html>
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
%s
</head>
<body>
</body>
</html>`, p5Src, fmt.Sprintf(scriptTag, bundleSrc))
}

func generateStaticHTMLContentSingle(p5Src, inlineJS string, isModule bool) string {
	scriptTag := `<script>%s</script>`
	if isModule {
		scriptTag = `<script type="module">%s</script>`
	}

	return fmt.Sprintf(`<!DOCTYPE html>
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
%s
</head>
<body>
</body>
</html>`, p5Src, fmt.Sprintf(scriptTag, inlineJS))
}

func resolveP5ForBuild(cfg Config, outdir string) (string, error) {
	source := strings.ToLower(strings.TrimSpace(cfg.P5Source))
	if source == "" {
		source = "auto"
	}

	cdn := fmt.Sprintf("https://cdn.jsdelivr.net/npm/p5@%s/lib/p5.min.js", cfg.P5Version)

	useLocal := false
	localPath := cfg.P5LocalPath
	if localPath != "" {
		localPath = expandPath(localPath)
		if _, err := os.Stat(localPath); err == nil {
			useLocal = true
		}
	}

	switch source {
	case "local":
		if !useLocal {
			return "", fmt.Errorf("p5Source=local but p5LocalPath is invalid or not found: %s", cfg.P5LocalPath)
		}
	case "cdn":
		return cdn, nil
	case "auto":
		if !useLocal {
			return cdn, nil
		}
	default:
		return cdn, nil
	}

	dst := filepath.Join(outdir, "p5.min.js")
	if err := copyFile(localPath, dst); err != nil {
		return "", fmt.Errorf("failed to copy p5.min.js: %w", err)
	}
	return "./p5.min.js", nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}
