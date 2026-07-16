# hirejack - ATS security auditing tool

# Default: build everything
default: build

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

# Dev: run frontend dev server
dev-frontend:
    cd frontend && pnpm dev

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
