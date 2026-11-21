# trading-app — Frontend

This React frontend (v8, TypeScript, Vite, Tailwind CSS) connects to the Go modular monolith backend via `http://localhost:8000/api`.

## Getting Started

```sh
# Install dependencies
pnpm install

# Copy env config
cp frontend/.env.example frontend/.env

# Run the frontend (default: http://localhost:8000)
pnpm run dev
```

Available scripts (from `package.json`):
- `dev` — start Vite dev server (proxies `/api` to gateway)
- `build` — produce production build
- `preview` — preview production build locally

## Environment

Copy `frontend/.env.example` → `frontend/.env` and adjust:

| Variable | Default | Description |
| --- | --- | --- |
| `PORT` | `8000` | Vite dev server port |
| `GATEWAY_URL` | `http://localhost:3001` | Backend gateway base URL (for proxy) |
| `VITE_CSRF_COOKIE_NAME` | `auth_token` | CSRF cookie name (must match gateway config) |

## Architecture

- **Routing**: `react-router` v8 with `createBrowserRouter`, `loaders` for auth + data fetching
- **State**: `@tanstack/react-query` for server state (refetch on window focus disabled, retry: 1)
- **Forms**: `react-hook-form` with Zod-like validation (via `response.NewValidator()` on backend)
- **Styling**: Tailwind CSS 4 (config via `tailwindcss` Vite plugin)
- **Icons**: `lucide-react` (full library, import specific icons: `import { Menu } from "lucide-react"`)
- **Toasts**: `sonner` for error/success notifications
- **HTTP client**: Custom `ApiClient` in `src/lib/http.ts` with:
  - Automatic retry on 5xx server errors (2 attempts, 1s/2s backoff)
  - Centralised `ApiError` handling (status 0 = network, status 500 = server error)
  - CSRF token management from non-HttpOnly cookie
  - Axios error → user-friendly message via `getErrorMessage()` in `src/lib/apiError.ts`

### Key Directories

```
src/
  app.tsx          — Root component with Outlet + OfflineBanner
  lib/
    http.ts        — ApiClient with retry & error mapping
    apiError.ts    — getErrorMessage() utility
    router.ts      — React Router configuration + ErrorBoundary on routes
    queryClient.ts — TanStack Query config
    useOnlineStatus.ts  — online/offline detection
    csrf.ts        — CSRF cookie handling
  components/
    ErrorBoundary.tsx    — catches React rendering errors, shows retry UI
    OfflineBanner.tsx    — shows "You are offline" banner when navigator.onLine=false
    DepthChart.tsx       — depth chart visualization
    OrderBook.tsx        — order book display (bids/asks)
    Sidebar.tsx          — navigation sidebar
  pages/
    SecurityDetailPage.tsx   — security detail with ErrorBoundary + OrderBook + DepthChart
    OrderDetailsPage.tsx    — order detail page with ErrorBoundary
    BlotterRoute.tsx       — orders blotter view with inline error handling
    HomePage.tsx           — home page
    LoginPage.tsx          — authentication
    RegisterPage.tsx       — registration
```

## Error Handling

When the backend returns errors, the user sees friendly messages:

| Error Type | User Message |
| --- | --- |
| Network failure (no route to host) | "Can't reach the server. Check your connection." |
| Backend 500 error | "Internal server error. Please try again later or contact support." |
| Other HTTP errors | The HTTP status text or "error" field from the response |
| User goes offline | Red banner at top of app: "You are offline. Some features may not work properly." |
| React rendering error | ErrorBoundary UI with "Unexpected Application Error!" and Retry button |

Each page uses `getErrorMessage()` to convert `ApiError` into readable text, and routes with data-fetching have `ErrorBoundary` mounted via `errorElement` to catch unexpected rendering errors.

## Recent Improvements

- **ErrorBoundary** on `/securities/:securityId`, `/orders/:orderId`, and `/blotter` routes — catches React errors and shows a clean UI instead of raw stack traces
- **Offline detection** — banner appears when connection is lost, disappears when back online
- **Retry logic** — ApiClient automatically retries 5xx errors with exponential backoff
- **Better 500 messages** — specific "Internal server error..." message instead of generic fallthrough
- **ApiError.status 0 handling** — "Can't reach the server. Check your connection." for network failures