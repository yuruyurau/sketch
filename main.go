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

	if *serverMode || flag.NArg() > 0 {
		runServer(cfg)
		return
	}
	runInteractive(cfg)
}

func runInteractive(cfg Config) {
	os.MkdirAll(cfg.SketchDir, 0o755)

	modes := `today : 今日のスケッチを開く
fuzzy : 内容のあいまい検索
exact : 内容の完全一致検索
file  : ファイル名検索
add   : スケッチ作成
record: 録画する
config: 設定ファイルを開く`

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
		path := filepath.Join(cfg.SketchDir, today+".js")
		ensureTemplate(path)
		startServerAndEdit(cfg, path, "")
	case "add":
		fmt.Print("新しいスケッチ名: ")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		name := strings.TrimSpace(scanner.Text())
		if name == "" {
			return
		}
		if !strings.HasSuffix(name, ".js") {
			name += ".js"
		}
		path := filepath.Join(cfg.SketchDir, name)
		ensureTemplate(path)
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
	case "config":
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatal(err)
		}
		configPath := filepath.Join(home, ".config", "sketch", "config.json")
		openEditor(cfg.Editor, configPath, "")
	default:
		log.Println("未対応のモードです")
	}
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
			"done        : 録画を開始",
			"quit        : キャンセル",
		}
		selected := fzfSelect(strings.Join(options, "\n"), []string{
			"--prompt", "録画オプション> ",
			"--layout", "reverse",
			"--style", "minimal",
			"--margin", "3,7",
			"--height", "22",
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
		"--preview-window", "hidden,border-none",
		"--bind", "ctrl-t:toggle-preview,J:preview-down,K:preview-up,ctrl-q:abort",
		"--cycle",
	}, sketchDir)
	if selected == "" {
		return ""
	}
	return filepath.Join(sketchDir, selected)
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
		"--preview-window", "hidden,border-none,+{2}+3/3",
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

func ensureTemplate(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(filepath.Dir(path), 0o755)
		os.WriteFile(path, []byte(sketchTemplate), 0o644)
	}
}

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

	openEditor(cfg.Editor, target, line)

	if cmd != nil && cmd.Process != nil {
		log.Printf("ポート %d は閉じられました。", cfg.Server.Port)
	}
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

const sketchTemplate = `p5.disableFriendlyErrors = true;

function setup() {
  createCanvas(400, 400);
}

function draw() {
  background(frameCount%256);
}`

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
