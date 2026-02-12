//go:build mage

package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
)

var (
	dbName  = "test.db"
	dbPath  = "file:./" + dbName
	tmpDir  = "./tmp"
	browser = getenv("BROWSER", "vivaldi")

	isMac = runtime.GOOS == "darwin"
	isWSL = detectWSL()

	bgProcs []*exec.Cmd
)

const PERM = 0755

// ---------------- Command Abstraction ----------------

type CommandType int

const (
	CmdExec CommandType = iota
	CmdGoBuild
	CmdGoRun
	CmdTmpExec
)

type Command struct {
	Type CommandType
	Cmd  string
	Out  string
	Tags []string
	Args []string
	Desc string
}

func run(c Command) error {
	var cmd *exec.Cmd

	switch c.Type {
	case CmdGoBuild:
		args := []string{"build"}
		if len(c.Tags) > 0 {
			args = append(args, "-tags="+strings.Join(c.Tags, " "))
		}
		if c.Out != "" {
			args = append(args, "-o", c.Out)
		}
		args = append(args, c.Args...)
		cmd = exec.Command("go", args...)

	case CmdGoRun:
		args := []string{"run"}
		if len(c.Tags) > 0 {
			args = append(args, "-tags="+strings.Join(c.Tags, " "))
		}
		args = append(args, c.Args...)
		cmd = exec.Command("go", args...)

	case CmdTmpExec:
		bin := filepath.Join(tmpDir, c.Cmd)
		cmd = exec.Command(bin, c.Args...)

	case CmdExec:
		cmd = exec.Command(c.Cmd, c.Args...)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if c.Out != "" {
		cmd.Env = append(os.Environ(), "OUT="+c.Out)
	}
	return cmd.Run()
}

// ---------------- Background Process Management ----------------

func init() {
	setupSignalHandler()
}

func runBackground(cmd *exec.Cmd) error {
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return err
	}
	bgProcs = append(bgProcs, cmd)
	return nil
}

func cleanupBackground() {
	for _, p := range bgProcs {
		if p.Process != nil {
			fmt.Printf("Killing background process: %s (pid %d)\n", p.Path, p.Process.Pid)
			_ = p.Process.Kill()
		}
	}
	bgProcs = nil
}

func setupSignalHandler() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for range c {
			cleanupBackground()
			os.Exit(1)
		}
	}()
}

// ---------------- Helpers ----------------

func checkMigrations() error {
	cmd := exec.Command(filepath.Join(tmpDir, "goose"), "status")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to check migration status: %w", err)
	}
	if strings.Contains(string(output), "Pending") {
		return fmt.Errorf("there are pending migrations. Run 'mage DBUp' to apply them")
	}
	return nil
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func detectWSL() bool {
	if runtime.GOOS != "linux" {
		return false
	}
	data, err := os.ReadFile("/proc/version")
	return err == nil && strings.Contains(strings.ToLower(string(data)), "microsoft")
}

func openBrowser(url string) {
	if isWSL {
		exec.Command("cmd.exe", "/c", "start", browser, url).Run()
	} else if isMac {
		exec.Command("open", "-a", browser, url).Run()
	}
}

func checkProc(name string, port int) bool {
	out, err := exec.Command("pgrep", "-f", name).CombinedOutput()
	if err != nil || len(out) == 0 {
		return false
	}
	out, err = exec.Command("lsof", fmt.Sprintf("-i:%d", port)).CombinedOutput()
	return err == nil && len(out) > 0
}

func detectDistro() string {
	data, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return ""
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "ID=") {
			return strings.Trim(strings.SplitN(line, "=", 2)[1], "\"")
		}
	}
	return ""
}

// ---------------- Mage Targets ----------------

func Build() error {
	if err := GenAll(); err != nil {
		return err
	}
	return run(Command{
		Type: CmdGoBuild,
		Out:  filepath.Join(tmpDir, "main"),
		Tags: []string{"fts5", "staticfs", "imageprocessing"},
	})
}

func BuildWeb() error {
	return run(Command{
		Type: CmdGoBuild,
		Out:  filepath.Join(tmpDir, "web"),
		Tags: []string{"fts5", "staticfs"},
	})
}

func serve(
	airpath string,
	app string,
) error {
	if err := checkMigrations(); err != nil {
		return err
	}
	if err := GenAll(); err != nil {
		return err
	}
	openBrowser("http://localhost:7331/cchoice")

	templCmd := exec.Command("go", "tool", "templ", "generate",
		"--watch", "--proxy=http://localhost:2626", "--open-browser=false")
	if err := runBackground(templCmd); err != nil {
		return err
	}

	airCmd := exec.Command("go", "tool", "air", "-c", airpath, app)
	airCmd.Stdout = os.Stdout
	airCmd.Stderr = os.Stderr
	return airCmd.Run()
}

