# Quickstart: Kubernetes TUI Manager

## Prerequisites

- Go 1.21 or later
- Access to a Kubernetes cluster
- kubectl configured with valid context
- Terminal with color support (recommended)

## Installation

```bash
# Clone the repository
git clone https://github.com/anindyar/kuber.git
cd kuber

# Build the application
go build -o kuber ./cmd/kuber

# Make it executable
chmod +x kuber

# Optional: Install to PATH
sudo mv kuber /usr/local/bin/
```

## Quick Start

### 1. Launch the Application

```bash
# Use default kubectl context
kuber

# Or specify a context
kuber --context=my-cluster

# Or specify kubeconfig file
kuber --kubeconfig=/path/to/config
```

### 2. Basic Navigation

- **Tab**: Switch between panels
- **Arrow Keys**: Navigate within panels
- **Enter**: Select/activate item
- **Esc**: Go back/cancel
- **q**: Quit application
- **h**: Show help

### 3. Core Workflows

#### Browse Resources
1. Launch kuber
2. Use arrow keys to navigate resource types in sidebar
3. Press Enter to view resources of selected type
4. Navigate through resource list with arrow keys
5. Press Enter to view resource details

#### View Pod Logs
1. Navigate to Pods in sidebar
2. Select a pod from the list
3. Press 'l' to view logs
4. Use Page Up/Down to scroll through logs
5. Press 'f' to follow logs in real-time
6. Press Esc to return to pod list

#### Access Pod Shell
1. Navigate to Pods in sidebar
2. Select a running pod
3. Press 's' to open shell
4. Interactive shell session opens
5. Type commands as normal
6. Press Ctrl+D or type 'exit' to close shell

#### Edit Resources
1. Navigate to any resource
2. Select the resource you want to edit
3. Press 'e' to edit
4. YAML editor opens with resource configuration
5. Make changes using text editor controls
6. Press Ctrl+S to save changes
7. Press Esc to cancel without saving

#### View Metrics
1. Press 'm' from any view to open metrics
2. Select resource type from sidebar
3. Choose specific resource
4. View real-time performance charts
5. Use number keys (1-5) to switch time ranges:
   - 1: Last 1 minute
   - 2: Last 5 minutes  
   - 3: Last 15 minutes
   - 4: Last hour
   - 5: Last 6 hours

## Testing the Setup

### Test Scenarios

#### Scenario 1: Basic Connectivity
**Given** kuber is launched
**When** application starts
**Then** cluster overview dashboard is displayed
**And** connection status shows "Connected"
**And** cluster information is visible (version, nodes)

#### Scenario 2: Resource Browsing
**Given** kuber is connected to cluster
**When** user navigates to Pods section
**Then** list of pods is displayed with status
**And** pods can be filtered by namespace
**And** pod details are accessible

#### Scenario 3: Log Viewing
**Given** a running pod exists
**When** user selects pod and presses 'l'
**Then** logs are displayed in real-time
**And** logs can be scrolled
**And** timestamps are visible

#### Scenario 4: Resource Editing
**Given** user has edit permissions
**When** user selects a ConfigMap and presses 'e'
**Then** YAML editor opens
**And** changes can be saved
**And** Kubernetes resource is updated

#### Scenario 5: Metrics Display
**Given** metrics server is available
**When** user navigates to metrics view
**Then** CPU and memory charts are displayed
**And** data updates in real-time
**And** historical data is available

### Validation Commands

```bash
# Verify cluster connectivity
kubectl cluster-info

# Check available resources
kubectl api-resources

# Verify metrics server (if using)
kubectl top nodes

# Test RBAC permissions
kubectl auth can-i '*' '*' --all-namespaces
```

## Troubleshooting

### Common Issues

#### Cannot Connect to Cluster
- **Symptom**: "Connection failed" error on startup
- **Solution**: 
  ```bash
  # Verify kubectl works
  kubectl cluster-info
  
  # Check context
  kubectl config current-context
  
  # List available contexts
  kubectl config get-contexts
  ```

#### No Resources Visible
- **Symptom**: Empty resource lists
- **Solution**:
  ```bash
  # Check RBAC permissions
  kubectl auth can-i list pods
  kubectl auth can-i list services
  
  # Verify resources exist
  kubectl get all --all-namespaces
  ```

#### Metrics Not Available
- **Symptom**: "Metrics unavailable" message
- **Solution**:
  ```bash
  # Check if metrics server is installed
  kubectl get deployment metrics-server -n kube-system
  
  # Test metrics API
  kubectl top nodes
  kubectl top pods
  ```

#### Shell Access Fails
- **Symptom**: "Cannot execute command" error
- **Solution**:
  - Ensure pod is running
  - Verify container has shell (bash/sh)
  - Check RBAC permissions for exec

### Configuration

#### Custom Keybindings
Create `~/.kuber/config.yaml`:

```yaml
keybindings:
  quit: "q"
  help: "h" 
  logs: "l"
  shell: "s"
  edit: "e"
  metrics: "m"
  refresh: "r"

ui:
  theme: "dark"  # or "light"
  showTimestamps: true
  autoRefresh: 30  # seconds
```

#### Cluster Profiles
```yaml
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

## Performance Notes

- Application uses ~20-50MB memory idle
- Real-time updates every 5 seconds by default
- Log streaming limited to 1000 lines by default
- Metrics retained for 1 hour by default

## Next Steps

- Configure multiple cluster profiles
- Set up custom themes and keybindings
- Explore advanced filtering and search
- Set up metrics alerts and thresholds

For detailed documentation, see the [User Guide](./docs/user-guide.md).

For API documentation, see the [API Reference](./docs/api-reference.md).

For troubleshooting, see the [FAQ](./docs/faq.md).