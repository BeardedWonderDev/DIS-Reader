## Agent TUI (`agentctl`)

### What it is
An interactive terminal UI that talks to the agent’s local control API to configure and manage the service (install/start/stop/restart) without editing files by hand.

### Prerequisites
- Agent control API enabled in `agent.yaml` (or env):
  ```yaml
  control:
    enabled: true
    addr: "127.0.0.1:7777"
    token: "strong-token"
  ```
- The agent process running (it can be started manually once to expose the control API).

### Install
1) Download the `agentctl` binary for your OS/arch from the GitHub release assets.
2) Make it executable and place it on PATH (e.g., `chmod +x agentctl && sudo mv agentctl /usr/local/bin`).

### Use
```bash
agentctl --addr http://127.0.0.1:7777 --token <token>
```

From the TUI:
- **Config form**: set `serverURL`, `clientID/Secret`, DIS host/user/password, JDBC port/java path, TLS flags, and auto-connect. Save & Apply writes via `/config/apply`.
- **Service controls**: Install, Start, Stop, Restart, Status (invokes `/service/*` on the control API).
- **Status view**: shows bridge connected state, agent ID, server URL, last heartbeat/error.

### First-run workflow
1. Fill config fields, choose “Save & Apply”.
2. Choose “Install Service” (systemd/launchd/SCM, depending on OS).
3. Choose “Start Service”.
4. Verify status shows bridge connected and heartbeats progressing.

### Notes
- If `control.token` is empty, the API is unauthenticated—only do this in isolated/dev environments.
- The TUI stores nothing on disk; all persistence is through the control API/apply call.
