# lele-dev — Developer Workstation Toolkit

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
go run ./cmd/lele-dev          # development
make build && ./lele-dev       # binary lokal
make install                    # go install
```

Cross compile (PRD §30-31):

```bash
make cross
# lele-dev-linux-amd64, lele-dev-linux-arm64,
# lele-dev-darwin-amd64, lele-dev-darwin-arm64
```

## TUI

```bash
lele-dev
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
| `1`-`9` | dashboard shortcut |

Screens: **Dashboard, Processes (+detail & kill), Ports (+detail, kill, open),
Runtimes (15 bahasa/toolchain), Network (interface/gateway/DNS),
Dev Tools, Containers (Docker/Podman), Projects (deteksi Go/Node/PHP/Rust/...),
System (CPU/RAM/disk + environment dengan secrets disembunyikan),
Packages (npm/pip/brew/pacman/AUR/apt/cargo/gem/go + ukuran).**

## CLI (non-interaktif, script/CI/AI-agent friendly)

```bash
lele-dev ports            # list listening ports
lele-dev ports 3000       # detail port 3000
lele-dev process          # list processes (alias: processes, ps)
lele-dev process 18231    # detail PID
lele-dev runtimes         # detect runtimes + versi + path
lele-dev network          # interfaces, gateway, DNS
lele-dev doctor           # diagnostic environment (✓/⚠/✗ + ringkasan)
lele-dev tools            # dev tools
lele-dev system           # CPU/RAM/disk/OS
lele-dev project [dir]    # detect project
lele-dev containers       # docker/podman ps
lele-dev packages         # package terinstall + ukuran (semua manager)
lele-dev packages npm     # filter satu manager: npm|pip|brew|pacman|aur|apt|cargo|gem|go
lele-dev cleanup          # kandidat dev ports
lele-dev cleanup --kill 3000,5173 --yes   # kill eksplisit (wajib --yes)
```

Semua command mendukung `--json`:

```bash
lele-dev ports --json
# [{"port":3000,"protocol":"TCP","pid":18231,"process":"node",...}]
```

### Packages

Sumber yang didukung (yang tidak terinstall otomatis di-skip):

| Manager | Sumber | Ukuran dari |
|---------|--------|-------------|
| `npm` | `npm ls -g` (global) | direktori tiap package |
| `pip` | `pip list` | direktori install (`pip show`) |
| `brew` | `brew list --versions` | Cellar per formula |
| `pacman` | `pacman -Qn` (repo) | database `pacman -Qi` |
| `aur` | `yay/paru -Qm` (fallback `pacman -Qm`) | database `pacman -Qi` |
| `apt` | `dpkg-query -W` | kolom Installed-Size |
| `cargo` | `cargo install --list` | binary di `~/.cargo/bin` |
| `gem` | `gem list` | direktori gem |
| `go` | `GOPATH/bin` | ukuran binary (versi tak terlacak) |

```bash
lele-dev packages pacman --json
# [{"manager":"pacman","name":"nodejs","version":"22.2.0-1","size":"55.1 MB","size_bytes":57776537}]
```

## Config (optional)

Tidak membutuhkan config untuk berjalan. Jika ada, dibaca dari:

- Linux: `~/.config/lele-dev/config.toml`
- macOS: `~/Library/Application Support/lele-dev/config.toml`

```toml
refresh_interval = 3
default_screen = "dashboard"
confirm_kill = true
show_system = true
```

Debug log (default silent, tidak mencampur TUI):

```bash
lele-dev --debug   # log → ~/.local/state/lele-dev/lele-dev.log
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
cmd/lele-dev/main.go
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
project detector (Go/Node/PHP), container `ps` parser, package parsers
(npm/pip/brew/pacman/dpkg/cargo/gem/go) + ukuran, config defaults,
env sanitizer + filter.

## Roadmap

- **v0.1 (ini):** Dashboard, Processes, Ports, Runtimes, Network + CLI + JSON + doctor
- **v0.2:** Dev Tools, System, Project Inspector (sudah termasuk dasarnya)
- **v0.3:** Docker/Podman actions + logs (sudah termasuk dasarnya)
- **v0.4:** auto-refresh, process tree, relasi Port→Process→Project

## License

MIT — lihat [LICENSE](LICENSE).
