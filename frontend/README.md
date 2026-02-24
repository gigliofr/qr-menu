# QR Menu Frontend

Modern React 18 frontend for the QR Menu enterprise digital menu system.

## 📋 Features

- **Admin Dashboard**: Real-time analytics and performance metrics
- **Menu Builder**: Drag-and-drop menu editor with live preview
- **Responsive Design**: Mobile-first, fully responsive UI
- **Dark Mode**: Optional dark theme support
- **Analytics**: Detailed menu views, conversions, device tracking
- **User Management**: Role-based access control (RBAC)
- **Performance**: Optimized with Next.js and caching

## 🚀 Quick Start

### Prerequisites

- Node.js 18+
- npm 9+
- Go backend running on `http://localhost:8080`

### Installation

```bash
# Install dependencies
npm install

# Create environment file
cp .env.example .env.local

# Start development server
npm run dev
```

Visit http://localhost:3000 in your browser.

### Admin Access

- **Dashboard**: http://localhost:3000/admin/dashboard
- **Menu Editor**: http://localhost:3000/admin/menus
- **Analytics**: http://localhost:3000/admin/analytics
- **Users**: http://localhost:3000/admin/users
- **Settings**: http://localhost:3000/admin/settings

## 📁 Project Structure

```
frontend/
├── app/
│   ├── admin/              # Admin panel pages
│   │   ├── dashboard/      # Dashboard page
│   │   ├── menus/          # Menu management
│   │   ├── analytics/      # Analytics dashboard
│   │   ├── users/          # User management
│   │   └── settings/       # Settings page
│   ├── globals.css         # Global styles
│   ├── layout.tsx          # Root layout
│   └── page.tsx            # Home page
├── public/                 # Static assets
├── package.json           # Dependencies
├── tsconfig.json          # TypeScript config
├── next.config.js         # Next.js config
└── tailwind.config.js     # Tailwind CSS config
```

## 🛠️ Available Scripts

```bash
# Development
npm run dev              # Start dev server with hot reload

# Production
npm run build            # Build for production
npm start                # Start production server

# Code Quality
npm run lint             # Run ESLint
npm run type-check       # Run TypeScript type checking
npm run format           # Format code with Prettier

# Testing
npm test                 # Run tests
npm run test:watch       # Watch mode
npm run test:coverage    # Coverage report
```

## 🎨 Styling

- **Tailwind CSS**: Utility-first CSS framework
- **Custom Theme**: CSS variables in `globals.css`
- **Dark Mode**: Class-based dark mode support
- **Responsive**: Mobile-first responsive design

### Theme Colors

```css
--primary: #2563eb       /* Blue */
--secondary: #64748b     /* Slate */
--success: #10b981       /* Green */
--warning: #f59e0b       /* Amber */
--error: #ef4444         /* Red */
```

## 🔗 API Integration

Frontend connects to Go backend at `NEXT_PUBLIC_API_URL` (default: `http://localhost:8080`)

### Key Endpoints

```
GET  /api/health              # Health check
GET  /api/menus               # List menus
POST /api/menus               # Create menu
GET  /api/menus/:id           # Get menu
PUT  /api/menus/:id           # Update menu
DELETE /api/menus/:id         # Delete menu

GET  /api/admin/dashboard/stats    # Dashboard stats
GET  /api/cache/stats             # Cache statistics

GET  /api/analytics               # Analytics data
GET  /api/analytics/:menu_id      # Menu analytics
```

## 📊 Page Components

### Dashboard
- KPI cards (menus, views, users, response time)
- Cache performance metrics
- Chart visualizations (line, bar)
- Recent activity feed

### Menu Editor
- Create/edit menus
- Add/remove categories
- Add/remove menu items
- Drag-and-drop support (extensible)
- Price and description management

### Analytics
- Views over time (area chart)
- Device breakdown (pie chart)
- Top performing menus (table)
- Hourly distribution (bar chart)
- Customizable date range

### User Management
- User list with search
- Role-based badges (admin, manager, staff)
- User status indicators
- Edit/delete actions

### Settings
- General configuration
- Performance (caching, TTL)
- Notifications (email, SMS)
- Appearance (dark mode)

## 🔒 Security

- Secure API communication (HTTPS in production)
- CSRF protection via Next.js
- XSS protection via React's built-in escaping
- Secure headers configured in `next.config.js`
- Environment variables validation

## 📦 Dependencies

### Core
- `react`: 18.3.1 - UI library
- `next`: 14.1.0 - React framework
- `typescript`: 5.3.3 - Type safety

### UI & Visualization
- `react-icons`: Icons library
- `recharts`: Chart library
- `tailwindcss`: CSS framework
- `framer-motion`: Animations
- `react-hot-toast`: Notifications

### State & Data
- `zustand`: State management
- `axios`: HTTP client
- `zod`: Data validation
- `date-fns`: Date utilities

### Features
- `react-dnd`: Drag and drop
- `js-qr`: QR code generation
- `qrcode.react`: QR code component

## 🚢 Deployment

### Build for Production

```bash
npm run build
npm start
```

### Docker

```dockerfile
FROM node:18-alpine
WORKDIR /app
COPY . .
RUN npm install && npm run build
EXPOSE 3000
CMD ["npm", "start"]
```

### Environment Variables for Production

```bash
NEXT_PUBLIC_API_URL=https://api.example.com
NEXT_PUBLIC_ENVIRONMENT=production
```

## 📝 Environment Setup

Copy `.env.example` to `.env.local`:

```bash
cp .env.example .env.local
```

Then update with your values.

## 🐛 Troubleshooting

### Port 3000 already in use
```bash
npx kill-port 3000  # Or specify PORT=3001 npm run dev
```

### API connection failed
- Ensure Go backend is running on `localhost:8080`
- Check `NEXT_PUBLIC_API_URL` in `.env.local`
- Verify CORS configuration in Go backend

### Build fails
```bash
rm -rf node_modules .next
npm install
npm run build
```

## 🤝 Contributing

See [CONTRIBUTING.md](../CONTRIBUTING.md) in root directory.

## 📄 License

Part of QR Menu v2.0.0 - Enterprise Digital Menu System

---

**Frontend Version**: 2.0.0  
**Last Updated**: February 24, 2026  
**Next.js Version**: 14.1+  
**Node Version**: 18.0+
