## Agent GUI (`dis-agent-gui`)

### What it is
A Wails-based desktop UI to configure and manage the agent service via the local control API. It mirrors the TUI capabilities with a point-and-click experience.

### Prerequisites
- Control API enabled in `agent.yaml` (or env):
  ```yaml
  control:
    enabled: true
    addr: "127.0.0.1:7777"
    token: "strong-token"
  ```
- Agent running (at least once) to expose the control API.
- Platform deps:
  - macOS: native WebKit (already present).
  - Windows: WebView2 runtime (Wails downloads it if missing).
  - Linux: GTK/WebKit2 (install `libwebkit2gtk-4.1-dev`, `libgtk-3-dev` or distro equivalents).

### Install
1) Download `dis-agent-gui` for your OS/arch from the release assets (App bundle/EXE/tar.gz).
2) macOS: `chmod +x dis-agent-gui` (if tarball) and run; Windows: run EXE; Linux: `chmod +x dis-agent-gui && ./dis-agent-gui`.

### Use
1. Set **Control API** URL and **Token** (persisted locally).
2. (Optional) Set auto-refresh interval (seconds).
3. Fill config fields: `serverURL`, `clientID/Secret`, DIS host/user/password, JDBC port, Java path, TLS flags, auto-connect.
4. Actions:
   - **Save & Apply**: writes config via `/config/apply`.
   - **Save + Restart**: apply then restart service.
   - **Save + Install + Start**: best for first run—apply config, install service, start service.
   - **Service buttons**: Install, Start, Stop, Restart, Status.
5. Status pane shows bridge connected state, agent/server info, last heartbeat/error, and raw JSON status.

### First-run recipe
1. Enable control API in `agent.yaml`, start agent once.
2. Launch GUI, point to `http://127.0.0.1:7777` with your token.
3. Fill config, click **Save + Install + Start**.
4. Confirm status turns green and heartbeats advance; then you can close the GUI—the service keeps running.

### Notes
- GUI stores only UI preferences (API URL, token, refresh interval) in localStorage; agent config lives on the agent via the control API.
- Tenant/runtime/observability fields remain hidden per policy; remote-pushed overrides still apply on the agent side.
