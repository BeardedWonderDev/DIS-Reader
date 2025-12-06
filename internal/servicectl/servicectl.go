package servicectl

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
)

// Controller defines service lifecycle operations.
type Controller interface {
	Install(ctx context.Context) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Restart(ctx context.Context) error
	Status(ctx context.Context) (string, error)
}

// New returns an OS-specific controller for the dis-agent service.
func New() Controller {
	switch runtime.GOOS {
	case "linux":
		return &systemdController{}
	case "darwin":
		return &launchdController{}
	case "windows":
		return &windowsController{}
	default:
		return &noopController{}
	}
}

// --- Linux / systemd ---
type systemdController struct{}

func (s *systemdController) Install(ctx context.Context) error {
	// Ensure unit is reloaded and enabled; relies on packaged unit file.
	if err := run(ctx, "systemctl", "daemon-reload"); err != nil {
		return err
	}
	if err := run(ctx, "systemctl", "enable", "dis-agent"); err != nil {
		return err
	}
	return nil
}

func (s *systemdController) Start(ctx context.Context) error {
	return run(ctx, "systemctl", "start", "dis-agent")
}
func (s *systemdController) Stop(ctx context.Context) error {
	return run(ctx, "systemctl", "stop", "dis-agent")
}
func (s *systemdController) Restart(ctx context.Context) error {
	return run(ctx, "systemctl", "restart", "dis-agent")
}

func (s *systemdController) Status(ctx context.Context) (string, error) {
	out, err := runOutput(ctx, "systemctl", "is-active", "dis-agent")
	return out, err
}

// --- macOS / launchd ---
type launchdController struct{}

func (l *launchdController) Install(ctx context.Context) error {
	// Assumes plist installed at /Library/LaunchDaemons/com.dis.agent.plist
	return run(ctx, "launchctl", "bootstrap", "system", "/Library/LaunchDaemons/com.dis.agent.plist")
}

func (l *launchdController) Start(ctx context.Context) error {
	return run(ctx, "launchctl", "kickstart", "-k", "system/com.dis.agent")
}

func (l *launchdController) Stop(ctx context.Context) error {
	return run(ctx, "launchctl", "bootout", "system", "/Library/LaunchDaemons/com.dis.agent.plist")
}

func (l *launchdController) Restart(ctx context.Context) error {
	if err := l.Stop(ctx); err != nil {
		return err
	}
	return l.Start(ctx)
}

func (l *launchdController) Status(ctx context.Context) (string, error) {
	return runOutput(ctx, "launchctl", "print", "system/com.dis.agent")
}

// --- Windows / SCM ---
type windowsController struct{}

func (w *windowsController) Install(ctx context.Context) error {
	// Best effort: rely on existing install_service.ps1 in current dir.
	return run(ctx, "powershell", "-ExecutionPolicy", "Bypass", "-File", "install_service.ps1", "-NoStart")
}

func (w *windowsController) Start(ctx context.Context) error {
	return run(ctx, "sc.exe", "start", "dis-agent")
}
func (w *windowsController) Stop(ctx context.Context) error {
	return run(ctx, "sc.exe", "stop", "dis-agent")
}

func (w *windowsController) Restart(ctx context.Context) error {
	if err := w.Stop(ctx); err != nil {
		return err
	}
	return w.Start(ctx)
}

func (w *windowsController) Status(ctx context.Context) (string, error) {
	return runOutput(ctx, "sc.exe", "query", "dis-agent")
}

// --- Noop fallback ---
type noopController struct{}

func (n *noopController) Install(ctx context.Context) error {
	return fmt.Errorf("service control not supported on this platform")
}
func (n *noopController) Start(ctx context.Context) error            { return n.Install(ctx) }
func (n *noopController) Stop(ctx context.Context) error             { return n.Install(ctx) }
func (n *noopController) Restart(ctx context.Context) error          { return n.Install(ctx) }
func (n *noopController) Status(ctx context.Context) (string, error) { return "", n.Install(ctx) }

// --- helpers ---

func run(ctx context.Context, cmd string, args ...string) error {
	c := exec.CommandContext(ctx, cmd, args...)
	if out, err := c.CombinedOutput(); err != nil {
		return fmt.Errorf("%s: %w (output: %s)", cmd, err, string(out))
	}
	return nil
}

func runOutput(ctx context.Context, cmd string, args ...string) (string, error) {
	c := exec.CommandContext(ctx, cmd, args...)
	out, err := c.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("%s: %w (output: %s)", cmd, err, string(out))
	}
	return string(out), nil
}
