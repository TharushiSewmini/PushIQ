# PushIQ React Dashboard

**Status:** ✅ Complete (Component structure)
**Framework:** React 18 + Vite
**Styling:** Tailwind CSS
**State Management:** React Hooks + localStorage

## Project Structure

```
frontend/
├── src/
│   ├── components/          # Reusable UI components
│   │   ├── Header.jsx       # Top navigation bar
│   │   ├── Sidebar.jsx      # Left navigation menu
│   │   ├── MetricsCard.jsx  # Stats display card
│   │   ├── DeliveryChart.jsx    # Hourly trends chart
│   │   ├── FunnelChart.jsx      # Funnel visualization
│   │   └── PlatformBreakdown.jsx # Platform stats
│   ├── pages/               # Page components
│   │   ├── Dashboard.jsx    # Main analytics dashboard
│   │   ├── Notifications.jsx    # Notifications management
│   │   ├── Analytics.jsx    # Detailed analytics
│   │   └── Settings.jsx     # User settings & logout
│   ├── App.jsx              # Root component w/ routing
│   ├── main.jsx             # Entry point
│   └── index.css            # Global styles
├── index.html               # HTML template
├── package.json             # Dependencies
├── vite.config.js           # Vite configuration
└── .gitignore               # Git ignore rules
```

## Features Implemented

### Authentication
- Simple API key login interface
- Secure token storage in localStorage
- Auto-redirect to login when session expires

### Dashboard (Main Page)
- **Key Metrics** - Total sent, delivered, failed, online devices
- **Hourly Trends** - Bar chart visualization of notification volume
- **Platform Breakdown** - iOS/Android split with delivery rates
- **Delivery Funnel** - Step-by-step user journey from devices to deliveries
- **Top Notifications** - List of most impactful campaigns

### Navigation
- Responsive sidebar with menu items
- Mobile-friendly hamburger menu
- Active page highlighting
- Settings & logout functionality

### API Integration
- Fetches data from `http://localhost:8080/api/v1/analytics/*`
- Automatic data refresh on page load
- Manual refresh button
- Error handling with retry option
- Loading states

## Installation & Running

```bash
# Install dependencies
npm install

# Start development server (port 3000)
npm run dev

# Build for production
npm run build

# Preview production build
npm run preview
```

## Environment Setup

Dashboard expects backend running on:
- URL: `http://localhost:8080`
- API endpoints: `/api/v1/analytics/*`
- Authentication: `X-Api-Key` header

## Features by Page

### Dashboard (`/`)
- Real-time analytics overview
- Notification delivery metrics
- Device presence analysis
- Hourly trend visualization
- Fastest growing notifications

### Notifications (`/notifications`)
Placeholder for future functionality:
- Create new campaigns
- View sent notifications
- Schedule notifications

### Analytics (`/analytics`)
Placeholder for future functionality:
- Advanced filtering
- Custom date ranges
- Detailed breakdowns

### Settings (`/settings`)
- Logout functionality
- API key management (future)

## Styling Approach

- **Framework:** Tailwind CSS (via CDN for simplicity)
- **Design System:** Cohesive color scheme (blue brand color)
- **Responsiveness:** Mobile-first, breakpoints at md/lg
- **Components:** Reusable card, button, chart styles

## Future Enhancements

1. **Real-time Updates**: WebSocket integration for live metrics
2. **Advanced Charts**: Chart.js or Recharts for prettier visualizations
3. **Data Export**: CSV/PDF export of analytics
4. **Custom Date Ranges**: Date picker for flexible filtering
5. **Dark Mode**: Theme toggle for dark/light modes
6. **User Management**: Team collaboration features
7. **Notifications**: Toast notifications for errors and successes
8. **Performance**: Code splitting, lazy loading, caching strategies

## Notes

- Dashboard is fully responsive and works on mobile
- API calls include proper error handling
- Component architecture supports easy extensibility
- Styling can be enhanced with Chart.js or similar libraries
- No external UI framework dependencies (pure Tailwind)