func Serve() error {
	return serve(".air.api.toml", "api")
}

func ServeWeb() error {
	return serve(".air.web.toml", "web")
}

func BuildGoose() error {
	if err := run(Command{
		Type: CmdExec,
		Cmd:  "git",
		Args: []string{"submodule", "update", "--init", "--recursive"},
	}); err != nil {
		return err
	}

	const pathGoose = "./cmd/goose"
	if err := os.Chdir(pathGoose); err != nil {
		return err
	}
	if err := run(Command{Type: CmdExec, Cmd: "go", Args: []string{"mod", "tidy"}}); err != nil {
		return err
	}
	if err := run(Command{
		Type: CmdGoBuild,
		Out:  "../../" + filepath.Join(tmpDir, "goose"),
		Tags: []string{
			"no_postgres",
			"no_mysql",
			"no_clickhouse",
			"no_mssql",
			"no_vertica",
			"no_ydb",
		},
		Args: []string{pathGoose},
	}); err != nil {
		return err
	}
	return os.Chmod("../../"+filepath.Join(tmpDir, "goose"), PERM)
}

func Setup() error {
	const pathPreCommitHook = "./git/hooks/pre-commit"
	if _, err := os.Stat(pathPreCommitHook); os.IsNotExist(err) {
		if err := run(Command{Type: CmdExec, Cmd: "cp", Args: []string{"./scripts/pre-commit-unit-test.sh", pathPreCommitHook}}); err != nil {
			return err
		}
		if err := os.Chmod(pathPreCommitHook, PERM); err != nil {
			return err
		}
	}

	const pathEnv = "./.env"
	if _, err := os.Stat(pathEnv); os.IsNotExist(err) {
		return run(Command{Type: CmdExec, Cmd: "cp", Args: []string{"./.env.sample", pathEnv}})
	}

	const pathLogs = "./logs"
	if _, err := os.Stat(pathLogs); os.IsNotExist(err) {
		return run(Command{Type: CmdExec, Cmd: "mkdir", Args: []string{pathLogs}})
	}
	return nil
}

func SetupProd() error {
	cmdInstall := exec.Command("sudo", "apt", "install", "rclone")
	if err := cmdInstall.Run(); err != nil {
		return err
	}
	fmt.Println("Run `rclone config`")

	const pathScript = "/usr/local/bin/dbbackup.sh"
	if err := run(Command{
		Type: CmdExec,
		Cmd:  "cp",
		Args: []string{"./scripts/dbbackup.sh", pathScript},
	}); err != nil {
		return err
	}
	if err := os.Chmod(pathScript, PERM); err != nil {
		return err
	}

	cronJob := "0 3 * * * /usr/local/bin/dbbackup.sh >> /var/log/backup.log 2>&1"
	checkCmd := exec.Command("bash", "-c", fmt.Sprintf("crontab -l 2>/dev/null | grep -F '%s'", cronJob))
	if err := checkCmd.Run(); err != nil {
		addCronCmd := exec.Command(
			"bash",
			"-c",
			fmt.Sprintf(
				"(crontab -l 2>/dev/null; echo '%s') | crontab -",
				cronJob,
			),
		)
		if err := addCronCmd.Run(); err != nil {
			return err
		}
		fmt.Println("Cron job added successfully")
	}

	return nil
}

func GenBrandLogos() error {
	if err := run(Command{
		Type: CmdTmpExec,
		Cmd:  "genimages",
		Args: []string{
			"convert_images", "--inpath=./cmd/web/static/images/brand_logos",
			"--outpath=./cmd/web/static/images/brand_logos", "--format=webp"},
	}); err != nil {
		return err
	}
	return nil
}

func GenPaymentLogos() error {
	if err := run(Command{
		Type: CmdTmpExec,
		Cmd:  "genimages",
		Args: []string{
			"convert_images", "--inpath=./cmd/web/static/images/payments",
			"--outpath=./cmd/web/static/images/payments", "--format=webp"},
	}); err != nil {
		return err
	}
	return nil
}

func GenImages() error {
	if err := GenBrandLogos(); err != nil {
		return err
	}
	if err := GenPaymentLogos(); err != nil {
		return err
	}
	if err := run(Command{
		Type: CmdGoBuild,
		Out:  filepath.Join(tmpDir, "genimages"),
		Tags: []string{"imageprocessing", "staticfs"},
	}); err != nil {
		return err
	}
	if err := run(Command{
		Type: CmdTmpExec,
		Cmd:  "genimages",
		Args: []string{
			"prepare_image_variants",
			"--inpath=./cmd/web/static/images/product_images/bosch",
			"--outpath=./cmd/web/static/images/product_images",
		},
	}); err != nil {
		return err
	}
	return nil
}

