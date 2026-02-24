import type { Metadata } from 'next';
import './globals.css';

export const metadata: Metadata = {
  title: 'QR Menu - Enterprise Digital Menu System',
  description: 'Professional menu management with QR codes, analytics, and real-time updates',
  viewport: 'width=device-width, initial-scale=1, maximum-scale=5',
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body className="bg-gray-50">
        {children}
      </body>
    </html>
  );
}
