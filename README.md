# kUber - An Uber Kubernetes Manager

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Go Version](https://img.shields.io/badge/go-1.24+-blue.svg)
![Release](https://img.shields.io/github/v/release/anindyar/kuber?include_prereleases)

A powerful, intuitive terminal-based Kubernetes cluster manager built with Go and Bubble Tea. kUber provides an enhanced user experience for managing Kubernetes resources with real-time log streaming, multi-container support, and aggregated logging for deployments.

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

## 🚀 Quick Install

### One-line Installation (Linux/macOS)

```bash
curl -sSL https://raw.githubusercontent.com/anindyar/kuber/main/install.sh | sh
```

This will download the latest release and install it to `/usr/local/bin/kuber`.

### Manual Installation

#### Download Pre-built Binary

1. Download the latest release for your platform from the [releases page](https://github.com/anindyar/kuber/releases)
2. Extract and install:

```bash
# For Linux x64
wget https://github.com/anindyar/kuber/releases/latest/download/kuber-linux-amd64.tar.gz
tar -xzf kuber-linux-amd64.tar.gz
sudo mv kuber /usr/local/bin/
chmod +x /usr/local/bin/kuber
```

#### Build from Source

**Prerequisites:**
- Go 1.24 or later
- kubectl configured with valid cluster access

```bash
# Clone the repository
git clone https://github.com/anindyar/kuber.git
cd kuber

# Build the application
make build

# Install to system
sudo make install
```

## 🎯 Quick Start

```bash
# Launch with default kubectl context
kuber

# Use specific context
kuber --context=my-cluster

# Use custom kubeconfig
kuber --kubeconfig=/path/to/config
```

### 🎮 Basic Controls

| Key | Action |
|-----|--------|
| `Tab` | Switch between panels |
| `↑/↓` | Navigate lists |
| `Enter` | Select/view details |
| `l` | View pod logs |
| `c` | View cluster logs |
| `f` | Toggle log follow mode |
| `r` | Refresh current view |
| `s` | Open pod shell |
| `e` | Edit resource |
| `/` | Search/filter |
| `Esc` | Go back/cancel |
| `q` | Quit |
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

### Resource Editing
1. Select any resource (ConfigMap, Deployment, etc.)
2. Press `e` to open the YAML editor
3. Make your changes
4. Press `Ctrl+S` to save
5. Changes are applied directly to the cluster

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
# Development build
make build-dev

# Production build
make build

# Cross-platform builds
make build-all

# Create release
make release
```

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