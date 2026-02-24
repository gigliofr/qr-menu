'use client';

import { useState } from 'react';
import { FiSave, FiToggleRight } from 'react-icons/fi';

export default function SettingsPage() {
  const [settings, setSettings] = useState({
    appName: 'QR Menu',
    appDescription: 'Enterprise digital menu system',
    maxMenuSize: 100,
    cacheEnabled: true,
    cacheTTL: 3600,
    emailNotifications: true,
    smsNotifications: false,
    darkMode: false,
    language: 'en',
  });

  const [saved, setSaved] = useState(false);

  const handleSave = () => {
    setSaved(true);
    setTimeout(() => setSaved(false), 2000);
  };

  return (
    <div className="space-y-8 max-w-2xl">
      {/* General Settings */}
      <div className="card">
        <h2 className="mb-6">General Settings</h2>
        <div className="space-y-5">
          <div>
            <label className="block text-sm font-medium mb-2">Application Name</label>
            <input
              type="text"
              value={settings.appName}
              onChange={(e) => setSettings({ ...settings, appName: e.target.value })}
            />
          </div>
          <div>
            <label className="block text-sm font-medium mb-2">Description</label>
            <textarea
              value={settings.appDescription}
              onChange={(e) => setSettings({ ...settings, appDescription: e.target.value })}
              rows={3}
            />
          </div>
          <div>
            <label className="block text-sm font-medium mb-2">Language</label>
            <select
              value={settings.language}
              onChange={(e) => setSettings({ ...settings, language: e.target.value })}
            >
              <option value="en">English</option>
              <option value="es">Spanish</option>
              <option value="fr">French</option>
              <option value="it">Italian</option>
              <option value="de">German</option>
            </select>
          </div>
        </div>
      </div>

      {/* Performance Settings */}
      <div className="card">
        <h2 className="mb-6">Performance</h2>
        <div className="space-y-5">
          <div>
            <label className="flex items-center gap-3 cursor-pointer">
              <input
                type="checkbox"
                checked={settings.cacheEnabled}
                onChange={(e) => setSettings({ ...settings, cacheEnabled: e.target.checked })}
                className="w-4 h-4"
              />
              <span className="font-medium">Enable Response Caching</span>
            </label>
            <p className="text-sm text-gray-600 mt-2">
              Cache responses for up to 100x-10,000x faster performance
            </p>
          </div>
          {settings.cacheEnabled && (
            <div>
              <label className="block text-sm font-medium mb-2">Cache TTL (seconds)</label>
              <input
                type="number"
                value={settings.cacheTTL}
                onChange={(e) => setSettings({ ...settings, cacheTTL: Number(e.target.value) })}
                min="60"
                max="86400"
              />
              <p className="text-sm text-gray-600 mt-1">
                Currently set to {Math.floor(settings.cacheTTL / 60)} minutes
              </p>
            </div>
          )}
          <div>
            <label className="block text-sm font-medium mb-2">Max Menu Items</label>
            <input
              type="number"
              value={settings.maxMenuSize}
              onChange={(e) => setSettings({ ...settings, maxMenuSize: Number(e.target.value) })}
              min="10"
              max="1000"
            />
          </div>
        </div>
      </div>

      {/* Notifications */}
      <div className="card">
        <h2 className="mb-6">Notifications</h2>
        <div className="space-y-4">
          <label className="flex items-center gap-3 cursor-pointer">
            <input
              type="checkbox"
              checked={settings.emailNotifications}
              onChange={(e) => setSettings({ ...settings, emailNotifications: e.target.checked })}
              className="w-4 h-4"
            />
            <span className="font-medium">Email Notifications</span>
          </label>
          <label className="flex items-center gap-3 cursor-pointer">
            <input
              type="checkbox"
              checked={settings.smsNotifications}
              onChange={(e) => setSettings({ ...settings, smsNotifications: e.target.checked })}
              className="w-4 h-4"
            />
            <span className="font-medium">SMS Notifications</span>
          </label>
        </div>
      </div>

      {/* Appearance */}
      <div className="card">
        <h2 className="mb-6">Appearance</h2>
        <label className="flex items-center gap-3 cursor-pointer">
          <input
            type="checkbox"
            checked={settings.darkMode}
            onChange={(e) => setSettings({ ...settings, darkMode: e.target.checked })}
            className="w-4 h-4"
          />
          <span className="font-medium">Dark Mode</span>
        </label>
      </div>

      {/* Save Button */}
      <div className="flex items-center gap-4">
        <button onClick={handleSave} className="btn btn-primary gap-2">
          <FiSave /> Save Settings
        </button>
        {saved && <p className="text-green-600 font-medium">✓ Settings saved successfully</p>}
      </div>
    </div>
  );
}
