**TL;DR:** Kita bikin `dev-utils` sebagai **TUI developer workstation toolkit** berbasis Go, target utama macOS + Linux. Fokus MVP: **Processes, Ports, Runtimes, Network**, lalu berkembang ke Docker/Podman, disk, services, environment, dan project inspector. Arsitekturnya dibuat modular supaya command OS-specific tidak berantakan.

# PRD — dev-utils

## 1. Overview

**Nama:** dev-utils
**Type:** CLI / TUI Developer Workstation Toolkit
**Language:** Go
**Platform:** macOS & Linux
**UI:** Terminal User Interface (TUI)

### Konsep

`dev-utils` adalah toolbox interaktif untuk developer yang ingin melakukan inspeksi dan operasi umum pada komputer langsung dari terminal.

Pengguna cukup menjalankan:

```bash
dev-utils
```

Kemudian muncul dashboard:

```text
╭──────────────────────────────────────────────────────╮
│ dev-utils                              v0.1.0         │
├──────────────────────────────────────────────────────┤
│                                                      │
│  1  Processes       Running processes                │
│  2  Ports           Listening ports                  │
│  3  Runtimes        Programming languages            │
│  4  Network         Network information              │
│  5  Dev Tools       Installed developer tools        │
│  6  Containers      Docker / Podman                  │
│  7  Projects        Detect current project           │
│  8  System          CPU / RAM / Disk                 │
│                                                      │
│  q  Quit                                             │
╰──────────────────────────────────────────────────────╯
```

Tujuan utamanya bukan menggantikan terminal command seperti `ps`, `lsof`, `ip`, atau `docker`.

Tujuannya adalah menyediakan **satu tempat interaktif** untuk pekerjaan developer yang sering dilakukan berulang kali.

---

# 2. Problem

Developer sering menjalankan banyak command berbeda:

```bash
lsof -i :3000
ps aux
kill 12345
which go
go version
node --version
python --version
ifconfig
ip addr
docker ps
df -h
```

Masalahnya:

- command berbeda untuk setiap kebutuhan
- output berbeda-beda
- beberapa command berbeda antara macOS dan Linux
- sulit mengetahui tool apa saja yang terinstall
- sulit menemukan process berdasarkan port
- kill process membutuhkan command tambahan
- developer harus mengingat banyak command

`dev-utils` menyatukan aktivitas tersebut dalam satu TUI.

---

# 3. Goals

## Primary Goals

1. Melihat process yang sedang berjalan.
2. Melihat port yang sedang digunakan.
3. Kill process dari TUI.
4. Mendeteksi programming language/runtime yang terinstall.
5. Menampilkan versi runtime.
6. Menampilkan informasi network.
7. Menampilkan developer tools.
8. Berjalan di macOS dan Linux.
9. Tidak membutuhkan database.
10. Tetap bisa digunakan sepenuhnya offline.

## Secondary Goals

- Docker/Podman inspection.
- Project detection.
- System information.
- Environment inspection.
- Service management.
- Search/filter.
- Export informasi ke JSON.

## Non-Goals

Untuk versi awal jangan membuat:

- package manager
- IDE
- terminal emulator
- system monitor sekompleks btop
- remote server manager
- cloud dashboard
- automatic package installation

---

# 4. Target User

Developer yang bekerja menggunakan terminal.

Contoh:

- Web developer
- Backend developer
- Mobile developer
- DevOps beginner
- Student programming
- Linux/macOS power user

---

# 5. Platform Support

## Linux

Target:

- Ubuntu
- Debian
- Fedora
- Arch Linux
- distro Linux lainnya yang mengikuti standar POSIX/Linux

## macOS

Target:

- Apple Silicon
- Intel Mac

Minimal:

```text
macOS 13+
```

---

# 6. Technology Stack

## Core

```text
Go
```

Target:

```text
Go 1.24+
```

## TUI

Gunakan:

```text
Bubble Tea
```

Ekosistem:

```text
Bubble Tea
Bubbles
Lip Gloss
```

## CLI

Gunakan:

```text
Cobra
```

## System information

Pertimbangkan:

```text
gopsutil
```

untuk informasi system/process lintas platform.

Namun jangan bergantung pada library untuk semua hal.

Gunakan native OS command/API ketika lebih reliable.

