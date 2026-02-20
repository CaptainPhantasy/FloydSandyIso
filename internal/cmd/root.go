package cmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/fang"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/exp/charmtone"
	"github.com/charmbracelet/x/term"
	"github.com/CaptainPhantasy/FloydSandyIso/internal/app"
	"github.com/CaptainPhantasy/FloydSandyIso/internal/config"
	"github.com/CaptainPhantasy/FloydSandyIso/internal/db"
	"github.com/CaptainPhantasy/FloydSandyIso/internal/event"
	"github.com/CaptainPhantasy/FloydSandyIso/internal/projects"
	"github.com/CaptainPhantasy/FloydSandyIso/internal/telemetry"
	"github.com/CaptainPhantasy/FloydSandyIso/internal/ui/common"
	ui "github.com/CaptainPhantasy/FloydSandyIso/internal/ui/model"
	"github.com/CaptainPhantasy/FloydSandyIso/internal/version"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.PersistentFlags().StringP("cwd", "c", "", "Current working directory")
	rootCmd.PersistentFlags().StringP("data-dir", "D", "", "Custom floyd data directory")
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "Debug")
	rootCmd.Flags().BoolP("help", "h", false, "Help")
	rootCmd.PersistentFlags().BoolP("yolo", "y", false, "Automatically accept all permissions (dangerous mode)")

	rootCmd.AddCommand(
		runCmd,
		dirsCmd,
		projectsCmd,
		updateProvidersCmd,
		logsCmd,
		schemaCmd,
		loginCmd,
		statsCmd,
	)
}

var rootCmd = &cobra.Command{
	Use:   "floyd",
	Short: "An AI assistant for software development",
	Long:  "An AI assistant for software development and similar tasks with direct access to the terminal",
	Example: `
# Run in interactive mode
floyd

# Run with debug logging
floyd -d

# Run with debug logging in a specific directory
floyd -d -c /path/to/project

# Run with custom data directory
floyd -D /path/to/custom/.floyd

# Print version
floyd -v

# Run a single non-interactive prompt
floyd run "Explain the use of context in Go"

# Run in dangerous mode (auto-accept all permissions)
floyd -y
  `,
	RunE: func(cmd *cobra.Command, args []string) error {
		app, err := setupAppWithProgressBar(cmd)
		if err != nil {
			return err
		}
		defer app.Shutdown()

		event.AppInitialized()

		// Set up the TUI.
		var env uv.Environ = os.Environ()

		com := common.DefaultCommon(app)
		model := ui.New(com)

		program := tea.NewProgram(
			model,
			tea.WithEnvironment(env),
			tea.WithContext(cmd.Context()),
			tea.WithFilter(ui.MouseEventFilter), // Filter mouse events based on focus state
		)
		go app.Subscribe(program)

		if _, err := program.Run(); err != nil {
			event.Error(err)
			slog.Error("TUI run error", "error", err)
			return errors.New("Floyd crashed. If metrics are enabled, we were notified about it. If you'd like to report it, please copy the stacktrace above and open an issue at https://github.com/CaptainPhantasy/FloydSandyIso/issues/new?template=bug.yml") //nolint:staticcheck
		}
		return nil
	},
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if shouldSkipInit(cmd) {
			return nil
		}
		if err := initTelemetryForCmd(cmd); err != nil {
			slog.Debug("Telemetry init failed", "error", err)
		}
		return nil
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		telemetry.Default().Track(telemetry.EventCommandSuccess, map[string]any{
			"command": cmd.CommandPath(),
		})
		telemetry.Default().FlushSync()
	},
	PostRun: func(cmd *cobra.Command, args []string) {
		event.AppExited()
	},
}

