# Setup Guide for Windows (WSL)

This guide helps Windows developers set up their environment to work on Nova.

## Prerequisites

### 1. Install WSL 2

Open **PowerShell as Administrator** and run:

```powershell
wsl --install
```

Restart your computer when prompted.

### 2. Install Ubuntu

After restart, WSL will prompt you to create a Ubuntu user:

```
Enter new UNIX username: dev
Enter new password: ********
```

记住 this password — you'll need it for sudo commands.

### 3. Update Ubuntu

Open Ubuntu terminal and run:

```bash
sudo apt update && sudo apt upgrade -y
```

### 4. Install Docker Desktop

1. Download Docker Desktop from [docker.com](https://www.docker.com/products/docker-desktop/)
2. Install and restart Windows
3. Open Docker Desktop
4. Go to **Settings → General** and enable:
   - ✅ Use the WSL 2 based engine
5. Go to **Settings → Resources → WSL Integration**
   - ✅ Enable integration with your Ubuntu distro

### 5. Install VS Code (Windows)

1. Download from [code.visualstudio.com](https://code.visualstudio.com/)
2. Install with default options
3. Install the **Dev Containers** extension:
   - Press `Ctrl+Shift+X`
   - Search "Dev Containers"
   - Click Install

### 6. Clone the project

In Ubuntu terminal:

```bash
cd ~
mkdir -p src
cd src
git clone <your-repo-url> nova
cd nova
```

## Opening the Project

### Method 1: From VS Code (Recommended)

1. Open **VS Code in Windows** (not in Ubuntu)
2. Press `Ctrl+O` or File → Open Folder
3. Navigate to: `\\wsl$\Ubuntu\home\<your-user>\src\nova`
4. Click Select Folder
5. VS Code will detect `.devcontainer` and show a notification:
   > "Folder contains a Dev Container configuration. Reopen in Container?"
6. Click **Reopen in Container**
7. Wait for the container to build (first time: ~5 minutes)
8. Done! You have a fully configured development environment.

### Method 2: From Ubuntu Terminal

```bash
cd ~/src/nova
code .
```

VS Code will open and prompt to reopen in container.

## First Time Setup

Once inside the container, run:

```bash
# Install dependencies
npm install

# Generate Prisma client
npm run db:generate

# Start development servers
make dev
```

You should see:
- Frontend: http://localhost:4200
- Backend: http://localhost:4000
- API Docs: http://localhost:4000/docs

## Troubleshooting

### "Docker daemon is not running"

1. Open Docker Desktop app in Windows
2. Wait for it to fully start
3. Try reopening VS Code

### "WSL integration is not enabled"

1. Open Docker Desktop
2. Settings → Resources → WSL Integration
3. Enable your Ubuntu distro
4. Apply & Restart

### Container fails to build

Check the output for errors. Common issues:

- **Port already in use**: Stop other services using ports 4200, 4000, 5432
- **Out of memory**: Increase Docker Desktop memory in Settings → Resources

### Slow performance in WSL

1. Open Docker Desktop
2. Settings → Resources
3. Increase CPU and Memory:
   - CPU: 4 cores minimum
   - Memory: 4GB minimum
   - Swap: 2GB
   - Disk image: 60GB+

### Git credential issues

```bash
git config --global credential.helper "/mnt/c/Program\ Files/Git/mingw64/bin/git-credential-manager.exe"
```

## Project Commands

```bash
make dev          # Start all services
make db:migrate   # Run database migrations
make db:studio    # Open Prisma Studio
make test         # Run tests
make lint         # Lint code
make lint:fix     # Auto-fix linting
```

## Useful VS Code Shortcuts

| Shortcut | Action |
|----------|--------|
| `Ctrl+`` ` | Open terminal |
| `Ctrl+P` | Quick open file |
| `Ctrl+Shift+P` | Command palette |
| `F1` | Same as above |
| `Ctrl+B` | Toggle sidebar |
| `Ctrl+Shift+G` | Git view |

## Getting Help

If you're stuck:

1. Check this guide again
2. Ask in the team chat
3. Open an issue in the repository
