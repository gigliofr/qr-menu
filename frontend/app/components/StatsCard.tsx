interface StatsCardProps {
  label: string;
  value: string | number;
  change?: string;
  icon?: React.ReactNode;
}

export default function StatsCard({ label, value, change, icon }: StatsCardProps) {
  return (
    <div className="card">
      <div className="flex items-center justify-between">
        <div>
          <p className="text-gray-600">{label}</p>
          <p className="text-3xl font-bold text-gray-900">{value}</p>
          {change && <p className="text-sm text-green-600">{change}</p>}
        </div>
        {icon && <div className="text-4xl text-blue-600 opacity-20">{icon}</div>}
      </div>
    </div>
  );
}
