import { NavLink, Outlet } from 'react-router-dom';

const navItems = [
  { to: '/', label: '仪表盘' },
  { to: '/rules', label: '转发规则' },
];

export default function Layout() {
  return (
    <div className="flex h-screen bg-gray-100">
      <aside className="w-56 bg-white shadow-md flex flex-col">
        <div className="p-4 border-b">
          <h1 className="text-lg font-bold text-gray-800">TG 管理器</h1>
        </div>
        <nav className="flex-1 p-2">
          {navItems.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              className={({ isActive }) =>
                `block px-4 py-2 rounded text-sm ${
                  isActive
                    ? 'bg-blue-50 text-blue-700 font-medium'
                    : 'text-gray-600 hover:bg-gray-50'
                }`
              }
            >
              {item.label}
            </NavLink>
          ))}
        </nav>
      </aside>
      <main className="flex-1 overflow-auto p-6">
        <Outlet />
      </main>
    </div>
  );
}