func MigrateImagesToLinodeStorage() error {
	if err := run(Command{
		Type: CmdGoBuild,
		Out:  filepath.Join(tmpDir, "migrate_images_linode"),
		Tags: []string{"staticfs"},
	}); err != nil {
		return err
	}
	if err := run(Command{
		Type: CmdTmpExec,
		Cmd:  "migrate_images_linode",
		Args: []string{
			"migrate_images_linode",
			"--dry-run=false",
			"--bucket=PRIVATE",
		},
	}); err != nil {
		return err
	}
	return nil
}

func MigrateImagesToCloudflare() error {
	if err := run(Command{
		Type: CmdGoBuild,
		Out:  filepath.Join(tmpDir, "migrate_images_cloudflare"),
		Tags: []string{"staticfs"},
	}); err != nil {
		return err
	}
	if err := run(Command{
		Type: CmdTmpExec,
		Cmd:  "migrate_images_cloudflare",
		Args: []string{
			"migrate_images_cloudflare",
			"--dry-run=false",
		},
	}); err != nil {
		return err
	}
	return nil
}

func GenMaps() error {
	return run(Command{
		Type: CmdGoRun,
		Tags: []string{"staticfs"},
		Args: []string{"./main.go", "parse_map",
			"--filepath=./assets/xlsx/PSGC-2Q-2025-Publication-Datafile.xlsx",
			"--json=true",
		},
	})
}

func CleanDB() error {
	fmt.Println("Cleaning", dbName)
	os.Remove(dbName)
	os.Remove(dbName + "-shm")
	os.Remove(dbName + "-wal")

	if err := GenSQL(); err != nil {
		return err
	}
	if err := run(Command{
		Type: CmdTmpExec,
		Cmd:  "goose",
		Args: []string{"up"},
	}); err != nil {
		return err
	}
	return run(Command{
		Type: CmdGoRun,
		Tags: []string{"fts5", "staticfs"},
		Args: []string{"./main.go", "parse_products",
			"-p", "assets/xlsx/bosch.xlsx",
			"-s", "DATABASE", "-t", "BOSCH",
			"--use_db", "--db_path", dbPath,
			"--verify_prices=1", "--panic_on_error=1",
			"--images_basepath=./cmd/web/static/images/product_images/bosch/original/",
			"--images_format=webp",
		},
	})
}

func Deps() error {
	if _, err := os.Stat("./tailwindcss"); os.IsNotExist(err) {
		bin := "tailwindcss-linux-x64"
		if isMac {
			bin = "tailwindcss-macos-arm64"
		}
		url := fmt.Sprintf("https://github.com/tailwindlabs/tailwindcss/releases/latest/download/%s", bin)
		if err := run(Command{Type: CmdExec, Cmd: "curl", Args: []string{"-LO", url}}); err != nil {
			return err
		}
		if err := os.Chmod(bin, PERM); err != nil {
			return err
		}
		if err := os.Rename(bin, "tailwindcss"); err != nil {
			return err
		}
	} else {
		fmt.Println("You already have tailwindcss binary")
	}

	if isWSL {
		switch detectDistro() {
		case "arch":
			if err := DepsArch(); err != nil {
				return err
			}
		case "debian":
			if err := DepsDebian(); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported distro")
		}
	} else if isMac {
		if err := DepsMac(); err != nil {
			return err
		}
	}

	return BuildGoose()
}

func DepsArch() error {
	fmt.Println("Installing dependencies for Arch...")
	return run(Command{Type: CmdExec, Cmd: "yay", Args: []string{
		"-S", "--noconfirm",
		"base-devel", "glib2", "expat1", "libdeflate",
		"libvips", "libmagick", "openslide", "libxml2",
		"libjxl", "golangci-lint-bin",
	}})
}

func DepsDebian() error {
	fmt.Println("Installing dependencies for Debian...")
	if err := run(Command{Type: CmdExec, Cmd: "sudo", Args: []string{"apt", "update"}}); err != nil {
		return err
	}
	return run(Command{Type: CmdExec, Cmd: "sudo", Args: []string{"apt", "install", "-y",
		"build-essential", "golang-go", "git", "sqlite3", "libsqlite3-dev",
		"libvips-dev", "libmagickwand-dev", "openslide-tools",
		"libxml2-dev", "libjxl-dev", "curl"}})
}

