'use client';

import { useEffect, useState } from 'react';
import { FiTrendingUp, FiUsers, FiMenu, FiBarChart3 } from 'react-icons/fi';
import {
  LineChart,
  Line,
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from 'recharts';

interface DashboardStats {
  totalMenus: number;
  totalViews: number;
  activeUsers: number;
  avgResponseTime: number;
}

interface CacheStats {
  hitRate: number;
  totalHits: number;
  totalMisses: number;
  avgLatency: number;
}

export default function DashboardPage() {
  const [stats, setStats] = useState<DashboardStats>({
    totalMenus: 0,
    totalViews: 0,
    activeUsers: 0,
    avgResponseTime: 0,
  });
  const [cacheStats, setCacheStats] = useState<CacheStats>({
    hitRate: 0,
    totalHits: 0,
    totalMisses: 0,
    avgLatency: 0,
  });
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchDashboardData();
    const interval = setInterval(fetchDashboardData, 30000); // Refresh every 30s
    return () => clearInterval(interval);
  }, []);

  async function fetchDashboardData() {
    try {
      // Fetch from Go backend
      const [statsRes, cacheRes] = await Promise.all([
        fetch('/api/admin/dashboard/stats'),
        fetch('/api/cache/stats'),
      ]);

      if (statsRes.ok) {
        const data = await statsRes.json();
        setStats(data);
      }

      if (cacheRes.ok) {
        const data = await cacheRes.json();
        setCacheStats(data);
      }
    } catch (error) {
      console.error('Failed to fetch dashboard data:', error);
    } finally {
      setLoading(false);
    }
  }

  const chartData = [
    { name: 'Mon', views: 2400, conversions: 240 },
    { name: 'Tue', views: 1398, conversions: 221 },
    { name: 'Wed', views: 9800, conversions: 229 },
    { name: 'Thu', views: 3908, conversions: 200 },
    { name: 'Fri', views: 4800, conversions: 218 },
    { name: 'Sat', views: 3800, conversions: 250 },
    { name: 'Sun', views: 4300, conversions: 210 },
  ];

  return (
    <div className="space-y-8">
      {/* Stats Grid */}
      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-4">
        <StatCard
          icon={<FiMenu className="h-8 w-8" />}
          label="Total Menus"
          value={stats.totalMenus}
          change="+12% from last month"
        />
        <StatCard
          icon={<FiTrendingUp className="h-8 w-8" />}
          label="Total Views"
          value={stats.totalViews.toLocaleString()}
          change="+23% from last week"
        />
        <StatCard
          icon={<FiUsers className="h-8 w-8" />}
          label="Active Users"
          value={stats.activeUsers}
          change="+5% from yesterday"
        />
        <StatCard
          icon={<FiBarChart3 className="h-8 w-8" />}
          label="Avg Response"
          value={`${stats.avgResponseTime}ms`}
          change="↓ 45% with cache"
        />
      </div>

      {/* Cache Performance */}
      <div className="card">
        <h3 className="mb-6">Cache Performance</h3>
        <div className="grid gap-4 md:grid-cols-4">
          <div className="rounded-lg bg-green-50 p-4">
            <div className="text-sm text-gray-600">Hit Rate</div>
            <div className="text-3xl font-bold text-green-600">{(cacheStats.hitRate * 100).toFixed(1)}%</div>
          </div>
          <div className="rounded-lg bg-blue-50 p-4">
            <div className="text-sm text-gray-600">Total Hits</div>
            <div className="text-3xl font-bold text-blue-600">{cacheStats.totalHits.toLocaleString()}</div>
          </div>
          <div className="rounded-lg bg-yellow-50 p-4">
            <div className="text-sm text-gray-600">Total Misses</div>
            <div className="text-3xl font-bold text-yellow-600">{cacheStats.totalMisses.toLocaleString()}</div>
          </div>
          <div className="rounded-lg bg-purple-50 p-4">
            <div className="text-sm text-gray-600">Avg Latency</div>
            <div className="text-3xl font-bold text-purple-600">{cacheStats.avgLatency}µs</div>
          </div>
        </div>
      </div>

      {/* Charts */}
      <div className="grid gap-6 lg:grid-cols-2">
        {/* Line Chart */}
        <div className="card">
          <h3 className="mb-6">Menu Views Trend</h3>
          <ResponsiveContainer width="100%" height={300}>
            <LineChart data={chartData}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="name" />
              <YAxis />
              <Tooltip />
              <Legend />
              <Line
                type="monotone"
                dataKey="views"
                stroke="#2563eb"
                strokeWidth={2}
                dot={{ fill: '#2563eb' }}
              />
            </LineChart>
          </ResponsiveContainer>
        </div>

        {/* Bar Chart */}
        <div className="card">
          <h3 className="mb-6">Conversions by Day</h3>
          <ResponsiveContainer width="100%" height={300}>
            <BarChart data={chartData}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="name" />
              <YAxis />
              <Tooltip />
              <Legend />
              <Bar dataKey="conversions" fill="#10b981" />
            </BarChart>
          </ResponsiveContainer>
        </div>
      </div>

      {/* Recent Activity */}
      <div className="card">
        <h3 className="mb-6">Recent Activity</h3>
        <div className="space-y-4">
          {[
            { time: '2 hours ago', action: 'Menu "Pasta" updated', user: 'Admin' },
            { time: '4 hours ago', action: 'New menu "Desserts" created', user: 'Admin' },
            { time: '1 day ago', action: 'User "John Doe" registered', user: 'System' },
            { time: '2 days ago', action: 'Cache cleared', user: 'Admin' },
          ].map((activity, idx) => (
            <div
              key={idx}
              className="flex items-center justify-between border-b border-gray-100 pb-4 last:border-0"
            >
              <div>
                <p className="font-medium text-gray-900">{activity.action}</p>
                <p className="text-sm text-gray-500">{activity.time}</p>
              </div>
              <span className="inline-block rounded-full bg-gray-100 px-3 py-1 text-sm text-gray-600">
                {activity.user}
              </span>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

function StatCard({ icon, label, value, change }: any) {
  return (
    <div className="card">
      <div className="flex items-center justify-between">
        <div>
          <p className="text-gray-600">{label}</p>
          <p className="text-3xl font-bold text-gray-900">{value}</p>
          <p className="text-sm text-green-600">{change}</p>
        </div>
        <div className="text-4xl text-blue-600 opacity-20">{icon}</div>
      </div>
    </div>
  );
}
