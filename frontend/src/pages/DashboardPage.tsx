import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { rpc } from '../lib/rpc';

type AuthStatus = { authorized: boolean; user?: { id: number; first_name: string; last_name: string } };
type DialogInfo = {
  id: number;
  name: string;
  type: string;
  unread_count: number;
  last_message: string;
  access_hash: number;
};

const typeBadge: Record<string, string> = {
  user: 'bg-green-100 text-green-800',
  group: 'bg-yellow-100 text-yellow-800',
  channel: 'bg-blue-100 text-blue-800',
};

export default function DashboardPage() {
  const navigate = useNavigate();
  const [user, setUser] = useState<AuthStatus['user']>();
  const [dialogs, setDialogs] = useState<DialogInfo[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    rpc<AuthStatus>('auth.status').then((res) => {
      if (!res.authorized) {
        navigate('/auth', { replace: true });
        return;
      }
      setUser(res.user);
      return rpc<DialogInfo[]>('dialogs.list').then(setDialogs);
    }).catch(() => navigate('/auth', { replace: true }))
      .finally(() => setLoading(false));
  }, [navigate]);

  if (loading) {
    return <div className="flex items-center justify-center h-64 text-gray-500">Loading...</div>;
  }

  return (
    <div>
      <div className="mb-6">
        <h2 className="text-xl font-bold text-gray-800">Dashboard</h2>
        {user && (
          <p className="text-sm text-gray-500 mt-1">
            Logged in as {user.first_name} {user.last_name}
          </p>
        )}
      </div>

      <div className="bg-white rounded-lg shadow">
        <div className="px-4 py-3 border-b">
          <h3 className="font-medium text-gray-700">Dialogs ({dialogs.length})</h3>
        </div>
        <div className="divide-y">
          {dialogs.map((d) => (
            <div key={d.id} className="px-4 py-3 flex items-center gap-3">
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2">
                  <span className="font-medium text-gray-800 truncate">{d.name}</span>
                  <span className={`text-xs px-2 py-0.5 rounded-full ${typeBadge[d.type] ?? 'bg-gray-100 text-gray-600'}`}>
                    {d.type}
                  </span>
                </div>
                {d.last_message && (
                  <p className="text-sm text-gray-500 truncate mt-0.5">{d.last_message}</p>
                )}
              </div>
              {d.unread_count > 0 && (
                <span className="bg-blue-600 text-white text-xs font-medium px-2 py-0.5 rounded-full">
                  {d.unread_count}
                </span>
              )}
            </div>
          ))}
          {dialogs.length === 0 && (
            <div className="px-4 py-8 text-center text-gray-400">No dialogs found</div>
          )}
        </div>
      </div>
    </div>
  );
}
