# Setup Guide for macOS / Linux

This guide helps macOS and Linux developers set up their environment to work on Nova.

## Prerequisites

### macOS

#### 1. Install Homebrew

```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
```

#### 2. Install Docker Desktop

1. Download Docker Desktop from [docker.com](https://www.docker.com/products/docker-desktop/)
2. Install and open Docker Desktop
3. Wait for it to start (icon in menu bar)

#### 3. Install VS Code

```bash
brew install --cask visual-studio-code
```

Or download from [code.visualstudio.com](https://code.visualstudio.com/)

#### 4. Install Dev Containers extension

1. Open VS Code
2. Press `Cmd+Shift+X`
3. Search "Dev Containers"
4. Click Install

### Linux (Ubuntu/Debian)

#### 1. Install Docker

```bash
sudo apt update
sudo apt install -y docker.io docker-compose
sudo usermod -aG docker $USER
```

Logout and login for group changes to take effect.

#### 2. Install VS Code

```bash
# Import Microsoft GPG key
curl -fsSL https://packages.microsoft.com/keys/microsoft.asc | gpg --dearmor -o /usr/share/keyrings/packages.microsoft.gpg

# Add repository
echo "deb [arch=amd64 signed-by=/usr/share/keyrings/packages.microsoft.gpg] https://packages.microsoft.com/repos/vscode stable main" | sudo tee /etc/apt/sources.list.d/vscode.list

# Install
sudo apt update
sudo apt install -y code
```

#### 3. Install Dev Containers extension

1. Open VS Code
2. Press `Ctrl+Shift+X`
3. Search "Dev Containers"
4. Click Install

## Clone the Project

```bash
cd ~
mkdir -p src
cd src
git clone <your-repo-url> nova
cd nova
```

## Opening the Project

### From VS Code

1. Open **VS Code**
2. Press `Cmd+O` (macOS) or `Ctrl+O` (Linux)
3. Navigate to `~/src/nova`
4. Click Open
5. VS Code will detect `.devcontainer` and show a notification:
   > "Folder contains a Dev Container configuration. Reopen in Container?"
6. Click **Reopen in Container**
7. Wait for the container to build (first time: ~5 minutes)
8. Done!

### From Terminal

```bash
cd ~/src/nova
code .
```

## First Time Setup

Once inside the container, run:

```bash
# Install frontend dependencies
cd frontend && npm install && cd ..

# Run database migrations (Go with golang-migrate)
make db:migrate

# Start development servers
make dev
```

You should see:
- Frontend: http://localhost:4200
- Backend: http://localhost:4000

## Project Commands

```bash
make dev          # Start all services
make db:migrate   # Run database migrations
make db:seed      # Run database seeds
make test         # Run tests
make lint         # Lint code
make lint:fix     # Auto-fix linting
```

## Useful VS Code Shortcuts

### macOS

| Shortcut | Action |
|----------|--------|
| `Cmd+`` ` | Open terminal |
| `Cmd+P` | Quick open file |
| `Cmd+Shift+P` | Command palette |
| `Cmd+B` | Toggle sidebar |
| `Cmd+Shift+G` | Git view |

### Linux

| Shortcut | Action |
|----------|--------|
| `Ctrl+`` ` | Open terminal |
| `Ctrl+P` | Quick open file |
| `Ctrl+Shift+P` | Command palette |
| `Ctrl+B` | Toggle sidebar |
| `Ctrl+Shift+G` | Git view |

## Troubleshooting

### "Docker daemon is not running"

**macOS**: Open Docker Desktop app and wait for it to start.

**Linux**: Run:
```bash
sudo systemctl start docker
sudo systemctl enable docker
```

### Permission denied errors (Linux)

```bash
sudo usermod -aG docker $USER
# Logout and login again
```

### Slow performance

1. Increase Docker resources:
   - Docker Desktop → Settings → Resources
   - CPU: 4 cores minimum
   - Memory: 4GB minimum

### Container build fails

Check the output for errors. Common issues:

- **Port already in use**: Stop other services using ports 4200, 4000, 5432
- **Out of disk space**: Docker Desktop → Settings → Resources → Disk image size

## Optional: Install GitHub CLI

### macOS
```bash
brew install gh
```

### Linux
```bash
sudo apt install gh
```

Login:
```bash
gh auth login
```

This enables easier interaction with GitHub from terminal.

## Getting Help

1. Check this guide again
2. Ask in the team chat
3. Open an issue in the repository