func DepsMac() error {
	fmt.Println("Installing dependencies for MacOS...")
	return run(Command{Type: CmdExec, Cmd: "brew", Args: []string{"install",
		"go", "git", "sqlite", "vips", "imagemagick",
		"openslide", "libxml2", "jpeg-xl", "curl", "golangci-lint"}})
}

func GenSQL() error {
	return run(Command{Type: CmdExec, Cmd: "go", Args: []string{"tool", "sqlc", "generate"}})
}

func GenTempl() error {
	if err := run(Command{
		Type: CmdExec,
		Cmd:  "./tailwindcss",
		Args: []string{"-m", "-i", "./cmd/web/static/css/main.css", "-o", "./cmd/web/static/css/tailwind.css"},
	}); err != nil {
		return err
	}
	return run(Command{Type: CmdExec, Cmd: "go", Args: []string{"tool", "templ", "generate", "templ", "-v"}})
}

func GenAll() error {
	if err := run(Command{Type: CmdExec, Cmd: "go", Args: []string{"generate", "./..."}}); err != nil {
		return err
	}
	// Run genversion from internal/conf directory so it writes version_gen.go in the correct location
	originalDir, err := os.Getwd()
	if err != nil {
		return err
	}
	if err := os.Chdir("internal/conf"); err != nil {
		return err
	}
	if err := run(Command{Type: CmdExec, Cmd: "go", Args: []string{"run", "../../cmd/genversion/genversion.go"}}); err != nil {
		_ = os.Chdir(originalDir)
		return err
	}
	if err := os.Chdir(originalDir); err != nil {
		return err
	}
	if err := GenSQL(); err != nil {
		return err
	}
	return GenTempl()
}

func GenChlog() error {
	fmt.Println("Always create a `git tag` first before running this command")
	fmt.Print("Do you want to proceed? (y/N): ")

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	response = strings.TrimSpace(strings.ToLower(response))
	if response != "y" && response != "yes" {
		fmt.Println("Cancelled.")
		return nil
	}

	if err := GenAll(); err != nil {
		return err
	}

	return run(Command{
		Type: CmdExec,
		Cmd:  "go",
		Args: []string{
			"tool",
			"git-chglog",
			"-o",
			"CHANGELOGS.md",
		},
	})
}

func SC() error {
	steps := [][]string{
		{"go", "fmt", "./..."},
		{"go", "mod", "tidy"},
		{"go", "vet", "./..."},
		{"go", "tool", "templ", "fmt", "./cmd/web/components"},
		{"go", "tool", "betteralign", "-apply", "./..."},
		{"go", "tool", "nilaway", "./..."},
		{"go", "tool", "smrcptr", "./..."},
		{"go", "tool", "unconvert", "./..."},
		{"go", "run", "golang.org/x/tools/gopls/internal/analysis/modernize/cmd/modernize@latest", "-test", "./..."},
		{"go", "tool", "govulncheck", "./..."},
	}
	for _, step := range steps {
		if err := run(Command{Type: CmdExec, Cmd: step[0], Args: step[1:]}); err != nil {
			return err
		}
	}
	return nil
}

func hasGoFileChanges() (bool, error) {
	cmd := exec.Command("git", "diff", "--name-only", "HEAD", "--", "*.go")
	output, err := cmd.Output()
	if err != nil {
		cmd = exec.Command("git", "diff", "--name-only", "--cached", "--", "*.go")
		output, err = cmd.Output()
		if err != nil {
			return true, nil
		}
	}
	return strings.TrimSpace(string(output)) != "", nil
}

func hasPackageChanges(packages []string) (bool, error) {
	for _, pkg := range packages {
		changed, err := hasPackageChanged(pkg)
		if err != nil {
			return false, err
		}
		if changed {
			return true, nil
		}
	}
	return false, nil
}

func hasPackageChanged(pkg string) (bool, error) {
	path := fmt.Sprintf("internal/%s/", pkg)
	args := []string{"diff", "--name-only", "HEAD", "--", path}
	cmd := exec.Command("git", args...)
	output, err := cmd.Output()
	if err != nil {
		args = []string{"diff", "--name-only", "--cached", "--", path}
		cmd = exec.Command("git", args...)
		output, err = cmd.Output()
		if err != nil {
			return true, nil
		}
	}
	return strings.TrimSpace(string(output)) != "", nil
}