---

# 7. Architecture

Gunakan arsitektur modular.

```text
dev-utils/
│
├── cmd/
│   └── dev-utils/
│       └── main.go
│
├── internal/
│   ├── app/
│   │   ├── app.go
│   │   └── router.go
│   │
│   ├── ui/
│   │   ├── dashboard/
│   │   ├── processes/
│   │   ├── ports/
│   │   ├── runtimes/
│   │   ├── network/
│   │   ├── tools/
│   │   ├── containers/
│   │   ├── projects/
│   │   └── system/
│   │
│   ├── services/
│   │   ├── process/
│   │   ├── port/
│   │   ├── runtime/
│   │   ├── network/
│   │   ├── tools/
│   │   ├── container/
│   │   └── project/
│   │
│   ├── platform/
│   │   ├── linux/
│   │   └── darwin/
│   │
│   ├── models/
│   │   ├── process.go
│   │   ├── port.go
│   │   ├── runtime.go
│   │   └── network.go
│   │
│   └── config/
│       └── config.go
│
├── pkg/
│
├── assets/
│
├── go.mod
├── go.sum
├── Makefile
├── README.md
└── LICENSE
```

Prinsip:

```text
UI
 ↓
Service
 ↓
Platform abstraction
 ↓
OS
```

Jangan membuat UI langsung menjalankan:

```go
exec.Command(...)
```

UI hanya meminta:

```go
processService.List()
```

---

# 8. Core Domain Models

## Process

```go
type Process struct {
    PID        int
    Name       string
    Command    string
    User       string
    CPUPercent float64
    MemoryMB   float64
}
```

## Port

```go
type Port struct {
    Port        int
    Protocol    string
    Address     string
    PID         int
    ProcessName string
}
```

## Runtime

```go
type Runtime struct {
    Name    string
    Command string
    Version string
    Path    string
    Installed bool
}
```

## Network Interface

```go
type NetworkInterface struct {
    Name      string
    Address   string
    Type      string
    Status    string
}
```

---

# 9. Main Dashboard

Saat menjalankan:

```bash
dev-utils
```

Tampilkan:

```text
╭────────────────────────────────────────────────────╮
│ dev-utils                                            │
│ Developer Workstation Toolkit                       │
├────────────────────────────────────────────────────┤
│                                                    │
│  Processes       124 running                       │
│  Ports            7 listening                      │
│  Runtimes         8 detected                       │
│  Network          2 interfaces                     │
│  Dev Tools       12 detected                       │
│                                                    │
│  [1] Processes                                     │
│  [2] Ports                                         │
│  [3] Runtimes                                      │
│  [4] Network                                       │
│  [5] Dev Tools                                     │
│  [6] Containers                                    │
│  [7] Projects                                      │
│  [8] System                                        │
│                                                    │
│  q Quit                                             │
╰────────────────────────────────────────────────────╯
```

Keyboard navigation:

```text
↑ ↓       Navigate
Enter     Open
Esc       Back
q         Quit
r         Refresh
/         Search
```

---

# 10. Process Manager

Menu:

```text
PROCESS MANAGER

PID      NAME          CPU       MEMORY
------------------------------------------------
18231    node          2.4%      342 MB
19342    go            1.1%      120 MB
20182    postgres      0.4%      212 MB
22102    chrome        8.2%      1.2 GB
```

Actions:

```text
Enter   Details
k       Kill
r       Refresh
/       Search
```

Detail:

```text
PROCESS

PID:       18231
Name:      node
Command:   node server.js
User:      pojan
CPU:       2.4%
Memory:    342 MB

Ports:
  :3000
  :5173

[k] Kill
```

Kill harus menggunakan confirmation:

```text
Kill process 18231 (node)?

[y] Yes
[n] No
```

Default action:

```text
SIGTERM
```

Jangan langsung menggunakan force kill.

Force kill hanya sebagai fallback.

---

# 11. Port Manager

Menu:

```text
PORTS

PORT     PROTOCOL    PID      PROCESS
------------------------------------------------
3000     TCP         18231    node
5173     TCP         19342    node
5432     TCP         912      postgres
6379     TCP         1021     redis
8080     TCP         20182    go
```

Actions:

```text
Enter   Details
k       Kill process
c       Copy port
r       Refresh
/       Search
```

