'use client';

import { ReactNode } from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import {
  FiHome,
  FiMenu,
  FiBarChart2,
  FiSettings,
  FiUsers,
  FiLogOut,
  FiMenu as FiMenuIcon,
} from 'react-icons/fi';
import { useState } from 'react';

export default function AdminLayout({ children }: { children: ReactNode }) {
  const pathname = usePathname();
  const [sidebarOpen, setSidebarOpen] = useState(true);

  const isActive = (path: string) => pathname.startsWith(path);

  const navItems = [
    { href: '/admin/dashboard', label: 'Dashboard', icon: FiHome },
    { href: '/admin/menus', label: 'Menus', icon: FiMenu },
    { href: '/admin/analytics', label: 'Analytics', icon: FiBarChart2 },
    { href: '/admin/users', label: 'Users', icon: FiUsers },
    { href: '/admin/settings', label: 'Settings', icon: FiSettings },
  ];

  return (
    <div className="flex h-screen bg-gray-100">
      {/* Sidebar */}
      <div
        className={`${
          sidebarOpen ? 'w-64' : 'w-20'
        } border-r border-gray-200 bg-white transition-all duration-300 flex flex-col`}
      >
        <div className="flex items-center justify-between p-4 border-b border-gray-200">
          <h1 className={`font-bold text-blue-600 ${!sidebarOpen && 'hidden'}`}>
            QR Menu
          </h1>
          <button
            onClick={() => setSidebarOpen(!sidebarOpen)}
            className="p-2 hover:bg-gray-100 rounded-lg transition-colors"
          >
            <FiMenuIcon className="h-5 w-5" />
          </button>
        </div>

        <nav className="flex-1 px-2 py-4 space-y-2">
          {navItems.map((item) => {
            const Icon = item.icon;
            return (
              <Link
                key={item.href}
                href={item.href}
                className={`flex items-center gap-3 px-3 py-2 rounded-lg transition-colors ${
                  isActive(item.href)
                    ? 'bg-blue-100 text-blue-600 font-semibold'
                    : 'text-gray-700 hover:bg-gray-100'
                }`}
              >
                <Icon className="h-5 w-5 flex-shrink-0" />
                <span className={!sidebarOpen ? 'hidden' : ''}>{item.label}</span>
              </Link>
            );
          })}
        </nav>

        <div className="border-t border-gray-200 p-4">
          <button
            className={`flex items-center gap-3 px-3 py-2 rounded-lg text-red-600 hover:bg-red-50 w-full transition-colors ${
              !sidebarOpen && 'justify-center'
            }`}
          >
            <FiLogOut className="h-5 w-5" />
            <span className={!sidebarOpen ? 'hidden' : ''}>Logout</span>
          </button>
        </div>
      </div>

      {/* Main Content */}
      <div className="flex-1 flex flex-col overflow-hidden">
        {/* Top Bar */}
        <div className="border-b border-gray-200 bg-white px-8 py-4 flex items-center justify-between">
          <h2 className="text-2xl font-semibold text-gray-800">
            {navItems.find((item) => isActive(item.href))?.label || 'Admin'}
          </h2>
          <div className="flex items-center gap-4">
            <div className="text-sm text-gray-600">Logged in as Admin</div>
          </div>
        </div>

        {/* Page Content */}
        <div className="flex-1 overflow-auto">
          <div className="p-8">{children}</div>
        </div>
      </div>
    </div>
  );
}
