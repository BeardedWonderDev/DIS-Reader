<!-- Improved compatibility of back to top link: See: https://github.com/othneildrew/Best-README-Template/pull/73 -->
<a id="readme-top"></a>

<!-- PROJECT SHIELDS -->
[![Contributors][contributors-shield]][contributors-url]
[![Forks][forks-shield]][forks-url]
[![Stargazers][stars-shield]][stars-url]
[![Issues][issues-shield]][issues-url]
[![project_license][license-shield]][license-url]
[![LinkedIn][linkedin-shield]][linkedin-url]

<!-- PROJECT LOGO -->
<br />
<div align="center">
  <a href="https://github.com/BeardedWonderDev/DIS-Reader">
    <img src="images/logo.png" alt="Logo" width="80" height="80">
  </a>

<h3 align="center">DIS Reader</h3>

  <p align="center">
    A bridge service that lets your applications read DIS (AS/400) data easily—embedded JDBC or remote agent, same API.
    <br />
    <a href="#about-the-project"><strong>Explore the docs »</strong></a>
    <br />
    <br />
    <a href="#usage">View Demo</a>
    &middot;
    <a href="https://github.com/BeardedWonderDev/DIS-Reader/issues/new?labels=bug&template=bug-report---.md">Report Bug</a>
    &middot;
    <a href="https://github.com/BeardedWonderDev/DIS-Reader/issues/new?labels=enhancement&template=feature-request---.md">Request Feature</a>
  </p>
</div>

<!-- TABLE OF CONTENTS -->
<details>
  <summary>Table of Contents</summary>
  <ol>
    <li>
      <a href="#about-the-project">About The Project</a>
      <ul>
        <li><a href="#built-with">Built With</a></li>
      </ul>
    </li>
    <li>
      <a href="#getting-started">Getting Started</a>
      <ul>
        <li><a href="#prerequisites">Prerequisites</a></li>
        <li><a href="#installation">Installation</a></li>
      </ul>
    </li>
    <li><a href="#usage">Usage</a></li>
    <li><a href="#roadmap">Roadmap</a></li>
    <li><a href="#contributing">Contributing</a></li>
    <li><a href="#license">License</a></li>
    <li><a href="#contact">Contact</a></li>
    <li><a href="#acknowledgments">Acknowledgments</a></li>
  </ol>
</details>

<!-- ABOUT THE PROJECT -->
## About The Project

[![Product Name Screen Shot][product-screenshot]](images/screenshot.png)

DIS Reader is a Go service that exposes DIS (AS/400) data to other applications with a stable, API-first contract. It runs in two modes:

- **Embedded mode (default):** ships the JDBC runner inside the binary and serves queries directly to the DIS host you configure.
- **Remote agent mode:** a lightweight LAN agent runs the JDBC bridge next to your DIS server and proxies requests to the cloud over gRPC—no inbound ports required.

Capabilities:
- Typed services for units, invoices, and parts with filter/sort/paging helpers.
- Health-checked lifecycle around the JDBC runner (start, connect, ping, shutdown).
- Debug Search batch runner that emits SQLite or CSV datasets for downstream tooling.
- Simple config surface via `disreader.yaml` or `DISREADER_*` environment variables.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

### Built With

* [![Go][Go-shield]][Go-url]
* [![gRPC][gRPC-shield]][gRPC-url]
* [![Protocol Buffers][Proto-shield]][Proto-url]
* [![SQLite][SQLite-shield]][SQLite-url]
* [![Java][Java-shield]][Java-url]

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- GETTING STARTED -->
## Getting Started

Choose embedded when you can reach DIS directly; choose remote when DIS lives behind firewalls and only outbound traffic is allowed.

### Prerequisites
* Go 1.24+
* Java 11+ on PATH (for embedded JDBC runner or agent)
* `protoc`, `protoc-gen-go`, `protoc-gen-go-grpc` (only if you regenerate bridge stubs)
* (Remote mode) TLS egress to the bridge server

### Installation
1. Clone the repo
   ```sh
   git clone https://github.com/BeardedWonderDev/DIS-Reader.git
   cd DIS-Reader
   ```
2. (Optional) Regenerate gRPC stubs
   ```sh
   protoc --go_out=. --go-grpc_out=. proto/bridge.proto
   ```
3. Build binaries
   ```sh
   go build ./cmd/...        # core service binaries
   go build ./cmd/agent/...  # remote agent daemon
   ```

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- USAGE -->
## Usage