Detail:

```text
PORT 3000

Address:       127.0.0.1
Protocol:      TCP
PID:           18231
Process:       node

Command:
node server.js

[k] Kill process
```

Tambahkan shortcut:

```text
o
```

untuk membuka:

```text
http://localhost:3000
```

Jika port menjalankan HTTP server.

---

# 12. Runtime Detector

Runtime yang didukung:

```text
Go
Node.js
Bun
Deno
Python
PHP
Ruby
Rust
Java
Kotlin
Dart
Flutter
Swift
GCC
Clang
```

Output:

```text
RUNTIMES

NAME          VERSION       PATH
------------------------------------------------
Go            1.25.1        /usr/bin/go
Node.js       24.8.0        /usr/bin/node
Bun           1.2.21        ~/.bun/bin/bun
Python        3.13.7        /usr/bin/python
PHP           8.4.15        /usr/bin/php
Rust          1.89.0        ~/.cargo/bin/rustc
Dart          3.x            ~/.fvm/...
Flutter       3.x            ~/.fvm/...
Java          25             /usr/bin/java
```

Setiap runtime memiliki detector:

```go
type RuntimeDetector interface {
    Detect() (*Runtime, error)
}
```

Contoh:

```go
func DetectGo() (*Runtime, error) {
    // find go
    // execute go version
    // return Runtime
}
```

Penting: jangan hanya mengecek environment variable.

Gunakan:

```bash
exec.LookPath()
```

kemudian:

```go
exec.Command(...)
```

untuk mendapatkan versi sebenarnya.

---

# 13. Network

Tampilkan:

```text
NETWORK

Interfaces

NAME       ADDRESS          STATUS
---------------------------------------
en0        192.168.1.20     UP
lo0        127.0.0.1        UP

Gateway
192.168.1.1

DNS
192.168.1.1
1.1.1.1
```

Linux interface:

```text
eth0
wlan0
docker0
lo
```

macOS:

```text
en0
en1
lo0
bridge0
```

Jangan mengasumsikan nama interface.

Ambil dari system API.

---

# 14. Developer Tools

Detector:

```text
Git
Docker
Podman
Make
CMake
NPM
PNPM
Yarn
Composer
Cargo
ADB
Gradle
Android SDK
Flutter
```

Output:

```text
DEV TOOLS

TOOL          VERSION       STATUS
----------------------------------------
Git           2.51.0        ✓
Docker        29.0.0        ✓
Podman        6.0.1         ✓
pnpm          10.x           ✓
Composer      2.x            ✓
ADB           36.x           ✓
Gradle        8.x            ✓
```

---

# 15. Container Manager

Support:

```text
Docker
Podman
```

Dashboard:

```text
CONTAINERS

NAME          IMAGE              STATUS       PORTS
------------------------------------------------------
postgres      postgres:17        running      5432
redis         redis:8            running      6379
school-api    school-api         running      8080
```

Actions:

```text
s       Start
t       Stop
r       Restart
l       Logs
k       Kill
```

Jika Docker/Podman tidak terinstall:

```text
Docker: Not installed
Podman: Installed

Press Enter for Podman
```

Jangan menganggap Docker harus ada.

---

# 16. Project Inspector

Jika menjalankan:

```bash
dev-utils
```

di dalam project:

```text
~/Projects/school-api
```

deteksi:

```text
PROJECT

Name:
school-api

Detected:

✓ Go
✓ Docker
✓ PostgreSQL

Files:

go.mod
Dockerfile
compose.yaml
.env
README.md
```

Jika:

```text
package.json
```

deteksi:

```text
Node.js project
Package manager: pnpm
Framework: Next.js
```

Jika:

```text
go.mod
```

deteksi:

```text
Go project
```

Jika:

```text
composer.json
```

deteksi:

```text
PHP project
```

---

# 17. System

Dashboard sederhana:

```text
SYSTEM

OS
Linux

Kernel
6.x

Architecture
x86_64

CPU
8 cores

Memory
16 GB
Used: 7.2 GB

Disk
512 GB
Used: 182 GB

Uptime
3 days
```

macOS:

```text
OS
macOS

Architecture
arm64
```

---

# 18. Search

Semua list utama harus mendukung:

