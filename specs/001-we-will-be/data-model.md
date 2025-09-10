# Data Model: Kubernetes TUI Manager

## Core Entities

### Cluster
Represents a connected Kubernetes cluster.

**Fields**:
- `Name`: string - Display name for the cluster
- `Endpoint`: string - Kubernetes API server URL
- `Context`: string - kubectl context name
- `Auth`: AuthConfig - Authentication configuration
- `Status`: ClusterStatus - Connection and health status
- `Version`: string - Kubernetes version
- `Nodes`: int - Number of nodes
- `LastSeen`: time.Time - Last successful API call

**Relationships**:
- One-to-many with Resources
- One-to-many with Namespaces

**Validation Rules**:
- Name must be unique within application
- Endpoint must be valid URL with https scheme
- Context must exist in kubeconfig

**State Transitions**:
Connected → Disconnected → Reconnecting → Connected
Connected → Error → Reconnecting → Connected

### Resource
Generic representation of any Kubernetes resource.

**Fields**:
- `Kind`: string - Resource type (Pod, Service, Deployment, etc.)
- `Name`: string - Resource name
- `Namespace`: string - Kubernetes namespace (empty for cluster-scoped)
- `UID`: string - Kubernetes unique identifier
- `CreationTime`: time.Time - When resource was created
- `Status`: ResourceStatus - Current state
- `Labels`: map[string]string - Kubernetes labels
- `Annotations`: map[string]string - Kubernetes annotations
- `Spec`: interface{} - Resource specification
- `StatusData`: interface{} - Resource status data
- `Events`: []Event - Related Kubernetes events

**Relationships**:
- Belongs to one Cluster
- Belongs to one Namespace (if namespaced)
- One-to-many with Events
- Many-to-many relationships with other Resources (via OwnerReferences)

**Validation Rules**:
- Kind must be valid Kubernetes resource type
- Name must be valid Kubernetes object name
- Namespace must exist in cluster (if specified)

### Namespace
Kubernetes namespace representation.

**Fields**:
- `Name`: string - Namespace name
- `Status`: NamespaceStatus - Active, Terminating
- `ResourceQuota`: map[string]string - Resource limits
- `CreationTime`: time.Time - When namespace was created

**Relationships**:
- Belongs to one Cluster
- One-to-many with Resources

### LogEntry
Individual log line from pods/containers.

**Fields**:
- `Timestamp`: time.Time - When log was generated
- `Source`: LogSource - Pod name and container name
- `Content`: string - Actual log message
- `Level`: LogLevel - Info, Warning, Error, Debug
- `Stream`: StreamType - stdout or stderr

**Relationships**:
- Belongs to one Resource (Pod)

**Validation Rules**:
- Timestamp must be valid time
- Source must reference existing pod and container
- Content must be UTF-8 text

### MetricDataPoint
Performance and resource usage measurement.

**Fields**:
- `Timestamp`: time.Time - When metric was collected
- `ResourceID`: string - Target resource identifier
- `MetricType`: MetricType - CPU, Memory, Network, Storage
- `Value`: float64 - Metric value
- `Unit`: string - Measurement unit (bytes, cores, etc.)
- `Labels`: map[string]string - Additional metric metadata

**Relationships**:
- Belongs to one Resource

**Validation Rules**:
- Value must be non-negative
- Unit must be valid metric unit
- MetricType must be supported type

### UserSession
Current user's application state and preferences.

**Fields**:
- `ActiveCluster`: string - Currently selected cluster
- `ActiveNamespace`: string - Currently selected namespace
- `CurrentView`: ViewType - Which screen is active
- `ViewHistory`: []ViewState - Navigation history
- `Preferences`: UserPreferences - UI settings
- `Filters`: map[string]FilterConfig - Active resource filters

**Relationships**:
- References one active Cluster
- Contains multiple ViewStates

### NavigationContext
Current location within the resource hierarchy.

**Fields**:
- `ViewType`: ViewType - Dashboard, ResourceList, ResourceDetail, Logs, Metrics
- `ResourceKind`: string - Selected resource type (if applicable)
- `ResourceName`: string - Selected resource name (if applicable)
- `Namespace`: string - Current namespace filter
- `Breadcrumbs`: []NavigationStep - Path to current location

**Relationships**:
- Part of UserSession
- References current Resource (if applicable)

## Enumerations

### ClusterStatus
- `Connected` - Successfully connected and healthy
- `Disconnected` - Connection lost or failed
- `Reconnecting` - Attempting to restore connection
- `Error` - Authentication or permission error
- `Unknown` - Status not yet determined

### ResourceStatus  
- `Running` - Resource is active and healthy
- `Pending` - Resource is being created or scheduled
- `Failed` - Resource has encountered an error
- `Succeeded` - Resource completed successfully
- `Terminating` - Resource is being deleted
- `Unknown` - Status cannot be determined

### ViewType
- `Dashboard` - Cluster overview and summary
- `ResourceList` - List of resources of specific type
- `ResourceDetail` - Individual resource details
- `Logs` - Log viewer for pods/containers
- `Metrics` - Performance metrics and charts
- `Shell` - Interactive shell session

### LogLevel
- `Debug` - Detailed debugging information
- `Info` - General informational messages
- `Warning` - Warning conditions
- `Error` - Error conditions

### MetricType
- `CPU` - CPU usage and limits
- `Memory` - Memory usage and limits
- `Network` - Network I/O statistics
- `Storage` - Disk usage and I/O
- `Custom` - Application-specific metrics

## Data Flow Patterns

### Resource Updates
1. Background goroutine polls Kubernetes API
2. Changes detected via ResourceVersion comparison
3. Resource cache updated with new data
4. UI components notified via Bubble Tea messages
5. Views re-render with updated information

### Log Streaming
1. User selects pod for log viewing
2. Kubernetes API stream established
3. Log entries parsed and buffered
4. UI updated in real-time via Bubble Tea commands
5. Automatic scroll and filtering applied

### Metrics Collection
1. Metrics API polled at regular intervals
2. Data points aggregated and stored
3. Historical data maintained in memory
4. Charts and graphs updated via UI messages

## Caching Strategy

### Resource Cache
- In-memory storage with TTL expiration
- Clustered by namespace for efficient filtering
- LRU eviction when memory limits reached
- Persistent cache for offline viewing (optional)

### Metrics Cache
- Ring buffer for time-series data
- Configurable retention period
- Automatic aggregation for historical views
- Memory-efficient storage format