# SECRA Frontend

Modern Next.js frontend for SECRA (Security Resource Aggregator).

## 🏗️ Tech Stack

- **Framework**: Next.js 16 (App Router)
- **UI Library**: React 19
- **State Management**: Redux Toolkit + RTK Query
- **Styling**: TailwindCSS
- **Icons**: Lucide React
- **Charts**: Recharts
- **Type Safety**: TypeScript

## 📦 Project Structure

```
frontend/
├── src/
│   ├── app/                    # Next.js App Router pages
│   │   ├── page.tsx            # Home (CVE dashboard)
│   │   ├── layout.tsx          # Root layout
│   │   ├── login/              # Login page
│   │   ├── register/           # Registration page
│   │   ├── cves/               # CVE list page
│   │   ├── vendors/            # Vendor list page
│   │   ├── products/           # Product list page
│   │   ├── my/dashboard/       # User dashboard (subscriptions)
│   │   └── admin/users/        # Admin user management
│   ├── components/             # Reusable React components
│   │   ├── Navbar.tsx          # Navigation bar
│   │   ├── Pagination.tsx      # Pagination component
│   │   ├── ViewToggle.tsx      # View mode toggle
│   │   └── AuthInit.tsx        # Auth initialization
│   ├── lib/                    # Utilities and state management
│   │   ├── store.ts            # Redux store configuration
│   │   ├── types.ts            # TypeScript type definitions
│   │   ├── features/           # Redux slices
│   │   │   ├── apiSlice.ts     # RTK Query API endpoints
│   │   │   └── authSlice.ts    # Authentication state
│   │   └── gen/                # Generated Protobuf types
│   └── public/                 # Static assets
├── tests/                      # Test suites (planned)
│   ├── unit/                   # Unit tests
│   ├── integration/            # Component tests
│   └── e2e/                    # E2E tests
├── Dockerfile                  # Container image
├── next.config.ts              # Next.js configuration
├── tailwind.config.ts          # TailwindCSS configuration
├── tsconfig.json               # TypeScript configuration
└── package.json                # Dependencies
```

## 🛠️ Development

### Prerequisites

- Node.js 20+
- npm or yarn
- [Buf CLI](https://buf.build/docs/installation) (for generating Protobuf types)

### Installation

```bash
# Install Buf (macOS)
brew install bufbuild/buf/buf

# Install dependencies
npm install
```

### Generate Protobuf Types

**Important:** Generated TypeScript types from Protobuf are **not** committed to git. They must be generated before development or build:

```bash
# Generate types manually
npm run gen:proto

# Or they will auto-generate when running:
npm run dev    # Automatically runs predev hook
npm run build  # Automatically runs prebuild hook
```

### Development Server

```bash
npm run dev
```

Open [http://localhost:3000](http://localhost:3000) in your browser.

### Build for Production

```bash
npm run build
```

Static files will be exported to `out/` directory.

### Linting

```bash
npm run lint
```

## 🧪 Testing (Planned)

### Unit Tests

```bash
npm test
```

### Component Tests

```bash
npm run test:integration
```

### E2E Tests

```bash
npm run test:e2e
```

### Test Coverage

```bash
npm run test:coverage
```

## 🎨 Styling

### TailwindCSS

The project uses TailwindCSS for styling with custom configuration:

```typescript
// tailwind.config.ts
{
  theme: {
    extend: {
      colors: {
        // Custom color palette
      }
    }
  }
}
```

### Dark Theme

All pages support a cyberpunk-inspired dark theme by default.

## 📡 API Integration

### RTK Query

API calls are managed through RTK Query:

```typescript
// src/lib/features/apiSlice.ts
import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react';

export const apiSlice = createApi({
  baseQuery: fetchBaseQuery({
    baseUrl: '/api/v1',
    credentials: 'include',
  }),
  endpoints: (builder) => ({
    getMe: builder.query<User, void>({
      query: () => '/me',
    }),
    // More endpoints...
  }),
});
```

### Authentication

Session-based authentication using HTTP-only cookies:

```typescript
// Login
const { data } = await login({ username, password });

// Auto-fetch user on page load
const { data: user } = useGetMeQuery();
```

## 🗂️ State Management

### Redux Store

```typescript
// src/lib/store.ts
import { configureStore } from '@reduxjs/toolkit';
import { apiSlice } from './features/apiSlice';
import authReducer from './features/authSlice';

export const store = configureStore({
  reducer: {
    [apiSlice.reducerPath]: apiSlice.reducer,
    auth: authReducer,
  },
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware().concat(apiSlice.middleware),
});
```

### Auth Slice

```typescript
// src/lib/features/authSlice.ts
export const authSlice = createSlice({
  name: 'auth',
  initialState: {
    user: null,
    isAuthenticated: false,
  },
  reducers: {
    setUser: (state, action) => {
      state.user = action.payload;
      state.isAuthenticated = true;
    },
    logout: (state) => {
      state.user = null;
      state.isAuthenticated = false;
    },
  },
});
```

## 🔧 Configuration

### Environment Variables

```env
# API URL (proxied by Nginx in production)
NEXT_PUBLIC_API_URL=/api/v1
```

### Next.js Config

```typescript
// next.config.ts
const nextConfig = {
  output: 'export', // Static export
  // Other config...
};
```

## 🐳 Docker

### Build Image

```bash
docker build -f Dockerfile -t secra-frontend:latest ..
```

### Run Container

```bash
docker run -p 80:80 secra-frontend:latest
```

## 🚀 Deployment

### Production Build

The frontend is built as a static site and served via Nginx:

1. Next.js builds static files (`npm run build`)
2. Files are copied to `/usr/share/nginx/html`
3. Nginx serves static files and proxies `/api/*` to backend

### Nginx Configuration

Embedded in Dockerfile:

```nginx
server {
    listen 80;
    location / {
        root /usr/share/nginx/html;
        try_files $uri $uri.html /index.html;
    }
    location /api/ {
        proxy_pass http://server:8080;
        proxy_set_header Host $host;
    }
}
```

## 📄 Key Features

### Pages

- **Home (`/`)**: CVE dashboard with charts and recent CVEs
- **Login (`/login`)**: User authentication
- **Register (`/register`)**: New user registration
- **CVEs (`/cves`)**: Paginated CVE list with search
- **Vendors (`/vendors`)**: Vendor list with subscription
- **Products (`/products`)**: Product list with subscription
- **My Dashboard (`/my/dashboard`)**: User subscriptions with tabs and pagination
- **Admin Users (`/admin/users`)**: User management (admin only)

### Components

- **Navbar**: Responsive navigation with auth status
- **Pagination**: Reusable pagination component
- **ViewToggle**: Grid/List view switcher
- **AuthInit**: Auto-fetch user on app load

### State Features

- Auto-refresh user session
- Persistent login state
- API request caching
- Optimistic updates

## 📝 Code Style

### TypeScript

Strict mode enabled:

```json
{
  "compilerOptions": {
    "strict": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true
  }
}
```

### ESLint

```bash
npm run lint
```

## 📄 License

See [LICENSE](../LICENSE)
