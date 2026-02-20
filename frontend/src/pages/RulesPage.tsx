import { useState, useEffect, useCallback } from 'react';
import { rpc } from '../lib/rpc';

type ForwardRule = {
  id: number;
  source_channel_id: number;
  source_name: string;
  source_hash: string;
  target_channel_id: number;
  target_name: string;
  target_hash: string;
  match_pattern: string;
  enabled: boolean;
  created_at: string;
  updated_at: string;
};

type ChannelInfo = {
  id: number;
  name: string;
  type: string;
  access_hash: string;
};

const typeLabel: Record<string, string> = {
  channel: '频道',
  group: '群聊',
  user: '私聊',
};

export default function RulesPage() {
  const [rules, setRules] = useState<ForwardRule[]>([]);
  const [channels, setChannels] = useState<ChannelInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const [showForm, setShowForm] = useState(false);
  const [editingRule, setEditingRule] = useState<ForwardRule | null>(null);
  const [error, setError] = useState('');

  const [sourceId, setSourceId] = useState('');
  const [targetId, setTargetId] = useState('');
  const [matchPattern, setMatchPattern] = useState('');

  const loadData = useCallback(async () => {
    try {
      const [r, c] = await Promise.all([
        rpc<ForwardRule[]>('rules.list'),
        rpc<ChannelInfo[]>('channels.list'),
      ]);
      setRules(r ?? []);
      setChannels(c ?? []);
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : '加载数据失败');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { loadData(); }, [loadData]);

  const resetForm = () => {
    setSourceId('');
    setTargetId('');
    setMatchPattern('');
    setEditingRule(null);
    setShowForm(false);
    setError('');
  };

  const openEdit = (rule: ForwardRule) => {
    setEditingRule(rule);
    setSourceId(String(rule.source_channel_id));
    setTargetId(String(rule.target_channel_id));
    setMatchPattern(rule.match_pattern);
    setShowForm(true);
  };

  const handleSubmit = async () => {
    setError('');
    const source = channels.find((c) => String(c.id) === sourceId);
    const target = channels.find((c) => String(c.id) === targetId);

    if (!source || !target || !matchPattern) {
      setError('请填写所有字段');
      return;
    }

    try {
      if (editingRule) {
        await rpc('rules.update', {
          id: editingRule.id,
          source_channel_id: source.id,
          source_name: source.name,
          source_hash: source.access_hash,
          target_channel_id: target.id,
          target_name: target.name,
          target_hash: target.access_hash,
          match_pattern: matchPattern,
        });
      } else {
        await rpc('rules.create', {
          source_channel_id: source.id,
          source_name: source.name,
          source_hash: source.access_hash,
          target_channel_id: target.id,
          target_name: target.name,
          target_hash: target.access_hash,
          match_pattern: matchPattern,
        });
      }
      resetForm();
      await loadData();
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : '保存规则失败');
    }
  };

  const handleDelete = async (id: number) => {
    if (!confirm('确定删除此规则？')) return;
    try {
      await rpc('rules.delete', { id });
      await loadData();
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : '删除规则失败');
    }
  };

  const handleToggle = async (rule: ForwardRule) => {
    try {
      await rpc('rules.update', { id: rule.id, enabled: !rule.enabled });
      await loadData();
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : '切换规则状态失败');
    }
  };

  if (loading) {
    return <div className="flex items-center justify-center h-64 text-gray-500">加载中...</div>;
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-xl font-bold text-gray-800">转发规则</h2>
        <button
          onClick={() => { resetForm(); setShowForm(true); }}
          className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 text-sm"
        >
          新建规则
        </button>
      </div>

      {error && (
        <div className="mb-4 p-3 bg-red-50 text-red-700 rounded text-sm">{error}</div>
      )}

      {showForm && (
        <div className="bg-white rounded-lg shadow p-4 mb-6">
          <h3 className="font-medium text-gray-700 mb-4">
            {editingRule ? '编辑规则' : '创建规则'}
          </h3>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">来源</label>
              <select
                value={sourceId}
                onChange={(e) => setSourceId(e.target.value)}
                className="w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                <option value="">选择来源...</option>
                {channels.map((c) => (
                  <option key={c.id} value={c.id}>{c.name} [{typeLabel[c.type] ?? c.type}]</option>
                ))}
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">目标</label>
              <select
                value={targetId}
                onChange={(e) => setTargetId(e.target.value)}
                className="w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                <option value="">选择目标...</option>
                {channels.map((c) => (
                  <option key={c.id} value={c.id}>{c.name} [{typeLabel[c.type] ?? c.type}]</option>
                ))}
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">匹配规则 (正则)</label>
              <input
                type="text"
                value={matchPattern}
                onChange={(e) => setMatchPattern(e.target.value)}
                placeholder=".*keyword.*"
                className="w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
          </div>
          <div className="flex gap-2 mt-4">
            <button
              onClick={handleSubmit}
              className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 text-sm"
            >
              {editingRule ? '更新' : '创建'}
            </button>
            <button
              onClick={resetForm}
              className="px-4 py-2 bg-gray-200 text-gray-700 rounded-md hover:bg-gray-300 text-sm"
            >
              取消
            </button>
          </div>
        </div>
      )}

      <div className="bg-white rounded-lg shadow">
        <table className="w-full">
          <thead>
            <tr className="border-b text-left text-sm text-gray-500">
              <th className="px-4 py-3">来源</th>
              <th className="px-4 py-3">目标</th>
              <th className="px-4 py-3">匹配规则</th>
              <th className="px-4 py-3">状态</th>
              <th className="px-4 py-3">操作</th>
            </tr>
          </thead>
          <tbody className="divide-y">
            {rules.map((rule) => (
              <tr key={rule.id}>
                <td className="px-4 py-3 text-sm">{rule.source_name || rule.source_channel_id}</td>
                <td className="px-4 py-3 text-sm">{rule.target_name || rule.target_channel_id}</td>
                <td className="px-4 py-3 text-sm font-mono text-xs">{rule.match_pattern}</td>
                <td className="px-4 py-3">
                  <button
                    onClick={() => handleToggle(rule)}
                    className={`text-xs px-2 py-1 rounded-full ${
                      rule.enabled
                        ? 'bg-green-100 text-green-800'
                        : 'bg-gray-100 text-gray-500'
                    }`}
                  >
                    {rule.enabled ? '已启用' : '已禁用'}
                  </button>
                </td>
                <td className="px-4 py-3">
                  <div className="flex gap-2">
                    <button
                      onClick={() => openEdit(rule)}
                      className="text-xs text-blue-600 hover:underline"
                    >
                      编辑
                    </button>
                    <button
                      onClick={() => handleDelete(rule.id)}
                      className="text-xs text-red-600 hover:underline"
                    >
                      删除
                    </button>
                  </div>
                </td>
              </tr>
            ))}
            {rules.length === 0 && (
              <tr>
                <td colSpan={5} className="px-4 py-8 text-center text-gray-400">
                  暂无转发规则
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
