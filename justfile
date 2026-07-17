# hirejack - ATS security auditing tool

# Default: show help
default:
    @just --list

# Install frontend deps
install:
    cd frontend && pnpm install

# Build frontend
frontend:
    cd frontend && pnpm build

# Build Go binary (requires frontend built first)
binary: frontend
    go build -o hirejack .

# Build everything
build: install frontend binary
    @echo "Built: ./hirejack ($(du -h hirejack | cut -f1))"

# Dev: run frontend dev server (slower, HMR)
dev-frontend:
    cd frontend && pnpm dev

# Serve: production build preview (faster, no HMR)
serve-frontend:
    cd frontend && pnpm serve

# Dev: run Go server (hot reload with air if installed)
dev:
    go run .

# Clean build artifacts
clean:
    rm -f hirejack
    rm -rf frontend/dist frontend/node_modules

# Run the server (opens browser)
run: build
    ./hirejack

# Run headless (no browser)
run-headless: build
    ./hirejack -headless

# Run with custom port
run-port port="8080":
    ./hirejack -port {{port}}

# Generate man pages (via Cobra)
man:
    ./hirejack man -d ./man
    @echo "Generated man pages in ./man/"

# View main manpage
man-view:
    ./hirejack man -d /tmp/hirejack-man && man /tmp/hirejack-man/hirejack.1

# Install binary and manpages
install-bin: build
    sudo install -m 755 hirejack /usr/local/bin/
    ./hirejack man -d /tmp/hirejack-man
    sudo install -m 644 /tmp/hirejack-man/*.1 /usr/local/share/man/man1/
    sudo mandb -q
    @echo "Installed: hirejack + man pages"

# Uninstall
uninstall:
    sudo rm -f /usr/local/bin/hirejack
    sudo rm -f /usr/local/share/man/man1/hirejack*.1
    @echo "Uninstalled"
