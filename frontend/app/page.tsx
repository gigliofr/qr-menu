'use client';

import Link from 'next/link';
import { FiMenu3, FiBarChart3, FiShield, FiBell, FiDownload } from 'react-icons/fi';

export default function HomePage() {
  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 via-white to-purple-50">
      {/* Navigation */}
      <nav className="sticky top-0 z-50 border-b border-gray-200 bg-white/80 backdrop-blur">
        <div className="container-xl flex h-16 items-center justify-between">
          <div className="text-2xl font-bold text-blue-600">QR Menu</div>
          <div className="flex gap-4">
            <Link href="/admin/dashboard" className="btn btn-primary">
              Admin Dashboard
            </Link>
            <Link href="/api/health" className="btn btn-secondary">
              API Health
            </Link>
          </div>
        </div>
      </nav>

      {/* Hero Section */}
      <section className="container-xl py-20 text-center">
        <h1 className="mb-6 bg-gradient-to-r from-blue-600 to-purple-600 bg-clip-text text-transparent">
          Enterprise Digital Menu System
        </h1>
        <p className="mx-auto mb-8 max-w-2xl text-xl text-gray-600">
          Professional QR code menu management with real-time analytics, fast caching, and enterprise-grade security.
        </p>
        <div className="flex justify-center gap-4">
          <Link href="/admin/dashboard" className="btn btn-primary px-8 py-3 text-lg">
            Get Started
          </Link>
          <a href="/docs" className="btn btn-secondary px-8 py-3 text-lg">
            View Documentation
          </a>
        </div>
      </section>

      {/* Features Grid */}
      <section className="container-xl py-20">
        <h2 className="mb-12 text-center">Key Features</h2>
        <div className="grid gap-8 md:grid-cols-2 lg:grid-cols-3">
          <FeatureCard
            icon={<FiMenu3 className="h-8 w-8" />}
            title="Menu Management"
            description="Create and manage digital menus with categories, items, pricing, and images. Real-time updates across all devices."
          />
          <FeatureCard
            icon={<FiBarChart3 className="h-8 w-8" />}
            title="Analytics Dashboard"
            description="Track menu views, popular items, peak hours, and customer engagement with interactive visualizations."
          />
          <FeatureCard
            icon={<FiShield className="h-8 w-8" />}
            title="Enterprise Security"
            description="Multi-layer authentication, role-based access control, and audit logs for compliance requirements."
          />
          <FeatureCard
            icon={<FiBell className="h-8 w-8" />}
            title="Real-Time Updates"
            description="WebSocket support for instant menu updates and notifications across all connected devices."
          />
          <FeatureCard
            icon={<FiDownload className="h-8 w-8" />}
            title="Performance Optimized"
            description="100x-10,000x faster with response caching and query result caching. Sub-100ms response times."
          />
          <FeatureCard
            icon={<FiMenu3 className="h-8 w-8" />}
            title="Multi-Language"
            description="Support 50+ languages with automatic translation and localization. RTL language support."
          />
        </div>
      </section>

      {/* Tech Stack */}
      <section className="container-xl py-20">
        <h2 className="mb-12 text-center">Technology Stack</h2>
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-4">
          <TechCard name="Go 1.24+" desc="High-performance backend" />
          <TechCard name="React 18" desc="Modern frontend" />
          <TechCard name="PostgreSQL" desc="Reliable database" />
          <TechCard name="Redis" desc="Caching layer" />
        </div>
      </section>

      {/* Performance Stats */}
      <section className="container-xl py-20 text-center">
        <h2 className="mb-12">Performance</h2>
        <div className="grid gap-8 md:grid-cols-3">
          <StatCard number="100x" label="Faster with caching" />
          <StatCard number="61+" label="Integration tests" />
          <StatCard number="0ms" label="Cold start time" />
        </div>
      </section>

      {/* Footer */}
      <footer className="border-t border-gray-200 bg-gray-50 py-12">
        <div className="container-xl text-center text-gray-600">
          <p className="mb-4">QR Menu v2.0.0</p>
          <div className="flex justify-center gap-6">
            <Link href="/docs" className="hover:text-gray-900">
              Documentation
            </Link>
            <Link href="/api/health" className="hover:text-gray-900">
              API Status
            </Link>
            <Link href="/admin" className="hover:text-gray-900">
              Admin
            </Link>
          </div>
        </div>
      </footer>
    </div>
  );
}

function FeatureCard({ icon, title, description }: any) {
  return (
    <div className="card group hover:shadow-lg">
      <div className="mb-4 inline-flex rounded-lg bg-blue-100 p-3 text-blue-600">
        {icon}
      </div>
      <h3 className="mb-2">{title}</h3>
      <p className="text-gray-600">{description}</p>
    </div>
  );
}

function TechCard({ name, desc }: any) {
  return (
    <div className="card text-center">
      <h4 className="mb-2">{name}</h4>
      <p className="text-sm text-gray-600">{desc}</p>
    </div>
  );
}

function StatCard({ number, label }: any) {
  return (
    <div className="card">
      <div className="text-4xl font-bold text-blue-600">{number}</div>
      <p className="text-gray-600">{label}</p>
    </div>
  );
}