func TestAll() error {
	hasChanges, err := hasGoFileChanges()
	if err != nil {
		return fmt.Errorf("failed to check for Go file changes: %w", err)
	}
	if !hasChanges {
		fmt.Println("No changes in Go files detected. Skipping tests.")
		return nil
	}

	if _, err := exec.LookPath("golangci-lint"); err == nil {
		if err := run(Command{Type: CmdExec, Cmd: "golangci-lint", Args: []string{"config", "verify"}}); err != nil {
			return err
		}
		if err := run(Command{Type: CmdExec, Cmd: "golangci-lint", Args: []string{"run"}}); err != nil {
			return err
		}
	}

	return run(Command{
		Type: CmdExec,
		Cmd:  "go",
		Args: append([]string{"test", "./...", "-failfast"}),
	})
}

func TestInteg() error {
	type integTest struct {
		pkg  string
		args []string
	}

	tests := []integTest{
		{pkg: "storage/linode", args: []string{"test_linode"}},
		{pkg: "receipt/scanner/googlevision", args: []string{"test_gvision"}},
		{pkg: "payments/paymongo", args: []string{"test_payment"}},
		{pkg: "shipping/lalamove", args: []string{"test_shipping"}},
		{pkg: "mail/maileroo", args: []string{"test_mail", "-t", "flamendless8@gmail.com"}},
	}

	var testsToRun []integTest
	for _, t := range tests {
		changed, err := hasPackageChanged(t.pkg)
		if err != nil {
			return fmt.Errorf("failed to check for package changes: %w", err)
		}
		if changed {
			testsToRun = append(testsToRun, t)
		}
	}

	if len(testsToRun) == 0 {
		fmt.Println("No changes in integration test packages detected. Skipping integration tests.")
		return nil
	}

	if err := run(Command{
		Type: CmdGoBuild,
		Out:  filepath.Join(tmpDir, "main"),
		Tags: []string{"fts5", "staticfs"},
	}); err != nil {
		return err
	}

	for _, t := range testsToRun {
		fmt.Printf("Running integration test for %s\n", t.pkg)
		if err := run(Command{
			Type: CmdTmpExec,
			Cmd:  "main",
			Args: t.args,
		}); err != nil {
			return err
		}
	}

	return nil
}

func TestSum() error {
	return run(Command{
		Type: CmdExec,
		Cmd:  "go",
		Args: []string{"tool", "gotestsum",
			"--debug", "--format=pkgname-and-test-fails", "--format-icons=default",
			"--format-hide-empty-pkg", "--hide-summary=skipped",
			"--", "-cover", "-shuffle=on", "-race", "-test.v", "./..."},
	})
}

func Benchmark() error {
	return run(Command{Type: CmdExec, Cmd: "go", Args: append([]string{"test", "-bench=.", "-benchmem", "./..."})})
}

func Prof(pkg, profType string) error {
	return run(Command{
		Type: CmdExec,
		Cmd:  "go",
		Args: []string{"test", "-cpuprofile", filepath.Join(tmpDir, "cpu.prof"),
			"-memprofile", filepath.Join(tmpDir, "mem.prof"),
			"-benchmem", "-bench=.", "-o", tmpDir, "./" + pkg},
	})
}

func Dev() error {
	if err := GenAll(); err != nil {
		return err
	}
	if err := run(Command{
		Type: CmdGoBuild,
		Out:  filepath.Join(tmpDir, "cchoicedev"),
		Tags: []string{"fts5", "staticfs", "imageprocessing"},
	}); err != nil {
		return err
	}
	fmt.Println("Run: ./tmp/cchoicedev api > out 2>&1 &")
	return nil
}

func Prod() error {
	if err := checkMigrations(); err != nil {
		return err
	}
	if err := GenAll(); err != nil {
		return err
	}
	if err := run(Command{
		Type: CmdGoBuild,
		Out:  filepath.Join(tmpDir, "cchoiceprod"),
		Tags: []string{"fts5", "staticfs", "imageprocessing"},
	}); err != nil {
		return err
	}
	fmt.Println("Run: ./tmp/cchoiceprod api > out 2>&1 &")
	return nil
}

func DBUp() error {
	return run(Command{
		Type: CmdTmpExec,
		Cmd:  "goose",
		Args: []string{"up"},
	})
}

func DBDown() error {
	return run(Command{
		Type: CmdTmpExec,
		Cmd:  "goose",
		Args: []string{"down"},
	})
}

func Prom() error {
	if !checkProc("./tmp/main", 7331) {
		return fmt.Errorf("api process not running")
	}
	openBrowser("http://localhost:9090/")
	return run(Command{
		Type: CmdExec,
		Cmd:  "prometheus",
		Args: []string{
			"--config.file=prometheus/prometheus.yml",
		},
	})
}

func Graf() error {
	if !checkProc("prometheus", 9090) {
		return fmt.Errorf("prometheus not running")
	}
	openBrowser("http://localhost:3000/")
	return nil
}