```text
/
```

Contoh:

```text
/ node
```

hasil:

```text
18231 node
19342 node
20182 node
```

Untuk port:

```text
/ 3000
```

---

# 19. Refresh

Gunakan:

```text
r
```

Untuk data dynamic seperti:

- processes
- ports
- containers
- system
- network

Jangan refresh seluruh aplikasi secara agresif.

Default:

```text
manual refresh
```

Kemudian versi berikutnya bisa memiliki auto-refresh:

```text
1s
3s
5s
10s
```

---

# 20. CLI Mode

Walaupun aplikasi utama adalah TUI, sediakan CLI commands.

Contoh:

```bash
dev-utils
```

TUI.

```bash
dev-utils ports
```

List ports.

```bash
dev-utils ports 3000
```

Detail port 3000.

```bash
dev-utils process
```

List process.

```bash
dev-utils runtimes
```

List runtimes.

```bash
dev-utils network
```

Network information.

```bash
dev-utils doctor
```

Diagnostic developer environment.

Output:

```text
DEV-UTILS DOCTOR

✓ Go
✓ Git
✓ Node.js
✓ Docker
✓ Podman
⚠ Java
✗ Android SDK

5 OK
1 Warning
1 Missing
```

---

# 21. JSON Output

Semua command CLI harus dapat:

```bash
dev-utils ports --json
```

Output:

```json
[
  {
    "port": 3000,
    "protocol": "tcp",
    "pid": 18231,
    "process": "node"
  }
]
```

Tujuannya supaya bisa digunakan oleh:

- shell script
- CI
- AI agent
- automation
- developer tooling lain

---

# 22. Configuration

Config optional.

Linux:

```text
~/.config/dev-utils/config.toml
```

macOS:

```text
~/Library/Application Support/dev-utils/config.toml
```

Config:

```toml
refresh_interval = 3
default_screen = "dashboard"
confirm_kill = true
show_system = true
```

Jangan membutuhkan config untuk menjalankan aplikasi.

---

# 23. Security

Ini penting karena aplikasi dapat melakukan operasi terhadap process.

Rules:

1. Jangan menjalankan aplikasi sebagai root.
2. Jangan meminta sudo saat startup.
3. Kill hanya process yang user punya permission.
4. Process system harus dilindungi.
5. Force kill membutuhkan confirmation kedua.
6. Jangan menjalankan arbitrary shell command dari input user.
7. Jangan mengekspos environment variable secara default.

Untuk environment:

```text
Environment

PATH       /usr/local/bin:...
SHELL      /bin/zsh
EDITOR     vim
```

Jangan tampilkan:

```text
API_KEY
TOKEN
PASSWORD
SECRET
```

---

# 24. Cross-Platform Strategy

Gunakan Go build tags jika diperlukan.

```text
platform/
├── linux/
│   ├── process.go
│   └── network.go
│
└── darwin/
    ├── process.go
    └── network.go
```

Interface:

```go
type ProcessService interface {
    List() ([]Process, error)
    Kill(pid int) error
}
```

Implementasi:

```text
LinuxProcessService
DarwinProcessService
```

UI tidak mengetahui perbedaan OS.

---

# 25. Error Handling

Jangan menampilkan panic.

Bad:

```text
panic: exit status 1
```

Good:

```text
Unable to inspect port 3000.

Reason:
Permission denied.

Press Enter to continue.
```

Jika tool tidak tersedia:

```text
Podman

Not installed.

This feature requires Podman.
```

---

# 26. TUI Design

Style:

```text
minimal
developer-oriented
dense
keyboard-first
```

Hindari:

- terlalu banyak warna
- animasi berlebihan
- giant cards
- dashboard seperti SaaS website

Prioritaskan informasi.

Contoh:

```text
PORTS
──────────────────────────────────────────────
3000   node       18231
5173   node       19342
5432   postgres   912
6379   redis      1021
8080   go         20182

[k] kill   [/] search   [r] refresh   [esc] back
```

---

# 27. Keyboard UX

Global:

```text
↑ ↓       Navigate
Enter     Select
Esc       Back
q         Quit
r         Refresh
/         Search
?         Help
```

Process:

```text
k         Kill
```

Port:

```text
k         Kill process
o         Open localhost
```

Container:

