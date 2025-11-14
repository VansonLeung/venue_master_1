# Admin CMS - Complete Setup Guide

## Quick Start

```bash
# Navigate to the project
cd frontend_codes/admin_cms

# Install dependencies
npm install

# Create environment file
cp .env.example .env

# Start development server
npm run dev
```

Then open [http://localhost:3001](http://localhost:3001) in your browser.

## Detailed Setup Instructions

### Step 1: Prerequisites

Make sure you have:
- ✅ Node.js 18 or higher installed
- ✅ npm or yarn package manager
- ✅ Backend API services running (docker-compose up)

Check your Node version:
```bash
node --version  # Should be v18.0.0 or higher
```

### Step 2: Install Dependencies

```bash
cd frontend_codes/admin_cms
npm install
```

This will install:
- React 18
- Vite
- React Router
- Axios
- Radix UI components
- Tailwind CSS
- Lucide icons
- All other dependencies

### Step 3: Environment Configuration

Create `.env` file:
```bash
cp .env.example .env
```

Edit `.env` and configure:
```env
VITE_API_BASE_URL=http://localhost:8080
```

**Note**: The Vite proxy is configured in `vite.config.js` to forward `/api/*` requests to the backend.

### Step 4: Start Development Server

```bash
npm run dev
```

The app will start on port 3001:
```
VITE v5.0.8  ready in 500 ms

➜  Local:   http://localhost:3001/
➜  Network: use --host to expose
➜  press h to show help
```

### Step 5: Login

Open [http://localhost:3001](http://localhost:3001) and login with:
- **Email**: `admin@example.com`
- **Password**: `Secret123!`

## Project Architecture

### Authentication Flow

```
┌─────────────┐
│ Login Page  │
└──────┬──────┘
       │
       ▼
┌─────────────────────────┐
│ POST /v1/auth/login     │
│ { email, password }     │
└──────┬──────────────────┘
       │
       ▼
┌────────────────────────────┐
│ Receive Tokens:            │
│ - accessToken              │
│ - refreshToken             │
│ - user object              │
└──────┬─────────────────────┘
       │
       ▼
┌────────────────────────────┐
│ Store in localStorage:     │
│ - admin_token              │
│ - admin_refresh_token      │
│ - admin_user               │
└──────┬─────────────────────┘
       │
       ▼
┌────────────────────────────┐
│ Navigate to /dashboard     │
└────────────────────────────┘
```

### API Request Flow

```
┌──────────────┐
│ Component    │
└──────┬───────┘
       │
       ▼
┌──────────────────┐
│ Service Layer    │
│ (e.g., venueService) │
└──────┬───────────┘
       │
       ▼
┌──────────────────────────┐
│ Axios Instance (api.js)  │
│ - Add Bearer token       │
│ - Add headers            │
└──────┬───────────────────┘
       │
       ▼
┌──────────────────────────┐
│ Backend API              │
│ http://localhost:8080    │
└──────┬───────────────────┘
       │
       ▼
┌──────────────────────────┐
│ Response                 │
│ - Success: Return data   │
│ - 401: Refresh token     │
│ - Error: Show toast      │
└──────────────────────────┘
```

### Component Structure

```
App.jsx (Router + AuthProvider)
├── LoginPage
├── RegisterPage
└── ProtectedRoute
    └── Layout (Sidebar + Header)
        ├── DashboardPage
        ├── VenuesPage
        ├── FacilitiesPage
        ├── BookingsPage
        └── UsersPage
```

## API Endpoints Reference

### Authentication
- `POST /v1/auth/login` - Login
- `POST /v1/auth/register` - Register
- `POST /v1/auth/refresh` - Refresh token

### Venues
- `GET /v1/venues` - List all venues
- `GET /v1/venues/:id` - Get venue by ID
- `POST /v1/venues` - Create venue
- `PUT /v1/venues/:id` - Update venue
- `DELETE /v1/venues/:id` - Delete venue

### Facilities
- `GET /v1/facilities` - List all facilities
- `GET /v1/facilities/:id` - Get facility by ID
- `POST /v1/facilities` - Create facility
- `PUT /v1/facilities/:id` - Update facility
- `DELETE /v1/facilities/:id` - Delete facility

### Bookings
- `GET /v1/bookings` - List all bookings
- `GET /v1/bookings/:id` - Get booking by ID
- `PATCH /v1/bookings/:id/status` - Update status
- `POST /v1/bookings/:id/confirm` - Confirm booking
- `POST /v1/bookings/:id/cancel` - Cancel booking

### Users
- `GET /v1/users` - List all users
- `GET /v1/users/:id` - Get user by ID
- `PATCH /v1/users/:id/activate` - Activate user
- `PATCH /v1/users/:id/deactivate` - Deactivate user

## Common Issues & Solutions

### Issue: "Cannot connect to API"

**Solution 1**: Check backend is running
```bash
# In the root directory
docker-compose ps
# All services should be "Up"
```

**Solution 2**: Verify API URL
```bash
# Test the health endpoint
curl http://localhost:8080/healthz
# Should return: {"status":"ok"}
```

**Solution 3**: Check CORS settings
The backend should allow requests from `http://localhost:3001`

### Issue: "Authentication failed"

**Solution 1**: Clear localStorage
```javascript
// In browser console
localStorage.clear()
location.reload()
```

**Solution 2**: Verify test user exists
```bash
# Run the API test script
./scripts/test-api.sh
# Check if admin@example.com exists
```

**Solution 3**: Check token format
Tokens should be valid JWT format (base64 encoded)

### Issue: "Components not styled correctly"

**Solution**: Ensure Tailwind is compiled
```bash
# Restart the dev server
npm run dev
```

### Issue: "Port 3001 already in use"

**Solution**: Change port in `vite.config.js`
```javascript
server: {
  port: 3002, // Change to available port
}
```

## Development Tips

### Hot Module Replacement (HMR)

Vite provides instant HMR. Changes to:
- `.jsx` files reload instantly
- `.css` files update without page refresh
- State is preserved during updates

### Browser DevTools

**React DevTools**: Install the browser extension for debugging
- View component tree
- Inspect props and state
- Track component re-renders

**Network Tab**: Monitor API calls
- Check request/response headers
- Verify auth tokens are sent
- Debug failed requests

### VS Code Extensions

Recommended extensions:
- ES7+ React/Redux/React-Native snippets
- Tailwind CSS IntelliSense
- ESLint
- Prettier

### Code Formatting

```bash
# Format all files
npm run lint
```

## Production Build

### Build for Production

```bash
npm run build
```

This creates a `dist/` folder with optimized production files:
- Minified JavaScript
- Optimized CSS
- Tree-shaken dependencies
- Asset hashing for cache busting

### Preview Production Build

```bash
npm run preview
```

Opens production build at [http://localhost:4173](http://localhost:4173)

### Deployment Checklist

- [ ] Update `VITE_API_BASE_URL` for production API
- [ ] Enable HTTPS
- [ ] Configure proper CORS on backend
- [ ] Set up authentication persistence
- [ ] Configure error monitoring (e.g., Sentry)
- [ ] Optimize images and assets
- [ ] Test on multiple browsers
- [ ] Test responsive design

## Performance Optimization

### Code Splitting

React Router automatically code-splits routes. Each page loads only when needed.

### Lazy Loading

To lazy load a component:
```javascript
import { lazy, Suspense } from 'react'

const VenuesPage = lazy(() => import('@/pages/VenuesPage'))

// Use with Suspense
<Suspense fallback={<Loading />}>
  <VenuesPage />
</Suspense>
```

### Caching

The app uses:
- localStorage for tokens (persistent)
- React state for component data (session)

## Security Best Practices

### Token Storage

✅ **Current**: localStorage (acceptable for admin panel)
⚠️ **Consider**: httpOnly cookies for production

### XSS Protection

- All user input is sanitized by React
- API responses are validated
- No `dangerouslySetInnerHTML` used

### CSRF Protection

- Backend should implement CSRF tokens
- Use SameSite cookies
- Validate origin headers

## Monitoring & Logging

### Console Logging

The app logs:
- API requests (debug level)
- API responses (debug level)
- Errors (error level)

### Error Boundaries

Add error boundaries for production:
```javascript
// ErrorBoundary.jsx
class ErrorBoundary extends React.Component {
  // Catch and display errors gracefully
}
```

## Testing

### Manual Testing Checklist

- [ ] Login with valid credentials
- [ ] Login with invalid credentials
- [ ] Token refresh on 401
- [ ] Create venue
- [ ] Edit venue
- [ ] Delete venue
- [ ] Create facility
- [ ] Edit facility
- [ ] Delete facility
- [ ] View bookings
- [ ] Update booking status
- [ ] View users
- [ ] Toggle user status
- [ ] Logout

### Browser Compatibility

Tested on:
- ✅ Chrome 120+
- ✅ Firefox 120+
- ✅ Safari 17+
- ✅ Edge 120+

## Additional Resources

- [Vite Documentation](https://vitejs.dev/)
- [React Documentation](https://react.dev/)
- [shadcn/ui Documentation](https://ui.shadcn.com/)
- [Tailwind CSS Documentation](https://tailwindcss.com/)
- [React Router Documentation](https://reactrouter.com/)

## Need Help?

1. Check the [README.md](./README.md) for general information
2. Review the API documentation
3. Check browser console for errors
4. Verify backend services are running
5. Test API endpoints with curl or Postman

---

**Happy Coding! 🚀**
