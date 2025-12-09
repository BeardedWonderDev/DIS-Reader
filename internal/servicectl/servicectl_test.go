package servicectl

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// helperCommandContext launches this test binary in helper mode to avoid executing real system commands.
func helperCommandContext(scenario string, record *[]string) func(context.Context, string, ...string) *exec.Cmd {
	return func(ctx context.Context, name string, args ...string) *exec.Cmd {
		if record != nil {
			*record = append(*record, name+" "+strings.Join(args, " "))
		}
		cmdArgs := append([]string{"-test.run=TestHelperProcess", "--", scenario}, args...)
		cmd := exec.CommandContext(ctx, os.Args[0], cmdArgs...)
		cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")
		return cmd
	}
}

// TestHelperProcess is executed as a subprocess. It simulates command success/failure.
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	args := os.Args
	sep := 0
	for i, a := range args {
		if a == "--" {
			sep = i
			break
		}
	}
	scenario := "ok"
	if sep < len(args)-1 {
		scenario = args[sep+1]
	}
	switch scenario {
	case "ok":
		os.Stdout.WriteString("ok")
		os.Exit(0)
	case "fail":
		os.Stdout.WriteString("fail")
		os.Exit(1)
	default:
		os.Stdout.WriteString(scenario)
		os.Exit(0)
	}
}

func TestRunAndRunOutput(t *testing.T) {
	old := commandContext
	defer func() { commandContext = old }()

	// success path
	commandContext = helperCommandContext("ok", nil)
	if err := run(context.Background(), "systemctl", "start"); err != nil {
		t.Fatalf("run should succeed: %v", err)
	}

	// failure path
	commandContext = helperCommandContext("fail", nil)
	if err := run(context.Background(), "systemctl", "start"); err == nil {
		t.Fatalf("run should fail when helper exits non-zero")
	}

	// runOutput returns stdout even on success
	commandContext = helperCommandContext("ok", nil)
	out, err := runOutput(context.Background(), "systemctl", "status")
	if err != nil {
		t.Fatalf("runOutput should succeed: %v", err)
	}
	if !strings.Contains(out, "ok") {
		t.Fatalf("expected output to contain ok, got %q", out)
	}
}

func TestControllersUseExpectedCommands(t *testing.T) {
	old := commandContext
	defer func() { commandContext = old }()

	var calls []string
	commandContext = helperCommandContext("ok", &calls)

	ctx := context.Background()

	// systemd
	sys := &systemdController{}
	if err := sys.Start(ctx); err != nil {
		t.Fatalf("systemd start: %v", err)
	}
	if !strings.HasPrefix(calls[len(calls)-1], "systemctl start dis-agent") {
		t.Fatalf("unexpected systemd start call: %v", calls[len(calls)-1])
	}

	// launchd
	l := &launchdController{}
	if err := l.Install(ctx); err != nil {
		t.Fatalf("launchd install: %v", err)
	}
	if !strings.HasPrefix(calls[len(calls)-1], "launchctl bootstrap system /Library/LaunchDaemons/com.dis.agent.plist") {
		t.Fatalf("unexpected launchd install call: %v", calls[len(calls)-1])
	}

	// windows
	w := &windowsController{}
	if err := w.Start(ctx); err != nil {
		t.Fatalf("windows start: %v", err)
	}
	if !strings.HasPrefix(calls[len(calls)-1], "sc.exe start dis-agent") {
		t.Fatalf("unexpected windows start call: %v", calls[len(calls)-1])
	}

	// noop
	n := &noopController{}
	if err := n.Install(ctx); err == nil {
		t.Fatalf("expected noop controller to error")
	}
}

func TestRestartSequencesAndFailures(t *testing.T) {
	old := commandContext
	defer func() { commandContext = old }()

	// launchd restart should issue bootout then kickstart in order
	var calls []string
	commandContext = helperCommandContext("ok", &calls)
	l := &launchdController{}
	if err := l.Restart(context.Background()); err != nil {
		t.Fatalf("launchd restart: %v", err)
	}
	if len(calls) < 2 || !strings.HasPrefix(calls[0], "launchctl bootout") || !strings.HasPrefix(calls[1], "launchctl kickstart") {
		t.Fatalf("unexpected launchd restart sequence: %v", calls)
	}

	// windows restart should stop then start; ensure error surfaces when stop fails
	commandContext = helperCommandContext("fail", nil)
	w := &windowsController{}
	if err := w.Restart(context.Background()); err == nil {
		t.Fatalf("expected restart to fail when stop fails")
	}
}