```text
s         Start
t         Stop
r         Restart
l         Logs
```

---

# 28. Logging

Default:

```text
silent
```

Debug:

```bash
dev-utils --debug
```

Log file:

```text
~/.local/state/dev-utils/
```

atau lokasi platform-specific yang sesuai.

Jangan mencampurkan debug log dengan TUI.

---

# 29. Testing

## Unit Test

Test:

```text
runtime parser
version parser
process model
port parser
network parser
project detector
```

Contoh:

```go
func TestParseGoVersion(t *testing.T) {
    input := "go version go1.25.1 linux/amd64"

    got := ParseVersion(input)

    if got != "1.25.1" {
        t.Fatal(...)
    }
}
```

## Integration Test

Test:

```text
process listing
port detection
runtime detection
network detection
```

## Platform Test

CI matrix:

```text
Ubuntu
macOS Intel
macOS Apple Silicon
```

---

# 30. Build

Development:

```bash
go run ./cmd/dev-utils
```

Build:

```bash
go build -o dev-utils ./cmd/dev-utils
```

Install:

```bash
go install ./cmd/dev-utils
```

Cross compile:

```bash
GOOS=linux GOARCH=amd64 go build
```

```bash
GOOS=darwin GOARCH=arm64 go build
```

---

# 31. Release

Binary:

```text
dev-utils-linux-amd64
dev-utils-linux-arm64
dev-utils-darwin-amd64
dev-utils-darwin-arm64
```

Future package manager support:

```text
Homebrew
AUR
Nix
.deb
.rpm
```

Tidak wajib untuk MVP.

---

# 32. MVP Scope

Version:

```text
v0.1.0
```

Hanya:

```text
Dashboard
Processes
Ports
Runtimes
Network
```

Required actions:

```text
Process:
- list
- search
- detail
- kill

Ports:
- list
- search
- detail
- kill
- open localhost

Runtimes:
- detect
- version
- path

Network:
- interface
- IP
- gateway
- DNS
```

CLI:

```bash
dev-utils
dev-utils ports
dev-utils process
dev-utils runtimes
dev-utils network
dev-utils doctor
```

---

# 33. v0.2

Tambahkan:

```text
Dev Tools
System
Project Inspector
```

---

# 34. v0.3

Tambahkan:

```text
Docker
Podman
Container logs
Container restart
```

---

# 35. v0.4

Tambahkan:

```text
Auto refresh
Process tree
Port → Process → Project relationship
```

Contoh:

```text
PORT 3000
  ↓
PID 18231
  ↓
node
  ↓
~/Projects/website
  ↓
Next.js
```

Ini bisa menjadi salah satu fitur paling keren dari aplikasi.

---

# 36. Future Feature — Dev Environment Doctor

Command:

```bash
dev-utils doctor
```

Analisis:

```text
DEV ENVIRONMENT

Go
✓ installed
✓ PATH configured

Node
✓ installed

PHP
✓ installed

Docker
✓ installed
⚠ daemon not running

Android
✓ SDK
⚠ adb unavailable

Git
✓ configured

Overall:
8 OK
2 Warning
0 Error
```

---

# 37. Future Feature — Port Cleanup

Command:

```bash
dev-utils cleanup
```

TUI:

```text
DEV PORT CLEANUP

3000   node      Next.js
5173   node      Vite
8000   php       Laravel
8080   go        API

[space] Select
[enter] Kill selected
```

Confirmation:

```text
Kill 3 development processes?

3000 node
5173 node
8000 php

[y] Yes
[n] Cancel
```

Jangan pernah melakukan cleanup otomatis tanpa confirmation.

---

# 38. Future Feature — Project Context

Saat berada di:

```bash
~/Projects/my-api
```

Dashboard dapat berubah menjadi:

```text
MY-API

Go 1.25.1
Git branch: main

Services

API       :8080   ✓
Postgres  :5432   ✓
Redis     :6379   ✓

Project tools

[1] Processes
[2] Ports
[3] Containers
[4] Logs
[5] Doctor
```

Ini membuat `dev-utils` terasa seperti **development cockpit**, bukan sekadar system monitor.

---

# 39. Performance Requirements

Startup:

```text
< 300ms
```

TUI harus tetap responsive.

Jangan menjalankan semua detector secara sequential.