var heartbit = lipgloss.NewStyle().Foreground(charmtone.Dolly).SetString(`
    __/\\\\\\\\\\\\\\\___/\\\____________________/\\\\\________/\\\________/\\\___/\\\\\\\\\\\\_______________
     _\/\\\///////////___\/\\\__________________/\\\///\\\_____\///\\\____/\\\/___\/\\\////////\\\_____________
      _\/\\\______________\/\\\________________/\\\/__\///\\\_____\///\\\/\\\/_____\/\\\______\//\\\____________
       _\/\\\\\\\\\\\______\/\\\_______________/\\\______\//\\\______\///\\\/_______\/\\\_______\/\\\____________
        _\/\\\///////_______\/\\\______________\/\\\_______\/\\\________\/\\\________\/\\\_______\/\\\____________
         _\/\\\______________\/\\\______________\//\\\______/\\\_________\/\\\________\/\\\_______\/\\\____________
          _\/\\\______________\/\\\_______________\///\\\__/\\\___________\/\\\________\/\\\_______/\\\_____________
           _\/\\\______________\/\\\\\\\\\\\\\\\_____\///\\\\\/____________\/\\\________\/\\\\\\\\\\\\/______________
            _\///_______________\///////////////________\/////______________\///_________\////////////________________
`)

// copied from cobra:
const defaultVersionTemplate = `{{with .DisplayName}}{{printf "%s " .}}{{end}}{{printf "version %s" .Version}}
`

func Execute() {
	for _, arg := range os.Args[1:] {
		if arg == "--version" || arg == "-v" {
			fmt.Printf("floyd version %s\n", version.Version)
			return
		}
	}
	// NOTE: very hacky: we create a colorprofile writer with STDOUT, then make
	// it forward to a bytes.Buffer, write the colored heartbit to it, and then
	// finally prepend it in the version template.
	// Unfortunately cobra doesn't give us a way to set a function to handle
	// printing the version, and PreRunE runs after the version is already
	// handled, so that doesn't work either.
	// This is the only way I could find that works relatively well.
	if term.IsTerminal(os.Stdout.Fd()) {
		var b bytes.Buffer
		w := colorprofile.NewWriter(os.Stdout, os.Environ())
		w.Forward = &b
		_, _ = w.WriteString(heartbit.String())
		rootCmd.SetVersionTemplate(b.String() + "\n" + defaultVersionTemplate)
	}
	if err := fang.Execute(
		context.Background(),
		rootCmd,
		fang.WithVersion(version.Version),
		fang.WithNotifySignal(os.Interrupt),
	); err != nil {
		os.Exit(1)
	}
}

// supportsProgressBar tries to determine whether the current terminal supports
// progress bars by looking into environment variables.
func supportsProgressBar() bool {
	if !term.IsTerminal(os.Stderr.Fd()) {
		return false
	}
	termProg := os.Getenv("TERM_PROGRAM")
	_, isWindowsTerminal := os.LookupEnv("WT_SESSION")

	return isWindowsTerminal || strings.Contains(strings.ToLower(termProg), "ghostty")
}

func setupAppWithProgressBar(cmd *cobra.Command) (*app.App, error) {
	app, err := setupApp(cmd)
	if err != nil {
		return nil, err
	}

	// Check if progress bar is enabled in config (defaults to true if nil)
	progressEnabled := app.Config().Options.Progress == nil || *app.Config().Options.Progress
	if progressEnabled && supportsProgressBar() {
		_, _ = fmt.Fprintf(os.Stderr, ansi.SetIndeterminateProgressBar)
		defer func() { _, _ = fmt.Fprintf(os.Stderr, ansi.ResetProgressBar) }()
	}

	return app, nil
}

