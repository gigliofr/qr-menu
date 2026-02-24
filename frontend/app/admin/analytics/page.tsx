'use client';

import { useState } from 'react';
import {
  LineChart,
  Line,
  AreaChart,
  Area,
  BarChart,
  Bar,
  PieChart,
  Pie,
  Cell,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from 'recharts';
import { FiDownload, FiFilter } from 'react-icons/fi';

export default function AnalyticsPage() {
  const [dateRange, setDateRange] = useState('7d');

  const viewsData = [
    { date: 'Mon', views: 2400, unique: 1240 },
    { date: 'Tue', views: 1398, unique: 1221 },
    { date: 'Wed', views: 9800, unique: 2290 },
    { date: 'Thu', views: 3908, unique: 2000 },
    { date: 'Fri', views: 4800, unique: 2181 },
    { date: 'Sat', views: 3800, unique: 2500 },
    { date: 'Sun', views: 4300, unique: 2100 },
  ];

  const topMenus = [
    { name: 'Pasta & Pizza', views: 3200, conversions: 485, rate: 15.2 },
    { name: 'Breakfast Menu', views: 2100, conversions: 298, rate: 14.2 },
    { name: 'Desserts', views: 1800, conversions: 216, rate: 12.0 },
    { name: 'Beverages', views: 1200, conversions: 168, rate: 14.0 },
  ];

  const deviceData = [
    { name: 'Mobile', value: 65, fill: '#3b82f6' },
    { name: 'Desktop', value: 28, fill: '#10b981' },
    { name: 'Tablet', value: 7, fill: '#f59e0b' },
  ];

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex items-center justify-between">
        <h1>Analytics</h1>
        <div className="flex gap-4">
          <select
            value={dateRange}
            onChange={(e) => setDateRange(e.target.value)}
            className="btn btn-secondary"
          >
            <option value="7d">Last 7 days</option>
            <option value="30d">Last 30 days</option>
            <option value="90d">Last 90 days</option>
            <option value="1y">Last year</option>
          </select>
          <button className="btn btn-primary gap-2">
            <FiDownload /> Export Report
          </button>
        </div>
      </div>

      {/* KPI Cards */}
      <div className="grid gap-6 md:grid-cols-4">
        <KPICard label="Total Views" value="45.2K" change="+12.5%" />
        <KPICard label="Unique Visitors" value="12.8K" change="+8.2%" />
        <KPICard label="Conversions" value="3.2K" change="+23.1%" />
        <KPICard label="Avg. Session" value="3:45" change="+2m 12s" />
      </div>

      {/* Charts */}
      <div className="grid gap-6 lg:grid-cols-2">
        {/* Views Over Time */}
        <div className="card">
          <h3 className="mb-6">Views Over Time</h3>
          <ResponsiveContainer width="100%" height={300}>
            <AreaChart data={viewsData}>
              <defs>
                <linearGradient id="colorViews" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="5%" stopColor="#3b82f6" stopOpacity={0.3} />
                  <stop offset="95%" stopColor="#3b82f6" stopOpacity={0} />
                </linearGradient>
              </defs>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="date" />
              <YAxis />
              <Tooltip />
              <Area
                type="monotone"
                dataKey="views"
                stroke="#3b82f6"
                fillOpacity={1}
                fill="url(#colorViews)"
              />
            </AreaChart>
          </ResponsiveContainer>
        </div>

        {/* Device Breakdown */}
        <div className="card">
          <h3 className="mb-6">Device Breakdown</h3>
          <ResponsiveContainer width="100%" height={300}>
            <PieChart>
              <Pie
                data={deviceData}
                cx="50%"
                cy="50%"
                labelLine={false}
                label={({ name, value }) => `${name}: ${value}%`}
                outerRadius={100}
                fill="#8884d8"
                dataKey="value"
              >
                {deviceData.map((entry, index) => (
                  <Cell key={`cell-${index}`} fill={entry.fill} />
                ))}
              </Pie>
              <Tooltip />
            </PieChart>
          </ResponsiveContainer>
        </div>
      </div>

      {/* Top Menus */}
      <div className="card">
        <h3 className="mb-6">Top Performing Menus</h3>
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead className="border-b border-gray-200 bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left font-semibold text-gray-900">Menu</th>
                <th className="px-6 py-3 text-left font-semibold text-gray-900">Views</th>
                <th className="px-6 py-3 text-left font-semibold text-gray-900">Conversions</th>
                <th className="px-6 py-3 text-left font-semibold text-gray-900">Conv. Rate</th>
              </tr>
            </thead>
            <tbody>
              {topMenus.map((menu, idx) => (
                <tr key={idx} className="border-b border-gray-100 hover:bg-gray-50">
                  <td className="px-6 py-4 font-medium text-gray-900">{menu.name}</td>
                  <td className="px-6 py-4 text-gray-900">{menu.views.toLocaleString()}</td>
                  <td className="px-6 py-4 text-gray-900">{menu.conversions.toLocaleString()}</td>
                  <td className="px-6 py-4">
                    <span className="inline-block rounded-full bg-green-100 px-3 py-1 text-sm font-medium text-green-800">
                      {menu.rate.toFixed(1)}%
                    </span>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      {/* Hourly Distribution */}
      <div className="card">
        <h3 className="mb-6">Views by Hour of Day</h3>
        <ResponsiveContainer width="100%" height={300}>
          <BarChart data={Array.from({ length: 24 }, (_, i) => ({
            hour: `${i}:00`,
            views: Math.floor(Math.random() * 500) + 100,
          }))}>
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis dataKey="hour" />
            <YAxis />
            <Tooltip />
            <Bar dataKey="views" fill="#3b82f6" />
          </BarChart>
        </ResponsiveContainer>
      </div>
    </div>
  );
}

function KPICard({ label, value, change }: any) {
  return (
    <div className="card">
      <p className="text-gray-600 mb-2">{label}</p>
      <div className="flex items-baseline gap-2">
        <p className="text-3xl font-bold text-gray-900">{value}</p>
        <p className="text-sm text-green-600">{change}</p>
      </div>
    </div>
  );
}
