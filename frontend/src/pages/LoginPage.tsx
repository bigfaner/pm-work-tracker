import { useState, useEffect } from 'react'
import { useNavigate, useSearchParams, Navigate } from 'react-router-dom'
import { loginApi } from '@/api/auth'
import { useAuthStore } from '@/store/auth'
import { Card, CardContent } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'

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

  const canSubmit = username.trim() !== '' && password.trim() !== '' && !loading

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!canSubmit) return
    setLoading(true)
    setError(null)
    try {
      const resp = await loginApi({ username, password })
      setAuth(resp.token, resp.user)
      const redirect = searchParams.get('redirect') || '/items'
      navigate(redirect)
    } catch (err: any) {
      const status = err?.response?.status
      const code = err?.response?.data?.code
      if (status === 401) {
        setError('账号或密码错误')
        setPassword('')
      } else if (status === 403 && code === 'USER_DISABLED') {
        setError('账号已被禁用，请联系管理员')
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
    <div
      data-testid="login-page"
      className="min-h-screen bg-bg-alt flex items-center justify-center p-5"
    >
      <Card className="w-full max-w-[400px]">
        <CardContent className="p-10 px-8">
          <div className="text-center mb-8">
            <div className="w-12 h-12 mx-auto mb-3 bg-primary-100 rounded-xl flex items-center justify-center">
              <svg
                width="28"
                height="28"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                className="text-primary-600"
                strokeWidth="2"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"
                />
              </svg>
            </div>
            <h2 className="text-xl font-semibold mb-1">PM Tracker</h2>
            <p className="text-sm text-tertiary">工作事项追踪系统</p>
          </div>

          {error && (
            <div
              data-testid="login-error"
              className="bg-error-bg border border-red-200 rounded-lg px-3.5 py-2.5 text-sm text-error-text mb-4"
            >
              {error}
            </div>
          )}

          <form onSubmit={handleSubmit}>
            <div className="mb-4">
              <label
                htmlFor="login-username"
                className="block text-sm font-medium text-secondary mb-1.5"
              >
                账号
              </label>
              <Input
                id="login-username"
                data-testid="login-username"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                placeholder="请输入账号"
              />
            </div>
            <div className="mb-5">
              <label
                htmlFor="login-password"
                className="block text-sm font-medium text-secondary mb-1.5"
              >
                密码
              </label>
              <Input
                id="login-password"
                data-testid="login-password"
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="请输入密码"
              />
            </div>
            <Button
              type="submit"
              disabled={!canSubmit}
              data-testid="login-submit"
              className="w-full h-11 text-base"
            >
              {loading ? '登录中...' : '登录'}
            </Button>
          </form>

          <p className="text-center mt-6 text-xs text-tertiary">v1.0.0</p>
        </CardContent>
      </Card>
    </div>
  )
}