### Embedded mode (default)
Start the service using your DIS host credentials:
```sh
DISREADER_DISCONFIG_HOST=10.0.0.5 DISREADER_DISCONFIG_USER=XXXXX DISREADER_DISCONFIG_PASSWORD=XXXXX go run ./...
```
- Config file: `disreader.yaml`
- Key settings: `disConfig.host|user|password`, JDBC `javaPath`/`jdbcPort`, connection pool sizes.
- Consumers call the exported Go services (or any gRPC/HTTP gateway you layer on) to fetch units, invoices, and parts.

### Remote mode (agent bridge)
1. Run the bridge server (cloud side):
   ```sh
   go run ./cmd/bridge-server --bridge.config=disreader.yaml
   ```
2. Run the LAN agent near DIS:
   ```sh
   ./cmd/agent/agent --config agent.yaml
   ```
3. Point clients at the bridge:
   ```sh
   DISREADER_BRIDGE_MODE=remote DISREADER_BRIDGE_SERVERURL="https://bridge.example.com:443" go run ./...
   ```
Bridge config keys:
- `bridge.mode: embedded|remote` (default embedded)
- `bridge.serverURL`, `bridge.clientID`, `bridge.clientSecret`, `bridge.tenantID`
- TLS: `bridge.tls.insecure`, `bridge.tls.caFile`

### Debug Search Output Modes
Produce ad-hoc datasets for analysis:
- Default SQLite: `debug-search.db`
- CSV: set `debugSearch.defaultOutputMode: csv` or `DISREADER_DEBUGSEARCH_DEFAULTOUTPUTMODE=csv`
- Path override: `debugSearch.defaultOutputPath` / `DISREADER_DEBUGSEARCH_DEFAULTOUTPUTPATH`

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- ROADMAP -->
## Roadmap

TODO: replace with the post–remote-agent roadmap once the current phase completes.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- CONTRIBUTING -->
## Contributing

Please run `go test ./...` and `go vet ./...` before opening a PR. Add regression tests for service changes and include sample configs for new bridge/agent settings.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- LICENSE -->
## License

Distributed under the project_license. See `LICENSE.txt` for more information.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- CONTACT -->
## Contact

DIS Reader Team - maintainer@example.com

Project Link: https://github.com/BeardedWonderDev/DIS-Reader

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- ACKNOWLEDGMENTS -->
## Acknowledgments

* IBM i / AS/400 community resources
* gRPC & Protocol Buffers maintainers
* SQLite project
* JT400 / JDBC bridge contributors

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- MARKDOWN LINKS & IMAGES -->
[contributors-shield]: https://img.shields.io/github/contributors/BeardedWonderDev/DIS-Reader.svg?style=for-the-badge
[contributors-url]: https://github.com/BeardedWonderDev/DIS-Reader/graphs/contributors
[forks-shield]: https://img.shields.io/github/forks/BeardedWonderDev/DIS-Reader.svg?style=for-the-badge
[forks-url]: https://github.com/BeardedWonderDev/DIS-Reader/network/members
[stars-shield]: https://img.shields.io/github/stars/BeardedWonderDev/DIS-Reader.svg?style=for-the-badge
[stars-url]: https://github.com/BeardedWonderDev/DIS-Reader/stargazers
[issues-shield]: https://img.shields.io/github/issues/BeardedWonderDev/DIS-Reader.svg?style=for-the-badge
[issues-url]: https://github.com/BeardedWonderDev/DIS-Reader/issues
[license-shield]: https://img.shields.io/github/license/BeardedWonderDev/DIS-Reader.svg?style=for-the-badge
[license-url]: https://github.com/BeardedWonderDev/DIS-Reader/blob/main/LICENSE.txt
[linkedin-shield]: https://img.shields.io/badge/-LinkedIn-black.svg?style=for-the-badge&logo=linkedin&colorB=555
[linkedin-url]: https://linkedin.com
[product-screenshot]: images/screenshot.png
[Go-shield]: https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white
[Go-url]: https://go.dev/
[gRPC-shield]: https://img.shields.io/badge/gRPC-0052CC?style=for-the-badge&logo=google&logoColor=white
[gRPC-url]: https://grpc.io/
[Proto-shield]: https://img.shields.io/badge/Protobuf-3367D6?style=for-the-badge&logo=google&logoColor=white
[Proto-url]: https://developers.google.com/protocol-buffers
[SQLite-shield]: https://img.shields.io/badge/SQLite-003B57?style=for-the-badge&logo=sqlite&logoColor=white
[SQLite-url]: https://www.sqlite.org
[Java-shield]: https://img.shields.io/badge/Java-ED8B00?style=for-the-badge&logo=openjdk&logoColor=white
[Java-url]: https://adoptium.net
