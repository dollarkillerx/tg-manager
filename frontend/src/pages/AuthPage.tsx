import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { rpc } from '../lib/rpc';

type AuthStatus = { authorized: boolean; user?: { id: number; first_name: string; last_name: string } };
type SendCodeResult = { code_type: string };
type VerifyCodeResult = { authorized: boolean; password_needed: boolean };

type Step = 'phone' | 'code' | 'password';

export default function AuthPage() {
  const navigate = useNavigate();
  const [step, setStep] = useState<Step>('phone');
  const [phone, setPhone] = useState('');
  const [code, setCode] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const [codeType, setCodeType] = useState('');

  useEffect(() => {
    rpc<AuthStatus>('auth.status').then((res) => {
      if (res.authorized) navigate('/', { replace: true });
    }).catch(() => {});
  }, [navigate]);

  const handleSendCode = async () => {
    setError('');
    setLoading(true);
    try {
      const res = await rpc<SendCodeResult>('auth.sendCode', { phone });
      setCodeType(res.code_type);
      setStep('code');
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : 'Failed to send code');
    } finally {
      setLoading(false);
    }
  };

  const handleVerifyCode = async () => {
    setError('');
    setLoading(true);
    try {
      const res = await rpc<VerifyCodeResult>('auth.verifyCode', { code });
      if (res.authorized) {
        navigate('/', { replace: true });
      } else if (res.password_needed) {
        setStep('password');
      }
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : 'Failed to verify code');
    } finally {
      setLoading(false);
    }
  };

  const handleSendPassword = async () => {
    setError('');
    setLoading(true);
    try {
      await rpc('auth.sendPassword', { password });
      navigate('/', { replace: true });
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : 'Failed to authenticate');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-100">
      <div className="bg-white p-8 rounded-lg shadow-md w-full max-w-md">
        <h1 className="text-2xl font-bold text-center mb-6">Telegram Login</h1>

        {error && (
          <div className="mb-4 p-3 bg-red-50 text-red-700 rounded text-sm">{error}</div>
        )}

        {step === 'phone' && (
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Phone Number</label>
              <input
                type="tel"
                value={phone}
                onChange={(e) => setPhone(e.target.value)}
                placeholder="+1234567890"
                className="w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                onKeyDown={(e) => e.key === 'Enter' && handleSendCode()}
              />
            </div>
            <button
              onClick={handleSendCode}
              disabled={loading || !phone}
              className="w-full py-2 px-4 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50"
            >
              {loading ? 'Sending...' : 'Send Code'}
            </button>
          </div>
        )}

        {step === 'code' && (
          <div className="space-y-4">
            <p className="text-sm text-gray-600">
              Code sent via <strong>{codeType}</strong> to {phone}
            </p>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Verification Code</label>
              <input
                type="text"
                value={code}
                onChange={(e) => setCode(e.target.value)}
                placeholder="12345"
                className="w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                onKeyDown={(e) => e.key === 'Enter' && handleVerifyCode()}
              />
            </div>
            <button
              onClick={handleVerifyCode}
              disabled={loading || !code}
              className="w-full py-2 px-4 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50"
            >
              {loading ? 'Verifying...' : 'Verify Code'}
            </button>
          </div>
        )}

        {step === 'password' && (
          <div className="space-y-4">
            <p className="text-sm text-gray-600">Two-factor authentication is enabled. Enter your password.</p>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">2FA Password</label>
              <input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="Password"
                className="w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                onKeyDown={(e) => e.key === 'Enter' && handleSendPassword()}
              />
            </div>
            <button
              onClick={handleSendPassword}
              disabled={loading || !password}
              className="w-full py-2 px-4 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50"
            >
              {loading ? 'Authenticating...' : 'Submit Password'}
            </button>
          </div>
        )}
      </div>
    </div>
  );
}
