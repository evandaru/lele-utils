# dev-utils — Developer Workstation Toolkit

TUI + CLI toolkit untuk inspeksi workstation developer di **macOS & Linux**.
Satu tempat interaktif untuk pekerjaan yang biasanya butuh banyak command
(`ps`, `lsof`, `go version`, `ip addr`, `docker ps`, `df -h`, ...).

> Positioning: **Developer Workstation Cockpit for macOS & Linux.**
> Bukan replacement terminal, `btop`, atau Docker CLI — melainkan layer
> interaktif di atas system/developer utilities.

## Requirements

- Go 1.24+
- Linux (Ubuntu/Debian/Fedora/Arch/...) atau macOS 13+ (Intel & Apple Silicon)
- Tidak butuh database, tidak butuh internet, tidak butuh root

## Install & Run

```bash
go run ./cmd/dev-utils          # development
make build && ./dev-utils       # binary lokal
make install                    # go install
```

Cross compile (PRD §30-31):

```bash
make cross
# dev-utils-linux-amd64, dev-utils-linux-arm64,
# dev-utils-darwin-amd64, dev-utils-darwin-arm64
```

## TUI

```bash
dev-utils
```

| Key | Aksi |
|-----|------|
| `↑ ↓` / `enter` / `esc` | navigate / open / back |
| `q` | quit |
| `r` | refresh (manual; data dinamis: processes, ports, containers, system, network) |
| `/` | search/filter (semua list utama) |
| `?` | help |
| `k` | kill process (SIGTERM + konfirmasi; `K` = force SIGKILL konfirmasi kedua) |
| `o` / `c` | port: open `http://localhost:PORT` / copy nomor port |
| `s` `t` `r` `l` `k` | container: start / stop / restart / logs / kill |
| `1`-`8` | dashboard shortcut |

Screens: **Dashboard, Processes (+detail & kill), Ports (+detail, kill, open),
Runtimes (15 bahasa/toolchain), Network (interface/gateway/DNS),
Dev Tools, Containers (Docker/Podman), Projects (deteksi Go/Node/PHP/Rust/...),
System (CPU/RAM/disk + environment dengan secrets disembunyikan).**

## CLI (non-interaktif, script/CI/AI-agent friendly)

```bash
dev-utils ports            # list listening ports
dev-utils ports 3000       # detail port 3000
dev-utils process          # list processes (alias: processes, ps)
dev-utils process 18231    # detail PID
dev-utils runtimes         # detect runtimes + versi + path
dev-utils network          # interfaces, gateway, DNS
dev-utils doctor           # diagnostic environment (✓/⚠/✗ + ringkasan)
dev-utils tools            # dev tools
dev-utils system           # CPU/RAM/disk/OS
dev-utils project [dir]    # detect project
dev-utils containers       # docker/podman ps
dev-utils cleanup          # kandidat dev ports
dev-utils cleanup --kill 3000,5173 --yes   # kill eksplisit (wajib --yes)
```

Semua command mendukung `--json`:

```bash
dev-utils ports --json
# [{"port":3000,"protocol":"TCP","pid":18231,"process":"node",...}]
```

## Config (optional)

Tidak membutuhkan config untuk berjalan. Jika ada, dibaca dari:

- Linux: `~/.config/dev-utils/config.toml`
- macOS: `~/Library/Application Support/dev-utils/config.toml`

```toml
refresh_interval = 3
default_screen = "dashboard"
confirm_kill = true
show_system = true
```

Debug log (default silent, tidak mencampur TUI):

```bash
dev-utils --debug   # log → ~/.local/state/dev-utils/dev-utils.log
```

## Security

- Jangan jalankan sebagai root (ada peringatan bila root).
- Tidak pernah meminta sudo saat startup.
- Kill hanya process yang user punya permission (cek signal-0).
- PID 0/1/2 (system) dilindungi.
- Force kill (SIGKILL) butuh konfirmasi kedua.
- Tidak ada arbitrary shell dari input user (semua exec pakai fixed args).
- Environment secrets (`*KEY*`, `*TOKEN*`, `*PASSWORD*`, `*SECRET*`) disembunyikan.

## Architecture

```
UI → Service → Platform abstraction → OS
```

UI tidak pernah `exec` langsung — hanya memanggil mis. `processService.List()`.
Kode OS-specific terisolasi di `internal/platform/{linux,darwin}`.

```
cmd/dev-utils/main.go
internal/app/            # Bubble Tea root model + router + views
internal/ui/{dashboard,processes,ports,runtimes,network,tools,containers,projects,system,shared}
internal/services/{process,port,runtime,network,tools,container,project,system,doctor}
internal/platform/{linux,darwin}
internal/models/         # Process, Port, Runtime, NetworkInterface, ...
internal/config/         # TOML optional
internal/logger/         # silent default, file saat --debug
internal/cli/            # cobra commands + --json
```

## Testing

```bash
make test   # atau: go test ./...
make vet
```

Unit test: runtime version parsers (15 toolchain), `ss`/`lsof` parser,
process search + proteksi PID, network integration (loopback),
project detector (Go/Node/PHP), container `ps` parser, config defaults,
env sanitizer + filter.

## Roadmap

- **v0.1 (ini):** Dashboard, Processes, Ports, Runtimes, Network + CLI + JSON + doctor
- **v0.2:** Dev Tools, System, Project Inspector (sudah termasuk dasarnya)
- **v0.3:** Docker/Podman actions + logs (sudah termasuk dasarnya)
- **v0.4:** auto-refresh, process tree, relasi Port→Process→Project

## License

MIT — lihat [LICENSE](LICENSE).
