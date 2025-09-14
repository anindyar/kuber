# kTop - Kubernetes Monitoring Tool (Read-Only)

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Go Version](https://img.shields.io/badge/go-1.24+-blue.svg)

A lightweight, read-only terminal interface for monitoring Kubernetes clusters with real-time dashboard and logs viewing.

## ✨ Features

- 📊 **Real-time Dashboard** - Comprehensive cluster performance monitoring
- 🖥️ **Multi-Node Monitoring** - Resource pressure analysis across all cluster nodes  
- 📜 **Cluster Log Viewer** - Read-only log streaming from system namespaces
- 🔍 **Advanced Search** - Real-time keyword filtering in logs
- 📈 **Performance Metrics** - CPU, memory, and storage utilization tracking
- 🚀 **Workload Overview** - Live counts of deployments, pods, services, etc.
- ⚡ **High Performance** - Optimized for large clusters with efficient polling
- 🔒 **Read-Only Access** - Secure monitoring without modification capabilities

## 🚀 Quick Install

### One-Line Installer (Recommended)

```bash
# Install kTop directly to /usr/local/bin
curl -fsSL https://raw.githubusercontent.com/anindyar/kuber/main/install-ktop.sh | bash
```

### Alternative Install Methods

```bash
# Download and install manually
wget https://raw.githubusercontent.com/anindyar/kuber/main/install-ktop.sh
chmod +x install-ktop.sh
./install-ktop.sh
```

### Build from Source

```bash
# Clone the repository
git clone https://github.com/anindyar/kuber.git
cd kuber

# Build kTop
go build -o ktop ./cmd/ktop

# Install to system
sudo mv ktop /usr/local/bin/
```

## 🎯 Quick Start

```bash
# Launch with default kubectl context
ktop

# Use specific context
ktop --context=my-cluster

# Use custom kubeconfig
ktop --kubeconfig=/path/to/config
```

### 🎮 Controls

**Main Navigation:**
| Key | Action |
|-----|--------|
| `Enter` | Navigate forward/Select resource |
| `↑/↓` | Navigate lists and tables |
| `Tab` | Switch between panes (tabs ↔ table) |
| `c` | View cluster logs (from dashboard) |
| `r` | Refresh current view |
| `Esc` | Go back/cancel |
| `q` | Quit application |

**Resource View Controls:**
| Key | Action |
|-----|--------|
| `l` | View logs (pods/deployments/statefulsets) |
| `d` | View resource details |
| `s` | Shell access (pods only - limited) |
| `Enter` | Select resource or view logs |

**Log View Controls:**
| Key | Action |
|-----|--------|
| `/` | Search/filter logs |
| `f` | Toggle follow mode |
| `r` | Refresh logs |
| `Esc` | Exit search mode or go back |

## 📖 Usage Examples

### Dashboard Overview
The main dashboard provides:
- **📊 Cluster Performance Monitor**: Real-time CPU, memory, and storage utilization
- **📈 Resource Pressure Metrics**: Multi-node resource pressure analysis  
- **🖥️ Per-Node Status**: Individual node health with resource scores
- **📊 Workload Counts**: Live counts of deployments, pods, services, etc.

### Cluster Log Monitoring
1. From the main dashboard, press `c` to access cluster logs
2. View aggregated logs from system namespaces (kube-system, default)
3. Use `/` to search across all cluster logs
4. Press `r` to manually refresh log content
5. Press `Esc` to return to dashboard

### Resource Navigation  
1. From dashboard, press `Enter` to view namespaces
2. Select a namespace with `Enter` to view resources
3. Use `Tab` to switch between resource types (left) and resource table (right)
4. Navigate resources with arrow keys, press `Enter` or `l` to view logs
5. For deployments/statefulsets, `l` shows aggregated logs from all pods
6. Press `d` to view detailed resource information
7. Use `Esc` to navigate back through the hierarchy

### Log Viewing
- **Pod logs**: Direct kubectl logs output 
- **Deployment/StatefulSet logs**: Aggregated logs from all associated pods
- **Search functionality**: Use `/` to filter logs in real-time
- **Debug mode**: Shows pod discovery process when no logs found

## 🏗️ Architecture

kTop is built on the same foundation as kUber but optimized for read-only monitoring:

```
├── kubernetes-client/    # K8s API communication (read-only)
├── tui-components/      # Lightweight UI widgets  
├── resource-manager/    # Resource caching (no watches)
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

# Cluster Profiles
clusters:
  - name: "production"
    context: "prod-cluster"
    namespace: "default"
    readOnly: true        # Always true for kTop
    
  - name: "development"
    context: "dev-cluster" 
    namespace: "dev"
    readOnly: true        # Always true for kTop
```

## 📊 Performance

- **Memory Usage**: ~15-30MB idle (lighter than kUber)
- **Update Frequency**: 30 seconds (configurable)  
- **Navigation**: <50ms response time
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

# Build kTop
make build-ktop

# Run with development flags
./ktop --debug
```

## 🤝 kTop vs kUber

| Feature | kTop | kUber |
|---------|------|-------|
| **Monitoring** | ✅ Full | ✅ Full |
| **Log Viewing** | ✅ Read-only + Aggregated | ✅ Full |
| **Resource Navigation** | ✅ Full | ✅ Full |
| **Resource Editing** | ❌ No | ✅ Yes |
| **Shell Access** | ⚠️ Limited (pods only) | ✅ Full |
| **Deployment Logs** | ✅ Aggregated | ✅ Full |
| **Memory Usage** | 15-30MB | 20-50MB |
| **Security** | High (read-only) | Medium |
| **Use Case** | Production monitoring | Development/Admin |

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
kubectl auth can-i list nodes
```

**Log Access Issues**
- Ensure you have read permissions for system namespaces
- Verify RBAC permissions for log access  
- Check if kubectl can access logs manually
- For deployment logs: verify pods exist and are running
- Use debug mode (press `l` on deployment) to see pod discovery process

**Navigation Issues**
- Use `Tab` to switch between resource tabs and table
- Arrow keys work in both panes when focused
- Resource-specific actions shown in status hints

## 📜 License

This project is licensed under the MIT License - see the [LICENSE](../../LICENSE) file for details.

## 🙏 Acknowledgments

- [k9s](https://github.com/derailed/k9s) - Inspiration for Kubernetes TUI management
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - Amazing TUI framework
- [Kubernetes](https://kubernetes.io/) - The platform we're monitoring

---

**kTop** - Lightweight Kubernetes monitoring made simple! 🚀

*For full editing capabilities and advanced features, see [kUber](../kuber/README.md)*