import { useState, useEffect, useCallback } from 'react';
import { rpc } from '../lib/rpc';

type ForwardRule = {
  id: number;
  source_channel_id: number;
  source_name: string;
  source_hash: number;
  target_channel_id: number;
  target_name: string;
  target_hash: number;
  match_pattern: string;
  enabled: boolean;
  created_at: string;
  updated_at: string;
};

type ChannelInfo = {
  id: number;
  name: string;
  type: string;
  access_hash: number;
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
      setError(e instanceof Error ? e.message : 'Failed to load data');
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
      setError('All fields are required');
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
      setError(e instanceof Error ? e.message : 'Failed to save rule');
    }
  };

  const handleDelete = async (id: number) => {
    if (!confirm('Delete this rule?')) return;
    try {
      await rpc('rules.delete', { id });
      await loadData();
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : 'Failed to delete rule');
    }
  };

  const handleToggle = async (rule: ForwardRule) => {
    try {
      await rpc('rules.update', { id: rule.id, enabled: !rule.enabled });
      await loadData();
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : 'Failed to toggle rule');
    }
  };

  if (loading) {
    return <div className="flex items-center justify-center h-64 text-gray-500">Loading...</div>;
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-xl font-bold text-gray-800">Forward Rules</h2>
        <button
          onClick={() => { resetForm(); setShowForm(true); }}
          className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 text-sm"
        >
          New Rule
        </button>
      </div>

      {error && (
        <div className="mb-4 p-3 bg-red-50 text-red-700 rounded text-sm">{error}</div>
      )}

      {showForm && (
        <div className="bg-white rounded-lg shadow p-4 mb-6">
          <h3 className="font-medium text-gray-700 mb-4">
            {editingRule ? 'Edit Rule' : 'Create Rule'}
          </h3>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Source Channel</label>
              <select
                value={sourceId}
                onChange={(e) => setSourceId(e.target.value)}
                className="w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                <option value="">Select source...</option>
                {channels.map((c) => (
                  <option key={c.id} value={c.id}>{c.name}</option>
                ))}
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Target Channel</label>
              <select
                value={targetId}
                onChange={(e) => setTargetId(e.target.value)}
                className="w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                <option value="">Select target...</option>
                {channels.map((c) => (
                  <option key={c.id} value={c.id}>{c.name}</option>
                ))}
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Match Pattern (regex)</label>
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
              {editingRule ? 'Update' : 'Create'}
            </button>
            <button
              onClick={resetForm}
              className="px-4 py-2 bg-gray-200 text-gray-700 rounded-md hover:bg-gray-300 text-sm"
            >
              Cancel
            </button>
          </div>
        </div>
      )}

      <div className="bg-white rounded-lg shadow">
        <table className="w-full">
          <thead>
            <tr className="border-b text-left text-sm text-gray-500">
              <th className="px-4 py-3">Source</th>
              <th className="px-4 py-3">Target</th>
              <th className="px-4 py-3">Pattern</th>
              <th className="px-4 py-3">Status</th>
              <th className="px-4 py-3">Actions</th>
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
                    {rule.enabled ? 'Enabled' : 'Disabled'}
                  </button>
                </td>
                <td className="px-4 py-3">
                  <div className="flex gap-2">
                    <button
                      onClick={() => openEdit(rule)}
                      className="text-xs text-blue-600 hover:underline"
                    >
                      Edit
                    </button>
                    <button
                      onClick={() => handleDelete(rule.id)}
                      className="text-xs text-red-600 hover:underline"
                    >
                      Delete
                    </button>
                  </div>
                </td>
              </tr>
            ))}
            {rules.length === 0 && (
              <tr>
                <td colSpan={5} className="px-4 py-8 text-center text-gray-400">
                  No forwarding rules yet
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
