# kUber - An Uber Kubernetes Manager

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Go Version](https://img.shields.io/badge/go-1.24+-blue.svg)
![Release](https://img.shields.io/github/v/release/anindyar/kuber?include_prereleases)

A powerful, intuitive terminal-based Kubernetes cluster manager built with Go and Bubble Tea. kUber provides an enhanced user experience for managing Kubernetes resources with real-time log streaming, multi-container support, and aggregated logging for deployments.

## 🎯 Two Versions Available

This repository provides two tools optimized for different use cases:

### 📊 **kTop** - Lightweight Monitoring (Read-Only)
- **Perfect for**: Production monitoring, operations teams, security-conscious environments
- **Features**: Full dashboard, logs viewing with search/follow, shell access, resource inspection
- **Security**: Read-only access, no resource modifications
- **Memory**: 15-30MB footprint
- **Documentation**: [kTop README](cmd/ktop/README.md)

### ⚡ **kUber** - Full-Featured Manager  
- **Perfect for**: Development, administration, full cluster management
- **Features**: Everything in kTop + resource editing, YAML editing
- **Security**: Full cluster access with modification capabilities  
- **Memory**: 20-50MB footprint
- **Documentation**: Continue reading below

---

## ✨ Features

- 🚀 **Intuitive Terminal UI** - Clean, responsive interface built with Bubble Tea
- 📊 **Resource Management** - Browse, view, and edit all Kubernetes resources
- 🔄 **Real-time Log Streaming** - Live log following with keyword search and highlighting
- 🐳 **Multi-container Support** - Automatic container detection and selection
- 📈 **Aggregated Logging** - Stream logs from multiple pods in deployments/statefulsets  
- 🔍 **Advanced Search** - Real-time keyword filtering with persistent search during follow mode
- 🎯 **Resource Editing** - In-terminal YAML editor with validation
- 🖥️ **Shell Access** - Direct pod shell access from the interface
- 📊 **Enhanced Dashboard** - Performance monitor with cluster metrics and resource utilization
- 🖥️ **Multi-Node Monitoring** - Real-time resource pressure across all cluster nodes
- 📜 **Cluster Log Viewer** - Centralized log streaming from system namespaces with search/follow
- ⚡ **High Performance** - Optimized for large clusters with caching and efficient streaming

## 🚀 Installation

### 📦 Independent Installation

Choose the version that best fits your needs:

#### Install kUber (Full-Featured Manager)

```bash
# One-line installation
curl -sSL https://raw.githubusercontent.com/anindyar/kuber/main/install.sh | sh -s kuber

# Or download manually
wget https://github.com/anindyar/kuber/releases/latest/download/kuber-linux-amd64.tar.gz
tar -xzf kuber-linux-amd64.tar.gz
sudo mv kuber /usr/local/bin/
chmod +x /usr/local/bin/kuber
```

#### Install kTop (Lightweight Monitoring)

```bash
# One-line installation (Recommended)
curl -fsSL https://raw.githubusercontent.com/anindyar/kuber/main/install-ktop.sh | bash

# Alternative download
wget https://raw.githubusercontent.com/anindyar/kuber/main/install-ktop.sh
chmod +x install-ktop.sh
./install-ktop.sh
```

#### Install Both Tools

```bash
# Install both kUber and kTop
curl -sSL https://raw.githubusercontent.com/anindyar/kuber/main/install.sh | sh

# Or build both from source
git clone https://github.com/anindyar/kuber.git
cd kuber
make build-all
sudo make install
```

### 🏗️ Build from Source

**Prerequisites:**
- Go 1.24 or later
- kubectl configured with valid cluster access

```bash
# Clone the repository
git clone https://github.com/anindyar/kuber.git
cd kuber

# Build specific version
make build-kuber  # Full-featured kUber
make build-ktop   # Lightweight kTop

# Or build both at once
make build-all

# Install to system
sudo make install  # Installs both kuber and ktop
```

## 🎯 Quick Start

### kUber (Full-Featured)
```bash
# Launch with default kubectl context
kuber

# Use specific context
kuber --context=my-cluster

# Use custom kubeconfig
kuber --kubeconfig=/path/to/config
```

### kTop (Monitoring Only)
```bash
# Launch lightweight monitoring tool
ktop

# Use specific context
ktop --context=my-cluster

# Use custom kubeconfig
ktop --kubeconfig=/path/to/config
```

### 🎮 Basic Controls

#### Common Controls (Both kUber & kTop)
| Key | Action |
|-----|--------|
| `↑/↓` | Navigate lists |
| `Enter` | Select/view details |
| `Tab` | Switch between panels |
| `c` | View cluster logs |
| `r` | Refresh current view |
| `/` | Search/filter logs |
| `l` | View pod logs |
| `f` | Toggle log follow mode |
| `s` | Open pod shell |
| `d` | Describe resource |
| `Esc` | Go back/cancel |
| `q` | Quit |

#### kUber-Only Controls (Full Version)
| Key | Action |
|-----|--------|
| `e` | **Edit resource (NEW!)** |
| `h` | Show help |

## 📖 Usage Examples

### Enhanced Dashboard
The main dashboard now shows comprehensive cluster information:
- **📊 Cluster Performance Monitor**: Real-time CPU, memory, and storage utilization
- **📈 Resource Pressure Metrics**: Multi-node resource pressure analysis  
- **🖥️ Per-Node Status**: Individual node health with resource scores
- **📊 Workload Counts**: Live counts of deployments, pods, services, etc.

### Cluster Log Monitoring
1. From the main dashboard, press `c` to access cluster logs
2. View aggregated logs from system namespaces (kube-system, default, cattle-system)
3. Use `/` to search across all cluster logs
4. Press `f` to enable live streaming mode
5. Press `r` to manually refresh log content

### Viewing Pod Logs
1. Navigate to **Pods** in the sidebar
2. Select a pod and press `l`
3. Press `f` to enable real-time streaming
4. Use `/` to search and highlight specific terms
5. For multi-container pods, kUber automatically selects the first container

### Aggregated Logging (Deployments/StatefulSets)
1. Navigate to **Deployments** or **StatefulSets**
2. Select a resource and press `l`
3. Press `f` to stream logs from all pods
4. Each log line is prefixed with `[pod-name]` for identification
5. Search functionality works across all pod logs

### Resource Editing 🆕
The new YAML editor provides full resource editing capabilities:

1. Navigate to any **Resource** (ConfigMaps, Deployments, Services, etc.)
2. Select a resource and press `e` to open the YAML editor
3. Edit the YAML using familiar vim-like controls:
   - **Ctrl+S**: Save changes to cluster
   - **Ctrl+Z**: Undo changes (revert to original)
   - **Esc**: Cancel editing (warns if unsaved changes)
4. Real-time validation and error reporting
5. Changes are applied directly using `kubectl apply`

#### Supported Resources for Editing:
- 📄 **ConfigMaps & Secrets** - Configuration management
- 🚀 **Deployments & StatefulSets** - Workload updates  
- 🌐 **Services & Ingresses** - Network configuration
- 📊 **PersistentVolumes & PVCs** - Storage management
- 🔧 **All other Kubernetes resources** - Full API support

### Shell Access
1. Navigate to a running pod
2. Press `s` to open an interactive shell
3. Execute commands as needed
4. Type `exit` or press `Ctrl+D` to close

## 🏗️ Architecture

kUber follows a library-first architecture with four core components:

```
├── kubernetes-client/    # K8s API communication
├── tui-components/      # Reusable UI widgets  
├── resource-manager/    # Resource caching & updates
└── metrics-collector/   # Performance monitoring
```

### Key Technologies
- **[Bubble Tea](https://github.com/charmbracelet/bubbletea)** - Modern TUI framework
- **[client-go](https://github.com/kubernetes/client-go)** - Official Kubernetes client
- **[Lipgloss](https://github.com/charmbracelet/lipgloss)** - Terminal styling
- **kubectl** - Fallback for log operations

## 🔧 Configuration

Create `~/.kuber/config.yaml` for custom settings:

```yaml
# UI Settings
ui:
  theme: "dark"           # "dark" or "light"  
  showTimestamps: true    # Show log timestamps
  autoRefresh: 30         # Auto-refresh interval (seconds)

# Custom Keybindings
keybindings:
  quit: "q"
  help: "h"
  logs: "l"
  shell: "s"
  edit: "e"
  follow: "f"
  search: "/"

# Cluster Profiles
clusters:
  - name: "production"
    context: "prod-cluster"
    namespace: "default"
    readOnly: true
    
  - name: "development"
    context: "dev-cluster" 
    namespace: "dev"
    readOnly: false
```

## 📊 Performance

- **Memory Usage**: ~20-50MB idle
- **Update Frequency**: 5 seconds (configurable)  
- **Log Buffer**: 1000 lines default
- **Navigation**: <100ms response time
- **Cluster Support**: Tested with 100+ node clusters

## 🛠️ Development

### Prerequisites
- Go 1.24+
- kubectl with cluster access
- Make

### Development Setup

```bash
# Clone the repository
git clone https://github.com/anindyar/kuber.git
cd kuber

# Install dependencies
go mod tidy

# Run tests
make test

# Build development version
make build-dev

# Run with development flags
./kuber --debug
```

### Project Structure

```
.
├── cmd/kuber/           # Main application
├── src/
│   ├── libraries/       # Core libraries
│   │   ├── kubernetes-client/
│   │   ├── tui-components/
│   │   ├── resource-manager/
│   │   └── metrics-collector/
│   └── models/          # Data models
├── tests/               # Test suites
├── scripts/             # Build and install scripts
└── specs/               # Design specifications
```

### Testing

```bash
# Run all tests
make test

# Run specific test suite
go test ./src/libraries/kubernetes-client/...

# Run integration tests (requires cluster)
make test-integration

# TUI component tests
go test ./src/libraries/tui-components/...
```

### Building

```bash
# Build both versions
make build-all

# Build specific versions
make build-kuber  # Full kUber 
make build-ktop   # Lightweight kTop

# Cross-platform builds
make build-cross

# Create release packages
make release
```

## 🔄 kUber vs kTop Comparison

| Feature | kUber (Full) | kTop (Monitoring) |
|---------|-------------|------------------|
| **Core Monitoring** | ✅ Full dashboard | ✅ Full dashboard |
| **Cluster Logs** | ✅ Full access | ✅ Full access (read-only) |  
| **Namespace Navigation** | ✅ Yes | ✅ Yes |
| **Resource Navigation** | ✅ All resource types | ✅ All resource types |
| **Resource Editing** | ✅ **YAML Editor** | ❌ Read-only |
| **Pod Shell Access** | ✅ Interactive | ✅ Interactive |
| **Pod Log Streaming** | ✅ Full streaming | ✅ Full streaming |
| **Resource Details** | ✅ Full details | ✅ Full details |
| **Search & Follow** | ✅ Advanced | ✅ Advanced |
| **Resource Description** | ✅ Yes | ✅ Yes |
| **Memory Usage** | 20-50MB | 15-30MB |
| **Security Level** | Medium | High (read-only) |
| **Target Users** | DevOps, Admins | Ops, Security, Monitoring |

### When to Use Each:

**🚀 Use kUber when:**
- Developing or debugging applications
- Need to edit resource configurations
- Require shell access to containers
- Managing cluster resources actively
- Working in development/staging environments

**📊 Use kTop when:**
- Production monitoring and observability  
- Security-sensitive environments requiring read-only access
- CI/CD pipelines and automation
- Need full cluster inspection without modification risks
- Lightweight resource monitoring with full feature set

## 🤝 Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

### Development Workflow
1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Write tests for your changes
4. Implement your feature
5. Run tests (`make test`)
6. Commit changes (`git commit -am 'Add amazing feature'`)
7. Push to branch (`git push origin feature/amazing-feature`)
8. Create a Pull Request

## 🐛 Troubleshooting

### Common Issues

**Connection Failed**
```bash
# Verify kubectl works
kubectl cluster-info

# Check current context  
kubectl config current-context
```

**No Resources Visible**
```bash
# Check RBAC permissions
kubectl auth can-i list pods
kubectl auth can-i list services
```

**Log Streaming Issues**
- Ensure pods are running and have logs
- Check if containers have shell access (bash/sh)
- Verify RBAC permissions for log access

For more troubleshooting, see our [FAQ](https://github.com/anindyar/kuber/wiki/FAQ).

## 📜 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [k9s](https://github.com/derailed/k9s) - Inspiration for Kubernetes TUI management
- [lazydocker](https://github.com/jesseduffield/lazydocker) - UX inspiration
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - Amazing TUI framework
- [Kubernetes](https://kubernetes.io/) - The platform we're managing

## 🔗 Links

- [Documentation](https://github.com/anindyar/kuber/wiki)
- [Releases](https://github.com/anindyar/kuber/releases)
- [Issues](https://github.com/anindyar/kuber/issues)
- [Discussions](https://github.com/anindyar/kuber/discussions)

---

**kUber** - Making Kubernetes management uber simple! 🚀