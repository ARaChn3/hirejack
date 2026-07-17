# hirejack

ATS (Applicant Tracking System) security auditing tool. Tests parsing vulnerabilities, injection vectors, prompt injection for LLM-based screening, and ML poisoning through weaponized resumes.

```
 ██╗  ██╗██╗██████╗ ███████╗     ██╗ █████╗  ██████╗██╗  ██╗
 ██║  ██║██║██╔══██╗██╔════╝     ██║██╔══██╗██╔════╝██║ ██╔╝
 ███████║██║██████╔╝█████╗       ██║███████║██║     █████╔╝
 ██╔══██║██║██╔══██╗██╔══╝  ██   ██║██╔══██║██║     ██╔═██╗
 ██║  ██║██║██║  ██║███████╗╚█████╔╝██║  ██║╚██████╗██║  ██╗
 ╚═╝  ╚═╝╚═╝╚═╝  ╚═╝╚══════╝ ╚════╝ ╚═╝  ╚═╝ ╚═════╝╚═╝  ╚═╝
```

## Features

- **84 payloads** across 6 categories (SQLi, XXE, DoS, Integrity, ML Poisoning, LLM Injection)
- **5 stealth levels** from raw payloads to polyglot/orphan object techniques
- **Canary tokens** to detect when ATS systems parse injected payloads
- **Web UI** for visual payload selection and PDF injection
- **CLI** for scripted/automated testing
- **Auto-generated man pages** via Cobra

## Requirements

- Go 1.21+
- Node.js 18+ / pnpm
- [SecLists](https://github.com/danielmiessler/SecLists) (optional, for extended payload generation)

### Installing SecLists

```bash
# Arch Linux
yay -S seclists

# Debian/Ubuntu
sudo apt install seclists

# macOS
brew install seclists

# Manual
git clone https://github.com/danielmiessler/SecLists /usr/share/seclists
```

## Installation

```bash
# Clone
git clone https://github.com/NovusEdge/hirejack
cd hirejack

# Build (requires just)
just build

# Or manually
cd frontend && pnpm install && pnpm build && cd ..
go build -o hirejack .

# Install system-wide (optional)
just install-bin
```

## Usage

### Web UI

```bash
hirejack                    # Start server, open browser
hirejack -p 3000            # Custom port
hirejack --headless         # Don't open browser
```

Then visit `http://localhost:8080`:
1. Upload a PDF resume
2. Select payloads from the matrix
3. Choose stealth level
4. Download poisoned PDF

### CLI

```bash
# List all payloads
hirejack list

# List payloads in a category
hirejack list -c llm_injection

# Inject payloads into a PDF
hirejack inject resume.pdf -P sqli.name_inject,xxe.dtd_external -o poisoned.pdf

# Use a profile
hirejack inject resume.pdf --profile full_audit --stealth 3

# With canary token
hirejack inject resume.pdf -P llm_injection.ignore_resume --canary
```

### Man Pages

```bash
# Generate man pages
hirejack man -d ./man

# View
man ./man/hirejack.1
```

## Payload Categories

| Category | Count | Description |
|----------|-------|-------------|
| `sqli` | 8 | SQL injection (MySQL, PostgreSQL, MSSQL) |
| `xxe` | 14 | XML External Entity / SSRF / CVE-2025-66516 |
| `dos` | 13 | Denial of service (XML bomb, ReDoS, compression bombs) |
| `integrity` | 13 | Parser confusion (homoglyphs, RTL, null bytes) |
| `ml_poison` | 11 | ML poisoning (invisible keywords, embedding anchors) |
| `llm_injection` | 25 | Prompt injection for AI resume screening |

## Stealth Levels

| Level | Technique | Detection Risk |
|-------|-----------|----------------|
| L1 | Raw payloads in metadata | High |
| L2 | Hex encoding | Medium |
| L3 | Split across fields | Medium-Low |
| L4 | Orphan PDF objects | Low |
| L5 | Polyglot + all techniques | Minimal |

## Profiles

- `full_audit` - All 84 payloads
- `sqli_focus` - SQL injection only
- `xxe_focus` - XXE attacks only
- `llm_injection_focus` - LLM prompt injection
- `stealth_recon` - High stealth payloads (L4+)

## API

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/payloads` | GET | List all payload categories |
| `/api/inject` | POST | Inject payloads into PDF |
| `/api/generate` | POST | Generate new CV with payloads |
| `/api/callbacks` | GET | View canary callbacks |
| `/docs` | GET | Payload documentation |

### Inject Example

```bash
curl -X POST http://localhost:8080/api/inject \
  -H "Content-Type: application/json" \
  -d '{
    "pdf": "<base64-encoded-pdf>",
    "payload_ids": ["sqli.name_inject", "xxe.dtd_external"],
    "stealth": 3,
    "add_canary": true
  }'
```

## Adding Payloads

Edit `internal/payloads/registry.go`:

```go
Payload{
    ID:          "xxe_custom",
    Name:        "Custom XXE",
    Description: "Reads /etc/passwd",
    Impact:      "critical",
    Stealth:     []int{1, 2, 3},
    Locations:   []string{"xmp", "metadata"},
    Template:    `<!DOCTYPE foo [<!ENTITY x SYSTEM "file:///etc/passwd">]>&x;`,
}
```

Template variables:
- `{{.CanaryURL}}` - Callback URL for canary tokens
- `{{.RandomID}}` - Random identifier
- `{{.Timestamp}}` - Current timestamp

## Stack

- **Backend**: Go + Fiber v2 + pdfcpu + Cobra
- **Frontend**: React 19 + Vite + Tailwind 4 + shadcn/ui
- **Payloads**: Based on SecLists + custom research

## Justfile Commands

```bash
just              # Show all commands
just build        # Build everything
just run          # Build and run
just dev          # Run Go server (dev mode)
just dev-frontend # Run Vite dev server
just man          # Generate man pages
just install-bin  # Install to /usr/local/bin
```

## Security Notice

This tool is for **authorized security testing only**. Using it against systems without permission is illegal. The authors are not responsible for misuse.

## License

MIT
