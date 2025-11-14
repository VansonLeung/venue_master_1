# Venue Master - Admin CMS

A modern, responsive admin content management system built with React, Vite, and shadcn/ui for managing venues, facilities, bookings, and users.

## 🚀 Tech Stack

- **Framework**: React 18
- **Build Tool**: Vite
- **UI Library**: shadcn/ui (Radix UI + Tailwind CSS)
- **Routing**: React Router v6
- **HTTP Client**: Axios
- **Styling**: Tailwind CSS
- **Icons**: Lucide React

## ✨ Features

### Authentication
- **Login**: Secure admin login with JWT tokens
- **Register**: New admin registration
- **Auto Token Refresh**: Automatic token refresh on 401 errors
- **Protected Routes**: Route guards for authenticated access

### Dashboard
- Overview statistics (Venues, Facilities, Bookings, Users)
- Quick action cards
- Responsive design

### Venues Management
- ✅ Create new venues
- ✅ View all venues in table format
- ✅ Edit venue details
- ✅ Delete venues
- Full address management (address, city, state, zip, country)

### Facilities Management
- ✅ Create new facilities
- ✅ View all facilities with availability status
- ✅ Edit facility details
- ✅ Delete facilities
- ✅ Set operating hours
- ✅ Configure weekday/weekend pricing
- ✅ Toggle availability status
- Link facilities to venues

### Bookings Management
- ✅ View all bookings
- ✅ Update booking status
- ✅ Confirm pending bookings
- ✅ Cancel bookings
- Status indicators (Pending, Confirmed, Cancelled, Completed)
- View booking details (facility, time, amount)

### Users Management
- ✅ View all registered users
- ✅ View user roles
- ✅ Activate/Deactivate users
- User status indicators

## 📁 Project Structure

```
frontend_codes/admin_cms/
├── public/
├── src/
│   ├── components/
│   │   ├── ui/              # shadcn/ui components
│   │   ├── Layout.jsx       # Main layout with sidebar
│   │   └── ProtectedRoute.jsx
│   ├── contexts/
│   │   └── AuthContext.jsx  # Authentication context
│   ├── hooks/
│   ├── lib/
│   │   └── utils.js         # Utility functions
│   ├── pages/
│   │   ├── LoginPage.jsx
│   │   ├── RegisterPage.jsx
│   │   ├── DashboardPage.jsx
│   │   ├── VenuesPage.jsx
│   │   ├── FacilitiesPage.jsx
│   │   ├── BookingsPage.jsx
│   │   └── UsersPage.jsx
│   ├── services/
│   │   ├── api.js           # Axios instance with interceptors
│   │   ├── auth.service.js
│   │   ├── venue.service.js
│   │   ├── facility.service.js
│   │   ├── booking.service.js
│   │   └── user.service.js
│   ├── App.jsx              # Main app with routing
│   ├── main.jsx             # Entry point
│   └── index.css            # Global styles
├── .env.example
├── .gitignore
├── index.html
├── package.json
├── postcss.config.js
├── tailwind.config.js
└── vite.config.js
```

## 🛠️ Setup & Installation

### Prerequisites

- Node.js 18+ and npm/yarn
- Backend services running on their respective ports:
  - Gateway: http://localhost:8080
  - Auth Service: http://localhost:8081
  - Booking Service: http://localhost:8083

### Installation Steps

1. **Navigate to the project directory**:
   ```bash
   cd frontend_codes/admin_cms
   ```

2. **Install dependencies**:
   ```bash
   npm install
   ```

3. **Create environment file**:
   ```bash
   cp .env.example .env
   ```

4. **Configure environment variables** (`.env`):
   ```env
   VITE_BASE_URL=http://localhost
   VITE_GATEWAY_PORT=8080
   VITE_AUTH_PORT=8081
   VITE_BOOKING_PORT=8083
   ```

   See [API_CONFIGURATION.md](API_CONFIGURATION.md) for detailed service routing.

5. **Start the development server**:
   ```bash
   npm run dev
   ```

6. **Open your browser**:
   ```
   http://localhost:3001
   ```

## 🔧 Available Scripts

- `npm run dev` - Start development server (port 3001)
- `npm run build` - Build for production
- `npm run preview` - Preview production build
- `npm run lint` - Run ESLint

## 🔑 Test Credentials

**Admin Account**:
- Email: `admin@example.com`
- Password: `Secret123!`

## 📡 API Integration

The admin CMS integrates with backend services on specific ports (following `test-api.sh` structure):

| Service | Port | Endpoints | Usage |
|---------|------|-----------|-------|
| **Gateway** | 8080 | `/v1/venues`, `/v1/users` | Venues & Users management |
| **Auth** | 8081 | `/v1/auth/login`, `/v1/auth/register` | Authentication |
| **Booking** | 8083 | `/v1/facilities`, `/v1/bookings` | Facilities & Bookings |

### Features:
- ✅ Direct service calls (no proxy)
- ✅ JWT token authentication
- ✅ Automatic token refresh on 401 errors
- ✅ Error handling with toast notifications
- ✅ Matches `scripts/test-api.sh` structure

**Detailed documentation**: See [API_CONFIGURATION.md](API_CONFIGURATION.md)

## 🎨 UI Components

The project uses **shadcn/ui** components built on top of:
- **Radix UI**: Unstyled, accessible component primitives
- **Tailwind CSS**: Utility-first CSS framework
- **class-variance-authority**: Type-safe variant management

### Available UI Components

- Button
- Card
- Dialog (Modal)
- Input
- Label
- Select
- Switch
- Table
- Toast (Notifications)

## 🔐 Authentication Flow

1. User enters credentials on login page
2. System calls `/v1/auth/login` endpoint
3. Receives access token and refresh token
4. Tokens stored in localStorage
5. All subsequent API calls include Bearer token
6. On 401 error, automatically refresh token
7. If refresh fails, redirect to login

## 📱 Responsive Design

- ✅ Desktop (1024px+)
- ✅ Tablet (768px - 1023px)
- ✅ Mobile (< 768px)
- Collapsible sidebar on mobile
- Responsive tables
- Touch-friendly buttons

## 🚦 Route Guards

All admin routes are protected with `ProtectedRoute` component:
- Checks authentication status
- Redirects to login if not authenticated
- Shows loading state while checking auth

## 📊 State Management

- **Authentication**: React Context API
- **Component State**: React useState/useEffect
- **API Calls**: Async/await with try-catch

## 🎯 Future Enhancements

- [ ] Advanced filtering and search
- [ ] Bulk operations
- [ ] Export data to CSV/Excel
- [ ] Analytics and reporting
- [ ] Real-time notifications
- [ ] Activity logs
- [ ] Role-based permissions (ADMIN vs SUPER_ADMIN)
- [ ] Dark mode support
- [ ] Multi-language support

## 🐛 Troubleshooting

### Cannot connect to API
- Ensure backend services are running
- Check `VITE_API_BASE_URL` in `.env` file
- Verify CORS is enabled on backend

### Authentication errors
- Clear localStorage: `localStorage.clear()`
- Check token expiration settings
- Verify credentials with backend

### UI components not styled
- Ensure Tailwind CSS is configured
- Check `tailwind.config.js` content paths
- Run `npm install` to install all dependencies

## 📄 License

This project is part of the Venue Master system.

## 🤝 Contributing

1. Follow the existing code structure
2. Use shadcn/ui components where possible
3. Maintain responsive design principles
4. Add error handling for all API calls
5. Test on multiple screen sizes

## 📞 Support

For issues or questions, please refer to the main project documentation or contact the development team.
