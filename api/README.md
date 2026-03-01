# SECRA API

Shared API definitions using Protocol Buffers (Protobuf) and gRPC.

## 📁 Structure

```
api/
├── v1/                         # API version 1 definitions
│   ├── cve.proto               # CVE resource
│   ├── cve_source.proto        # CVE source resource
│   ├── product.proto           # Product resource
│   ├── subscription.proto      # Subscription resource
│   ├── user.proto              # User resource
│   └── vendor.proto            # Vendor resource
├── gen/                        # Generated code
│   └── v1/                     # Generated Go code
│       ├── cve_grpc.pb.go
│       ├── cve.pb.go
│       ├── cve.pb.gw.go        # gRPC-Gateway REST proxy
│       ├── cve.swagger.json    # OpenAPI/Swagger docs
│       └── ...                 # Other generated files
├── buf.yaml                    # Buf configuration
└── buf.gen.yaml                # Code generation config
```

## 🔧 Technology

- **Protocol Buffers v3**: Interface definition language
- **gRPC**: High-performance RPC framework
- **gRPC-Gateway**: REST API proxy for gRPC services
- **Buf**: Modern Protobuf toolchain

## 📝 Proto Definitions

### Core Resources

1. **CVE** (`cve.proto`)
   - CVE details (ID, description, severity)
   - CVSS scores
   - References and weaknesses
   - Affected products

2. **Product** (`product.proto`)
   - Product information
   - Vendor association
   - CPE identifiers

3. **Vendor** (`vendor.proto`)
   - Vendor information
   - Associated products

4. **Subscription** (`subscription.proto`)
   - User subscriptions
   - Target types (vendor/product)
   - Notification preferences

5. **User** (`user.proto`)
   - User management
   - Authentication
   - Authorization (admin/user roles)

6. **CVE Source** (`cve_source.proto`)
   - Data source tracking (NVD)
   - Last sync timestamps

### Service Definitions

Each proto file defines:
- **Messages**: Data structures (requests/responses)
- **Services**: RPC methods (CRUD operations)
- **Annotations**: HTTP/REST mappings for gRPC-Gateway

Example:

```protobuf
service CVEService {
  // List CVEs with pagination
  rpc ListCVEs(ListCVEsRequest) returns (ListCVEsResponse) {
    option (google.api.http) = {
      get: "/api/v1/cves"
    };
  }
  
  // Get CVE by ID
  rpc GetCVE(GetCVERequest) returns (GetCVEResponse) {
    option (google.api.http) = {
      get: "/api/v1/cves/{cve_id}"
    };
  }
}
```

## 🛠️ Code Generation

### Prerequisites

```bash
# Install Buf
brew install bufbuild/buf/buf  # macOS
# or
go install github.com/bufbuild/buf/cmd/buf@latest
```

### Generate Code

**Important:** Generated code is **not** committed to git repository. It is automatically generated during:
- Docker build (both backend and frontend Dockerfiles)
- Local development (via npm scripts for frontend)
- CI/CD pipelines (recommended)

```bash
# From api/ directory
buf generate
```

Generated files are ignored by `.gitignore`:
- `api/gen/` - Go backend code
- `frontend/src/lib/gen/` - TypeScript frontend types

This generates:
- Go code (`gen/v1/*.pb.go`)
- gRPC service code (`gen/v1/*_grpc.pb.go`)
- gRPC-Gateway proxies (`gen/v1/*.pb.gw.go`)
- OpenAPI/Swagger specs (`gen/v1/*.swagger.json`)

### Configuration

**buf.yaml**:
```yaml
version: v2
modules:
  - path: v1
lint:
  use:
    - STANDARD
breaking:
  use:
    - FILE
```

**buf.gen.yaml**:
```yaml
version: v2
managed:
  enabled: true
  override:
    - file_option: go_package_prefix
      value: github.com/yourusername/secra/api/gen
plugins:
  - remote: buf.build/protocolbuffers/go
    out: gen
    opt:
      - paths=source_relative
  - remote: buf.build/grpc/go
    out: gen
    opt:
      - paths=source_relative
  - remote: buf.build/grpc-ecosystem/gateway
    out: gen
    opt:
      - paths=source_relative
  - remote: buf.build/grpc-ecosystem/openapiv2
    out: gen
    opt:
      - allow_merge=true
      - merge_file_name=api
```

## 🔗 Usage

### Backend (Go)

Import generated code:

```go
import (
    pb "github.com/yourusername/secra/api/gen/v1"
)

// Implement gRPC service
type cveService struct {
    pb.UnimplementedCVEServiceServer
}

func (s *cveService) ListCVEs(ctx context.Context, req *pb.ListCVEsRequest) (*pb.ListCVEsResponse, error) {
    // Implementation
}
```

### Frontend (TypeScript)

For REST API calls via gRPC-Gateway:

```typescript
// RTK Query endpoint
getCVE: builder.query<CVE, string>({
  query: (cveId) => `/cves/${cveId}`,
}),
```

## 📊 API Endpoints

Generated REST endpoints (via gRPC-Gateway):

### CVEs
- `GET /api/v1/cves` - List CVEs (paginated)
- `GET /api/v1/cves/{cve_id}` - Get CVE details
- `GET /api/v1/cves/recent` - Get recent CVEs

### Products
- `GET /api/v1/products` - List products (paginated)
- `GET /api/v1/products/{id}` - Get product details

### Vendors
- `GET /api/v1/vendors` - List vendors (paginated)
- `GET /api/v1/vendors/{id}` - Get vendor details

### Subscriptions
- `GET /api/v1/subscriptions` - List user subscriptions
- `POST /api/v1/subscriptions` - Create subscription
- `DELETE /api/v1/subscriptions/{id}` - Delete subscription

### Users
- `POST /api/v1/users/register` - Register new user
- `POST /api/v1/users/session` - Login
- `DELETE /api/v1/users/session` - Logout
- `GET /api/v1/me` - Get current user
- `GET /api/v1/admin/users` - List all users (admin)
- `PUT /api/v1/admin/users/{id}/role` - Update user role (admin)

### CVE Sources
- `GET /api/v1/sources` - List CVE sources
- `POST /api/v1/sources/sync` - Trigger sync (admin)

## 🔄 Protocol Evolution

### Adding New Fields

1. Edit proto file (add new field with unique number)
2. Regenerate code: `buf generate`
3. Update backend service implementation
4. Update frontend API calls

### Versioning

- Current version: `v1`
- Breaking changes require new version: `v2`
- Non-breaking changes can extend `v1`

### Best Practices

- Always add fields (never remove)
- Use `optional` for nullable fields
- Reserve removed field numbers
- Use semantic field numbers (1-15 for frequent fields)

## 🧪 Testing

### Validate Proto Files

```bash
buf lint
```

### Check Breaking Changes

```bash
buf breaking --against '.git#branch=main'
```

## 📚 Documentation

### OpenAPI/Swagger

Generated Swagger files are available at:
- `gen/v1/*.swagger.json`

Merge all specs:
```bash
buf generate  # Generates merged api.swagger.json
```

### Proto Docs

Generate documentation:
```bash
buf generate --template buf.gen.docs.yaml
```

## 🚀 Deployment

Generated code is committed to version control for:
- **Backend**: Go code consumed directly
- **Frontend**: REST endpoints via gRPC-Gateway

## 📝 Development Workflow

1. Edit `.proto` files in `v1/`
2. Run `buf generate` to generate code
3. Commit both proto files and generated code
4. Backend automatically uses new gRPC definitions
5. Frontend calls new REST endpoints via gRPC-Gateway

## 📄 License

See [LICENSE](../LICENSE)