// setupApp handles the common setup logic for both interactive and non-interactive modes.
// It returns the app instance, config, cleanup function, and any error.
func setupApp(cmd *cobra.Command) (*app.App, error) {
	debug, _ := cmd.Flags().GetBool("debug")
	yolo, _ := cmd.Flags().GetBool("yolo")
	dataDir, _ := cmd.Flags().GetString("data-dir")
	ctx := cmd.Context()

	cwd, err := ResolveCwd(cmd)
	if err != nil {
		return nil, err
	}

	cfg, err := config.Init(cwd, dataDir, debug)
	if err != nil {
		return nil, err
	}

	telemetry.InitDefault(shouldEnableMetrics(), telemetry.DefaultAdditionalData())
	telemetry.Default().Track(telemetry.EventCLIStart, map[string]any{
		"mode": "tui",
	})

	if cfg.Permissions == nil {
		cfg.Permissions = &config.Permissions{}
	}
	cfg.Permissions.SkipRequests = yolo

	if err := createDotFloydDir(cfg.Options.DataDirectory); err != nil {
		return nil, err
	}

	// Register this project in the centralized projects list.
	if err := projects.Register(cwd, cfg.Options.DataDirectory); err != nil {
		slog.Warn("Failed to register project", "error", err)
		// Non-fatal: continue even if registration fails
	}

	// Connect to DB; this will also run migrations.
	conn, err := db.Connect(ctx, cfg.Options.DataDirectory)
	if err != nil {
		return nil, err
	}

	appInstance, err := app.New(ctx, conn, cfg)
	if err != nil {
		slog.Error("Failed to create app instance", "error", err)
		return nil, err
	}

	if shouldEnableMetrics() {
		event.Init()
	}

	return appInstance, nil
}

func initTelemetryForCmd(cmd *cobra.Command) error {
	debug, _ := cmd.Flags().GetBool("debug")
	dataDir, _ := cmd.Flags().GetString("data-dir")
	cwd, err := ResolveCwd(cmd)
	if err != nil {
		return err
	}
	if _, err := config.Init(cwd, dataDir, debug); err != nil {
		return err
	}
	telemetry.InitDefault(shouldEnableMetrics(), telemetry.DefaultAdditionalData())
	telemetry.Default().Track(telemetry.EventCommandRun, map[string]any{
		"command": cmd.CommandPath(),
	})
	return nil
}

func shouldSkipInit(cmd *cobra.Command) bool {
	for _, arg := range os.Args {
		if arg == "--help" || arg == "-h" || arg == "--version" || arg == "-v" || arg == "help" || arg == "version" {
			return true
		}
	}
	if cmd != nil {
		if flag := cmd.Flag("help"); flag != nil && flag.Changed {
			return true
		}
		if flag := cmd.Flag("version"); flag != nil && flag.Changed {
			return true
		}
	}
	return false
}

func shouldEnableMetrics() bool {
	if v, _ := strconv.ParseBool(os.Getenv("FLOYD_DISABLE_METRICS")); v {
		return false
	}
	if v, _ := strconv.ParseBool(os.Getenv("DO_NOT_TRACK")); v {
		return false
	}
	if config.Get().Options.DisableMetrics {
		return false
	}
	return true
}

func MaybePrependStdin(prompt string) (string, error) {
	if term.IsTerminal(os.Stdin.Fd()) {
		return prompt, nil
	}
	fi, err := os.Stdin.Stat()
	if err != nil {
		return prompt, err
	}
	// Check if stdin is a named pipe ( | ) or regular file ( < ).
	if fi.Mode()&os.ModeNamedPipe == 0 && !fi.Mode().IsRegular() {
		return prompt, nil
	}
	bts, err := io.ReadAll(os.Stdin)
	if err != nil {
		return prompt, err
	}
	return string(bts) + "\n\n" + prompt, nil
}

func ResolveCwd(cmd *cobra.Command) (string, error) {
	cwd, _ := cmd.Flags().GetString("cwd")
	if cwd != "" {
		err := os.Chdir(cwd)
		if err != nil {
			return "", fmt.Errorf("failed to change directory: %v", err)
		}
		return cwd, nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current working directory: %v", err)
	}
	return cwd, nil
}

func createDotFloydDir(dir string) error {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("failed to create data directory: %q %w", dir, err)
	}

	gitIgnorePath := filepath.Join(dir, ".gitignore")
	if _, err := os.Stat(gitIgnorePath); os.IsNotExist(err) {
		if err := os.WriteFile(gitIgnorePath, []byte("*\n"), 0o644); err != nil {
			return fmt.Errorf("failed to create .gitignore file: %q %w", gitIgnorePath, err)
		}
	}

	return nil
}
