# Release Instructions for kUber v1.0.0

## Steps to Create GitHub Repository and Release

### 1. Create GitHub Repository

1. Go to [GitHub.com](https://github.com) and log in
2. Click the "+" icon in the top right and select "New repository"
3. Repository settings:
   - **Repository name**: `kuber`
   - **Description**: `kUber - An Uber Kubernetes Manager. A powerful, intuitive terminal-based Kubernetes cluster manager with real-time log streaming and enhanced UX.`
   - **Visibility**: Public (recommended for open source)
   - **Initialize**: Leave unchecked (we already have files)
4. Click "Create repository"

### 2. Push Code to GitHub

After creating the repository, run these commands from `/home/ar/Kuber/Kuber/`:

```bash
# Add the GitHub remote (replace anindyar with your GitHub username)
git remote add origin https://github.com/anindyar/kuber.git

# Create and switch to main branch
git branch -M main

# Push code to GitHub
git push -u origin main
```

### 3. Update Repository References

After pushing, update the installation script and README to use your actual repository:

```bash
# Replace anindyar with your actual GitHub username
sed -i 's/your-org\/kuber/anindyar\/kuber/g' install.sh README.md

# Commit the updates
git add install.sh README.md
git commit -m "docs: update repository references to actual GitHub repo"
git push
```

### 4. Create First Release

1. Go to your repository on GitHub
2. Click on "Releases" (or go to `https://github.com/anindyar/kuber/releases`)
3. Click "Create a new release"
4. Release settings:
   - **Tag version**: `v1.0.0`
   - **Release title**: `v1.0.0 - Initial Release`
   - **Description**: 
   ```markdown
   # kUber v1.0.0 - Initial Release 🚀
   
   The first stable release of kUber - An Uber Kubernetes Manager!
   
   ## ✨ Features
   - 🚀 Intuitive terminal-based Kubernetes cluster management
   - 📊 Real-time log streaming with keyword search and highlighting  
   - 🐳 Multi-container pod support with automatic detection
   - 📈 Aggregated logging for deployments and statefulsets
   - 🔍 Advanced search with real-time filtering
   - 🎯 In-terminal YAML resource editing
   - 🖥️ Direct pod shell access
   - ⚡ Optimized performance for large clusters
   
   ## 🚀 Installation
   
   ### One-line install (Linux/macOS):
   ```bash
   curl -sSL https://raw.githubusercontent.com/anindyar/kuber/main/install.sh | sh
   ```
   
   ### Manual installation:
   Download the appropriate binary for your platform from the assets below.
   
   ## 📖 Quick Start
   ```bash
   # Launch with default kubectl context
   kuber
   
   # Use specific context  
   kuber --context=my-cluster
   ```
   
   ## 📚 Documentation
   - [README](https://github.com/anindyar/kuber#readme) - Complete documentation
   - [Contributing Guide](https://github.com/anindyar/kuber/blob/main/CONTRIBUTING.md)
   
   Built with ❤️ using Go and Bubble Tea!
   ```
   - **Attach binaries**: The GitHub Action will automatically build and attach binaries when you create the tag
5. Click "Publish release"

### 5. Automated Builds (GitHub Actions)

The repository includes a GitHub Actions workflow (`.github/workflows/release.yml`) that will:
- Automatically trigger when you create a tag starting with `v` (like `v1.0.0`)
- Run tests and linting
- Build binaries for multiple platforms (Linux, macOS, Windows)
- Package them as tar.gz/zip files
- Attach them to the GitHub release
- Update the installation script with correct repository references

### 6. Test the Installation

After the release is created and binaries are built:

```bash
# Test the one-line installer
curl -sSL https://raw.githubusercontent.com/anindyar/kuber/main/install.sh | sh

# Test the binary
kuber --help
```

### 7. Share Your Release

Consider sharing on:
- Reddit (r/kubernetes, r/golang)
- Hacker News
- Twitter/X
- LinkedIn
- Dev.to
- Kubernetes community forums

## Repository Maintenance

### For future releases:
1. Make your changes and commit them
2. Create a new tag: `git tag v1.0.1`  
3. Push the tag: `git push origin v1.0.1`
4. GitHub Actions will automatically build and create the release

### Repository settings to configure:
- Enable Issues and Discussions
- Add topics: `kubernetes`, `tui`, `go`, `terminal`, `k8s`, `cli`, `kubectl`
- Set up branch protection rules for `main`
- Configure security alerts and dependency scanning

## Files Created/Modified for Release

- ✅ `README.md` - Comprehensive documentation
- ✅ `install.sh` - One-line installation script
- ✅ `LICENSE` - MIT license  
- ✅ `CONTRIBUTING.md` - Contribution guidelines
- ✅ `Makefile` - Enhanced with cross-platform builds
- ✅ `.github/workflows/release.yml` - Automated release pipeline
- ✅ `.gitignore` - Updated exclusions
- ✅ `go.mod` - Cleaned up dependencies

Your kUber project is ready for release! 🎉