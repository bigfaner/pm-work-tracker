import { useState } from 'react'
import { useNavigate, useSearchParams, Navigate } from 'react-router-dom'
import { loginApi } from '@/api/auth'
import { useAuthStore } from '@/store/auth'

export default function LoginPage() {
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated)
  const setAuth = useAuthStore((s) => s.setAuth)
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')

  if (isAuthenticated) {
    return <Navigate replace to="/items" />
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!username || !password) return
    setLoading(true)
    setError(null)
    try {
      const resp = await loginApi({ username, password })
      setAuth(resp.token, resp.user)
      const redirect = searchParams.get('redirect') || '/items'
      navigate(redirect)
    } catch (err: any) {
      const status = err?.response?.status
      if (status === 401) {
        setError('账号或密码错误')
        setPassword('')
      } else if (status === 400) {
        setError('请求参数校验失败')
      } else {
        setError('登录失败，请稍后重试')
      }
    } finally {
      setLoading(false)
    }
  }

  return (
    <div data-testid="login-page" style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '100vh', background: '#f0f2f5' }}>
      <div style={{ width: 400, padding: '40px 32px', background: '#fff', borderRadius: 8, boxShadow: '0 4px 24px rgba(0,0,0,0.08)' }}>
        <div style={{ textAlign: 'center', marginBottom: 32 }}>
          <h2 style={{ fontSize: 20, fontWeight: 600, margin: 0 }}>PM 工作事项追踪</h2>
        </div>
        <form onSubmit={handleSubmit}>
          <div style={{ marginBottom: 16 }}>
            <label htmlFor="login-username" style={{ display: 'block', marginBottom: 4 }}>账号</label>
            <input
              id="login-username"
              data-testid="login-username"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              placeholder="请输入账号"
              style={{ width: '100%', padding: '8px 12px', border: '1px solid #d9d9d9', borderRadius: 6, fontSize: 16 }}
            />
          </div>
          <div style={{ marginBottom: 16 }}>
            <label htmlFor="login-password" style={{ display: 'block', marginBottom: 4 }}>密码</label>
            <input
              id="login-password"
              data-testid="login-password"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="请输入密码"
              style={{ width: '100%', padding: '8px 12px', border: '1px solid #d9d9d9', borderRadius: 6, fontSize: 16 }}
            />
          </div>
          {error && (
            <div data-testid="login-error" style={{ color: '#ff4d4f', marginBottom: 16 }}>{error}</div>
          )}
          <button
            type="submit"
            disabled={loading}
            data-testid="login-submit"
            style={{ width: '100%', padding: '8px 16px', background: '#1677ff', color: '#fff', border: 'none', borderRadius: 6, fontSize: 16, cursor: loading ? 'not-allowed' : 'pointer' }}
          >
            {loading ? '登录中...' : '登录'}
          </button>
        </form>
      </div>
    </div>
  )
}
