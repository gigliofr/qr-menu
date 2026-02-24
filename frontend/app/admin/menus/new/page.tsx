'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { FiArrowLeft, FiSave, FiX, FiPlus } from 'react-icons/fi';
import Link from 'next/link';

interface MenuItem {
  id: string;
  name: string;
  description: string;
  price: number;
  image?: string;
  category: string;
}

interface MenuCategory {
  id: string;
  name: string;
  items: MenuItem[];
}

export default function MenuEditorPage() {
  const router = useRouter();
  const [menuName, setMenuName] = useState('');
  const [menuDescription, setMenuDescription] = useState('');
  const [categories, setCategories] = useState<MenuCategory[]>([
    { id: '1', name: 'Appetizers', items: [] },
    { id: '2', name: 'Main Courses', items: [] },
    { id: '3', name: 'Desserts', items: [] },
  ]);
  const [selectedCategory, setSelectedCategory] = useState('1');
  const [editingItem, setEditingItem] = useState<MenuItem | null>(null);
  const [saving, setSaving] = useState(false);

  const addCategory = () => {
    const newId = String(Math.max(...categories.map((c) => Number(c.id)), 0) + 1);
    setCategories([...categories, { id: newId, name: `Category ${newId}`, items: [] }]);
  };

  const updateCategory = (id: string, name: string) => {
    setCategories(categories.map((c) => (c.id === id ? { ...c, name } : c)));
  };

  const deleteCategory = (id: string) => {
    setCategories(categories.filter((c) => c.id !== id));
  };

  const addOrUpdateItem = () => {
    if (!editingItem || !editingItem.name || !editingItem.price) {
      alert('Please fill in all fields');
      return;
    }

    setCategories(
      categories.map((cat) => {
        if (cat.id === selectedCategory) {
          const existing = cat.items.find((i) => i.id === editingItem.id);
          if (existing) {
            return {
              ...cat,
              items: cat.items.map((i) => (i.id === editingItem.id ? editingItem : i)),
            };
          } else {
            return { ...cat, items: [...cat.items, editingItem] };
          }
        }
        return cat;
      })
    );
    setEditingItem(null);
  };

  const deleteItem = (categoryId: string, itemId: string) => {
    setCategories(
      categories.map((cat) =>
        cat.id === categoryId
          ? { ...cat, items: cat.items.filter((i) => i.id !== itemId) }
          : cat
      )
    );
  };

  const startEditingItem = (item: MenuItem) => {
    setSelectedCategory(item.category);
    setEditingItem({ ...item });
  };

  const createNewItem = () => {
    setEditingItem({
      id: String(Date.now()),
      name: '',
      description: '',
      price: 0,
      category: selectedCategory,
    });
  };

  const saveMenu = async () => {
    if (!menuName) {
      alert('Please enter a menu name');
      return;
    }

    setSaving(true);
    try {
      const payload = {
        name: menuName,
        description: menuDescription,
        categories: categories.map((c) => ({
          name: c.name,
          items: c.items,
        })),
      };

      const res = await fetch('/api/menus', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      });

      if (res.ok) {
        router.push('/admin/menus');
      } else {
        alert('Failed to save menu');
      }
    } catch (error) {
      console.error('Failed to save menu:', error);
      alert('Failed to save menu');
    } finally {
      setSaving(false);
    }
  };

  const currentCategory = categories.find((c) => c.id === selectedCategory);

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <Link href="/admin/menus" className="flex items-center gap-2 text-blue-600 hover:text-blue-700">
          <FiArrowLeft /> Back to Menus
        </Link>
        <button
          onClick={saveMenu}
          disabled={saving || !menuName}
          className="btn btn-primary gap-2"
        >
          <FiSave /> {saving ? 'Saving...' : 'Save Menu'}
        </button>
      </div>

      <div className="grid gap-6 lg:grid-cols-4">
        {/* Sidebar - Basic Info */}
        <div className="lg:col-span-1 space-y-4">
          <div className="card">
            <h3 className="mb-4">Menu Details</h3>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium mb-2">Menu Name</label>
                <input
                  type="text"
                  value={menuName}
                  onChange={(e) => setMenuName(e.target.value)}
                  placeholder="e.g., Lunch Menu"
                  className="w-full"
                />
              </div>
              <div>
                <label className="block text-sm font-medium mb-2">Description</label>
                <textarea
                  value={menuDescription}
                  onChange={(e) => setMenuDescription(e.target.value)}
                  placeholder="Optional description"
                  className="w-full h-24"
                />
              </div>
            </div>
          </div>

          {/* Categories */}
          <div className="card">
            <div className="flex items-center justify-between mb-4">
              <h3>Categories</h3>
              <button onClick={addCategory} className="p-2 hover:bg-gray-100 rounded">
                <FiPlus className="h-4 w-4" />
              </button>
            </div>
            <div className="space-y-2">
              {categories.map((cat) => (
                <div
                  key={cat.id}
                  onClick={() => setSelectedCategory(cat.id)}
                  className={`p-3 rounded-lg cursor-pointer transition-colors border-2 ${
                    selectedCategory === cat.id
                      ? 'border-blue-500 bg-blue-50'
                      : 'border-gray-200 hover:bg-gray-50'
                  }`}
                >
                  <input
                    type="text"
                    value={cat.name}
                    onChange={(e) => updateCategory(cat.id, e.target.value)}
                    onClick={(e) => e.stopPropagation()}
                    className="w-full bg-transparent font-medium"
                  />
                  <p className="text-sm text-gray-600 mt-1">{cat.items.length} items</p>
                </div>
              ))}
            </div>
          </div>
        </div>

        {/* Main Area - Items Editor */}
        <div className="lg:col-span-3 space-y-4">
          {/* Item Form */}
          {editingItem && (
            <div className="card border-2 border-blue-500">
              <div className="flex items-center justify-between mb-4">
                <h3>Edit Item</h3>
                <button
                  onClick={() => setEditingItem(null)}
                  className="p-2 hover:bg-gray-100 rounded"
                >
                  <FiX className="h-4 w-4" />
                </button>
              </div>
              <form
                onSubmit={(e) => {
                  e.preventDefault();
                  addOrUpdateItem();
                }}
                className="space-y-4"
              >
                <div className="grid gap-4 md:grid-cols-2">
                  <div>
                    <label className="block text-sm font-medium mb-2">Item Name*</label>
                    <input
                      type="text"
                      value={editingItem.name}
                      onChange={(e) =>
                        setEditingItem({ ...editingItem, name: e.target.value })
                      }
                      placeholder="e.g., Spaghetti Carbonara"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium mb-2">Price*</label>
                    <input
                      type="number"
                      value={editingItem.price}
                      onChange={(e) =>
                        setEditingItem({
                          ...editingItem,
                          price: Number(e.target.value),
                        })
                      }
                      placeholder="0.00"
                      step="0.01"
                    />
                  </div>
                </div>
                <div>
                  <label className="block text-sm font-medium mb-2">Description</label>
                  <textarea
                    value={editingItem.description}
                    onChange={(e) =>
                      setEditingItem({ ...editingItem, description: e.target.value })
                    }
                    placeholder="Item description..."
                    className="h-20"
                  />
                </div>
                <button type="submit" className="btn btn-primary w-full">
                  Save Item
                </button>
              </form>
            </div>
          )}

          {/* Items List */}
          <div className="card">
            <div className="flex items-center justify-between mb-6">
              <h3>{currentCategory?.name} ({currentCategory?.items.length || 0} items)</h3>
              <button onClick={createNewItem} className="btn btn-primary gap-2">
                <FiPlus /> Add Item
              </button>
            </div>

            {!currentCategory?.items.length ? (
              <p className="text-center text-gray-500 py-8">
                No items in this category. <button onClick={createNewItem} className="text-blue-600">Create one</button>
              </p>
            ) : (
              <div className="grid gap-4 md:grid-cols-2">
                {currentCategory?.items.map((item) => (
                  <div key={item.id} className="border border-gray-200 rounded-lg p-4 hover:shadow-md transition-shadow">
                    <div className="flex items-start justify-between mb-2">
                      <div className="flex-1">
                        <h4 className="font-semibold text-gray-900">{item.name}</h4>
                        <p className="text-sm text-gray-600 mt-1">{item.description}</p>
                      </div>
                      <p className="text-lg font-bold text-blue-600 ml-2">${item.price.toFixed(2)}</p>
                    </div>
                    <div className="flex gap-2 mt-4">
                      <button
                        onClick={() => startEditingItem(item)}
                        className="flex-1 btn btn-secondary text-sm"
                      >
                        Edit
                      </button>
                      <button
                        onClick={() => deleteItem(currentCategory!.id, item.id)}
                        className="p-2 text-red-600 hover:bg-red-50 rounded"
                      >
                        <FiTrash2 className="h-4 w-4" />
                      </button>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
