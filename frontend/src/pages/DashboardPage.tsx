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
  access_hash: string;
};
type MessageInfo = {
  id: number;
  date: string;
  text: string;
  sender_name: string;
  is_outgoing: boolean;
};

const typeBadge: Record<string, string> = {
  user: 'bg-green-100 text-green-800',
  group: 'bg-yellow-100 text-yellow-800',
  channel: 'bg-blue-100 text-blue-800',
};

const typeLabel: Record<string, string> = {
  user: '私聊',
  group: '群聊',
  channel: '频道',
};

export default function DashboardPage() {
  const navigate = useNavigate();
  const [user, setUser] = useState<AuthStatus['user']>();
  const [dialogs, setDialogs] = useState<DialogInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedDialog, setSelectedDialog] = useState<DialogInfo | null>(null);
  const [messages, setMessages] = useState<MessageInfo[]>([]);
  const [messagesLoading, setMessagesLoading] = useState(false);

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

  const handleDialogClick = (dialog: DialogInfo) => {
    setSelectedDialog(dialog);
    setMessagesLoading(true);
    setMessages([]);
    rpc<MessageInfo[]>('messages.history', {
      peer_id: dialog.id,
      peer_type: dialog.type,
      access_hash: dialog.access_hash,
      limit: 20,
    })
      .then((msgs) => setMessages(msgs ?? []))
      .catch(() => setMessages([]))
      .finally(() => setMessagesLoading(false));
  };

  if (loading) {
    return <div className="flex items-center justify-center h-64 text-gray-500">加载中...</div>;
  }

  return (
    <div>
      <div className="mb-6">
        <h2 className="text-xl font-bold text-gray-800">仪表盘</h2>
        {user && (
          <p className="text-sm text-gray-500 mt-1">
            已登录：{user.first_name} {user.last_name}
          </p>
        )}
      </div>

      <div className="flex gap-4" style={{ minHeight: '60vh' }}>
        {/* Dialog list */}
        <div className="bg-white rounded-lg shadow w-1/3 flex flex-col overflow-hidden">
          <div className="px-4 py-3 border-b">
            <h3 className="font-medium text-gray-700">对话列表 ({dialogs.length})</h3>
          </div>
          <div className="divide-y overflow-y-auto flex-1">
            {dialogs.map((d) => (
              <div
                key={d.id}
                onClick={() => handleDialogClick(d)}
                className={`px-4 py-3 flex items-center gap-3 cursor-pointer hover:bg-gray-50 transition-colors ${
                  selectedDialog?.id === d.id ? 'bg-blue-50 border-l-2 border-l-blue-600' : ''
                }`}
              >
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <span className="font-medium text-gray-800 truncate">{d.name}</span>
                    <span className={`text-xs px-2 py-0.5 rounded-full ${typeBadge[d.type] ?? 'bg-gray-100 text-gray-600'}`}>
                      {typeLabel[d.type] ?? d.type}
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
              <div className="px-4 py-8 text-center text-gray-400">暂无对话</div>
            )}
          </div>
        </div>

        {/* Message panel */}
        <div className="bg-white rounded-lg shadow w-2/3 flex flex-col overflow-hidden">
          {!selectedDialog ? (
            <div className="flex items-center justify-center h-full text-gray-400">
              选择对话查看消息
            </div>
          ) : (
            <>
              <div className="px-4 py-3 border-b">
                <h3 className="font-medium text-gray-700">{selectedDialog.name}</h3>
                <span className={`text-xs px-2 py-0.5 rounded-full ${typeBadge[selectedDialog.type] ?? 'bg-gray-100 text-gray-600'}`}>
                  {typeLabel[selectedDialog.type] ?? selectedDialog.type}
                </span>
              </div>
              <div className="flex-1 overflow-y-auto p-4 space-y-3">
                {messagesLoading ? (
                  <div className="flex items-center justify-center h-full text-gray-400">加载消息中...</div>
                ) : messages.length === 0 ? (
                  <div className="flex items-center justify-center h-full text-gray-400">暂无消息</div>
                ) : (
                  messages.map((msg) => (
                    <div
                      key={msg.id}
                      className={`flex ${msg.is_outgoing ? 'justify-end' : 'justify-start'}`}
                    >
                      <div
                        className={`max-w-[75%] rounded-lg px-3 py-2 ${
                          msg.is_outgoing
                            ? 'bg-blue-100 text-blue-900'
                            : 'bg-gray-100 text-gray-900'
                        }`}
                      >
                        {!msg.is_outgoing && msg.sender_name && (
                          <div className="text-xs font-semibold text-blue-600 mb-0.5">{msg.sender_name}</div>
                        )}
                        <div className="text-sm whitespace-pre-wrap break-words">{msg.text}</div>
                        <div className="text-xs text-gray-400 mt-1 text-right">{msg.date}</div>
                      </div>
                    </div>
                  ))
                )}
              </div>
            </>
          )}
        </div>
      </div>
    </div>
  );
}