Gunakan goroutine untuk operasi independen:

```text
Processes ─┐
Ports ─────┤
Runtimes ──┼──→ concurrent scan
Network ───┤
Tools ─────┘
```

Namun hindari goroutine berlebihan.

---

# 40. Core Principle

Arsitektur harus mengikuti prinsip:

```text
Simple first.

Native when reliable.

Library when it removes complexity.

Platform-specific code stays isolated.

UI never owns system logic.
```

Target akhirnya bukan membuat tool dengan 100 fitur.

Targetnya adalah:

> **“Kalau saya sedang ngoding dan butuh tahu apa yang terjadi di mesin saya, saya buka dev-utils.”**

---

# 41. Definition of Done — v0.1

Project dianggap MVP selesai jika:

- [ ] `dev-utils` bisa dijalankan dari terminal.
- [ ] TUI dashboard tampil.
- [ ] Process dapat dilihat.
- [ ] Process dapat dicari.
- [ ] Process detail dapat dilihat.
- [ ] Process dapat di-kill dengan confirmation.
- [ ] Port listening dapat dilihat.
- [ ] Port dapat dicari.
- [ ] Port → PID → process dapat diketahui.
- [ ] Process pada port dapat di-kill.
- [ ] Runtime programming language dapat dideteksi.
- [ ] Version runtime dapat ditampilkan.
- [ ] Path binary dapat ditampilkan.
- [ ] Network interface dapat ditampilkan.
- [ ] Local IP dapat ditampilkan.
- [ ] Berjalan di Linux.
- [ ] Berjalan di macOS Intel.
- [ ] Berjalan di macOS Apple Silicon.
- [ ] Tidak membutuhkan root untuk fitur normal.
- [ ] Tidak membutuhkan internet.
- [ ] Unit test tersedia.
- [ ] README tersedia.
- [ ] Binary dapat dibuat untuk Linux/macOS.

---

# 42. Recommended Development Order

Jangan langsung membuat seluruh menu.

Urutan implementasi:

```text
Phase 1
│
├── Go project setup
├── Cobra
└── Bubble Tea
        ↓
Phase 2
│
├── Process service
├── Process TUI
└── Kill process
        ↓
Phase 3
│
├── Port service
├── Port → PID
└── Kill from port
        ↓
Phase 4
│
├── Runtime detector
├── Version parser
└── Runtime TUI
        ↓
Phase 5
│
├── Network service
└── Network TUI
        ↓
Phase 6
│
└── Dashboard
        ↓
Phase 7
│
├── CLI commands
├── JSON output
└── doctor
```

Dengan urutan ini, setiap fase menghasilkan sesuatu yang bisa langsung dipakai.

---

# 43. Final Product Vision

```text
$ dev-utils

╭─────────────────────────────────────────────────────╮
│ dev-utils                         Developer Cockpit │
├─────────────────────────────────────────────────────┤
│                                                     │
│  SYSTEM                                             │
│  macOS arm64     CPU 12%     RAM 7.2/16 GB          │
│                                                     │
│  DEVELOPMENT                                        │
│  Go 1.25.1       Node 24.8      PHP 8.4             │
│  Git ✓           Podman ✓       Docker ✓            │
│                                                     │
│  SERVICES                                           │
│  :3000  Next.js        PID 18231       ●            │
│  :5173  Vite           PID 19342       ●            │
│  :8080  Go API         PID 20182       ●            │
│  :5432  PostgreSQL     PID 912         ●            │
│                                                     │
│  NETWORK                                            │
│  en0     192.168.1.20                               │
│                                                     │
├─────────────────────────────────────────────────────┤
│ [p] Processes  [t] Ports  [r] Runtimes  [n] Network │
│ [d] Dev Tools  [c] Containers  [s] System          │
│                                                     │
│ q Quit   / Search   ? Help                          │
╰─────────────────────────────────────────────────────╯
```

**Positioning:** `dev-utils` = **Developer Workstation Cockpit for macOS & Linux.**

Bukan replacement untuk terminal. Bukan replacement untuk `btop`. Bukan replacement untuk Docker CLI.

Ia adalah **layer interaktif di atas berbagai system/developer utilities** sehingga pekerjaan diagnostik sehari-hari menjadi satu workflow.
