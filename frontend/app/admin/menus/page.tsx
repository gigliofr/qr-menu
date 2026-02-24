'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { FiPlus, FiEdit, FiTrash2, FiEye, FiCopy, FiDownload } from 'react-icons/fi';

interface Menu {
  id: string;
  name: string;
  description: string;
  itemsCount: number;
  views: number;
  lastModified: string;
  status: 'published' | 'draft';
}

export default function MenusPage() {
  const [menus, setMenus] = useState<Menu[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState('');

  useEffect(() => {
    fetchMenus();
  }, []);

  async function fetchMenus() {
    try {
      const res = await fetch('/api/menus');
      if (res.ok) {
        const data = await res.json();
        setMenus(data || []);
      }
    } catch (error) {
      console.error('Failed to fetch menus:', error);
    } finally {
      setLoading(false);
    }
  }

  async function deleteMenu(id: string) {
    if (!confirm('Are you sure you want to delete this menu?')) return;
    
    try {
      const res = await fetch(`/api/menus/${id}`, { method: 'DELETE' });
      if (res.ok) {
        setMenus(menus.filter((m) => m.id !== id));
      }
    } catch (error) {
      console.error('Failed to delete menu:', error);
    }
  }

  const filteredMenus = menus.filter(
    (menu) =>
      menu.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
      menu.description.toLowerCase().includes(searchTerm.toLowerCase())
  );

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <h1>Menu Management</h1>
        <Link href="/admin/menus/new" className="btn btn-primary gap-2">
          <FiPlus /> Create Menu
        </Link>
      </div>

      {/* Search */}
      <div>
        <input
          type="text"
          placeholder="Search menus..."
          value={searchTerm}
          onChange={(e) => setSearchTerm(e.target.value)}
          className="w-full"
        />
      </div>

      {/* Table */}
      <div className="overflow-x-auto rounded-lg border border-gray-200 bg-white">
        <table className="w-full">
          <thead className="border-b border-gray-200 bg-gray-50">
            <tr>
              <th className="px-6 py-3 text-left font-semibold text-gray-900">Name</th>
              <th className="px-6 py-3 text-left font-semibold text-gray-900">Items</th>
              <th className="px-6 py-3 text-left font-semibold text-gray-900">Views</th>
              <th className="px-6 py-3 text-left font-semibold text-gray-900">Status</th>
              <th className="px-6 py-3 text-left font-semibold text-gray-900">Modified</th>
              <th className="px-6 py-3 text-left font-semibold text-gray-900">Actions</th>
            </tr>
          </thead>
          <tbody>
            {loading ? (
              <tr>
                <td colSpan={6} className="px-6 py-8 text-center text-gray-500">
                  Loading menus...
                </td>
              </tr>
            ) : filteredMenus.length === 0 ? (
              <tr>
                <td colSpan={6} className="px-6 py-8 text-center text-gray-500">
                  No menus found. <Link href="/admin/menus/new" className="text-blue-600">Create one</Link>
                </td>
              </tr>
            ) : (
              filteredMenus.map((menu) => (
                <tr key={menu.id} className="border-b border-gray-100 hover:bg-gray-50">
                  <td className="px-6 py-4">
                    <div>
                      <p className="font-medium text-gray-900">{menu.name}</p>
                      <p className="text-sm text-gray-600">{menu.description}</p>
                    </div>
                  </td>
                  <td className="px-6 py-4 text-gray-900">{menu.itemsCount}</td>
                  <td className="px-6 py-4 text-gray-900">{menu.views.toLocaleString()}</td>
                  <td className="px-6 py-4">
                    <span
                      className={`inline-block rounded-full px-3 py-1 text-sm font-medium ${
                        menu.status === 'published'
                          ? 'bg-green-100 text-green-800'
                          : 'bg-yellow-100 text-yellow-800'
                      }`}
                    >
                      {menu.status}
                    </span>
                  </td>
                  <td className="px-6 py-4 text-sm text-gray-600">{menu.lastModified}</td>
                  <td className="px-6 py-4">
                    <div className="flex gap-2">
                      <Link
                        href={`/admin/menus/${menu.id}`}
                        className="p-2 text-blue-600 hover:bg-blue-50 rounded"
                        title="Edit"
                      >
                        <FiEdit className="h-4 w-4" />
                      </Link>
                      <button
                        className="p-2 text-gray-600 hover:bg-gray-100 rounded"
                        title="Preview"
                      >
                        <FiEye className="h-4 w-4" />
                      </button>
                      <button
                        className="p-2 text-gray-600 hover:bg-gray-100 rounded"
                        title="Download QR"
                      >
                        <FiDownload className="h-4 w-4" />
                      </button>
                      <button
                        onClick={() => deleteMenu(menu.id)}
                        className="p-2 text-red-600 hover:bg-red-50 rounded"
                        title="Delete"
                      >
                        <FiTrash2 className="h-4 w-4" />
                      </button>
                    </div>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
