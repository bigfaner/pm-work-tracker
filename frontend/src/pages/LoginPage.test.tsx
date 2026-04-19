import { describe, it, expect, beforeEach, vi } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import LoginPage from './LoginPage'
import { useAuthStore } from '@/store/auth'
import type { User } from '@/types'

// Mock the login API
const mockLoginApi = vi.fn()
vi.mock('@/api/auth', () => ({
  loginApi: (...args: unknown[]) => mockLoginApi(...args),
}))

const mockUser: User = {
  id: 1,
  username: 'testuser',
  displayName: 'Test User',
  isSuperAdmin: false,
  canCreateTeam: false,
}

function renderLoginPage(initialPath = '/login', search = '') {
  return render(
    <MemoryRouter initialEntries={[{ pathname: initialPath, search }]}>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route path="/items" element={<div data-testid="items-page">Items</div>} />
        <Route path="/some-redirect" element={<div data-testid="redirect-page">Redirected</div>} />
      </Routes>
    </MemoryRouter>,
  )
}

function getSubmitButton() {
  return screen.getByRole('button', { name: /登录/ })
}

describe('LoginPage', () => {
  beforeEach(() => {
    useAuthStore.getState().clearAuth()
    mockLoginApi.mockReset()
  })

  it('renders a centered login card with data-testid', () => {
    renderLoginPage()
    expect(screen.getByTestId('login-page')).toBeInTheDocument()
  })

  it('renders logo/title and subtitle', () => {
    renderLoginPage()
    expect(screen.getByText('PM Tracker')).toBeInTheDocument()
    expect(screen.getByText('工作事项追踪系统')).toBeInTheDocument()
  })

  it('renders username and password fields', () => {
    renderLoginPage()
    expect(screen.getByLabelText('账号')).toBeInTheDocument()
    expect(screen.getByLabelText('密码')).toBeInTheDocument()
  })

  it('renders submit button with text 登录', () => {
    renderLoginPage()
    expect(getSubmitButton()).toBeInTheDocument()
  })

  it('disables submit button when both fields are empty', () => {
    renderLoginPage()
    expect(getSubmitButton()).toBeDisabled()
  })

  it('disables submit button when only username is filled', async () => {
    const user = userEvent.setup()
    renderLoginPage()
    await user.type(screen.getByLabelText('账号'), 'testuser')
    expect(getSubmitButton()).toBeDisabled()
  })

  it('disables submit button when only password is filled', async () => {
    const user = userEvent.setup()
    renderLoginPage()
    await user.type(screen.getByLabelText('密码'), 'password')
    expect(getSubmitButton()).toBeDisabled()
  })

  it('enables submit button when both fields are non-empty', async () => {
    const user = userEvent.setup()
    renderLoginPage()
    await user.type(screen.getByLabelText('账号'), 'testuser')
    await user.type(screen.getByLabelText('密码'), 'password')
    expect(getSubmitButton()).not.toBeDisabled()
  })

  it('does not call loginApi when submitting empty form', async () => {
    const user = userEvent.setup()
    renderLoginPage()
    // Button is disabled, but even if somehow clicked
    expect(getSubmitButton()).toBeDisabled()
    expect(mockLoginApi).not.toHaveBeenCalled()
  })

  it('calls loginApi exactly once with valid credentials', async () => {
    mockLoginApi.mockResolvedValue({ token: 'jwt-token-123', user: mockUser })

    const user = userEvent.setup()
    renderLoginPage()

    await user.type(screen.getByLabelText('账号'), 'testuser')
    await user.type(screen.getByLabelText('密码'), 'password123')
    await user.click(getSubmitButton())

    await waitFor(() => {
      expect(mockLoginApi).toHaveBeenCalledTimes(1)
    })
    expect(mockLoginApi).toHaveBeenCalledWith({
      username: 'testuser',
      password: 'password123',
    })
  })

  it('stores token and user via setAuth on successful login', async () => {
    mockLoginApi.mockResolvedValue({ token: 'jwt-token-123', user: mockUser })

    const user = userEvent.setup()
    renderLoginPage()

    await user.type(screen.getByLabelText('账号'), 'testuser')
    await user.type(screen.getByLabelText('密码'), 'password123')
    await user.click(getSubmitButton())

    await waitFor(() => {
      const state = useAuthStore.getState()
      expect(state.token).toBe('jwt-token-123')
      expect(state.user).toEqual(mockUser)
    })
  })

  it('navigates to /items on successful login', async () => {
    mockLoginApi.mockResolvedValue({ token: 'jwt-token-123', user: mockUser })

    const user = userEvent.setup()
    renderLoginPage()

    await user.type(screen.getByLabelText('账号'), 'testuser')
    await user.type(screen.getByLabelText('密码'), 'password123')
    await user.click(getSubmitButton())

    await waitFor(() => {
      expect(screen.getByTestId('items-page')).toBeInTheDocument()
    })
  })

  it('shows inline error 账号或密码错误 on 401 response', async () => {
    const error = new Error('Unauthorized') as any
    error.response = {
      status: 401,
      data: { code: 'UNAUTHORIZED', message: '账号或密码错误' },
    }
    mockLoginApi.mockRejectedValue(error)

    const user = userEvent.setup()
    renderLoginPage()

    await user.type(screen.getByLabelText('账号'), 'wrong')
    await user.type(screen.getByLabelText('密码'), 'wrong')
    await user.click(getSubmitButton())

    await waitFor(() => {
      expect(screen.getByText('账号或密码错误')).toBeInTheDocument()
    })
  })

  it('shows specific error for USER_DISABLED on 403 response', async () => {
    const error = new Error('Forbidden') as any
    error.response = {
      status: 403,
      data: { code: 'USER_DISABLED', message: '账号已被禁用' },
    }
    mockLoginApi.mockRejectedValue(error)

    const user = userEvent.setup()
    renderLoginPage()

    await user.type(screen.getByLabelText('账号'), 'disabled')
    await user.type(screen.getByLabelText('密码'), 'password')
    await user.click(getSubmitButton())

    await waitFor(() => {
      expect(screen.getByText('账号已被禁用，请联系管理员')).toBeInTheDocument()
    })
  })

  it('clears password field on login failure', async () => {
    const error = new Error('Unauthorized') as any
    error.response = {
      status: 401,
      data: { code: 'UNAUTHORIZED', message: '账号或密码错误' },
    }
    mockLoginApi.mockRejectedValue(error)

    const user = userEvent.setup()
    renderLoginPage()

    const passwordInput = screen.getByLabelText('密码')
    await user.type(screen.getByLabelText('账号'), 'testuser')
    await user.type(passwordInput, 'wrongpassword')
    await user.click(getSubmitButton())

    await waitFor(() => {
      expect(passwordInput).toHaveValue('')
    })
  })

  it('preserves username field on login failure', async () => {
    const error = new Error('Unauthorized') as any
    error.response = {
      status: 401,
      data: { code: 'UNAUTHORIZED', message: '账号或密码错误' },
    }
    mockLoginApi.mockRejectedValue(error)

    const user = userEvent.setup()
    renderLoginPage()

    const usernameInput = screen.getByLabelText('账号')
    await user.type(usernameInput, 'testuser')
    await user.type(screen.getByLabelText('密码'), 'wrongpassword')
    await user.click(getSubmitButton())

    await waitFor(() => {
      expect(screen.getByText('账号或密码错误')).toBeInTheDocument()
    })
    expect(usernameInput).toHaveValue('testuser')
  })

  it('shows inline error on 400 validation response', async () => {
    const error = new Error('Bad Request') as any
    error.response = {
      status: 400,
      data: {
        code: 'VALIDATION_ERROR',
        message: '字段缺失',
      },
    }
    mockLoginApi.mockRejectedValue(error)

    const user = userEvent.setup()
    renderLoginPage()

    await user.type(screen.getByLabelText('账号'), 'testuser')
    await user.type(screen.getByLabelText('密码'), 'password')
    await user.click(getSubmitButton())

    await waitFor(() => {
      expect(screen.getByText('请求参数校验失败')).toBeInTheDocument()
    })
  })

  it('redirects to /items if already authenticated', () => {
    useAuthStore.getState().setAuth('token', mockUser)
    renderLoginPage()

    expect(screen.getByTestId('items-page')).toBeInTheDocument()
    expect(screen.queryByTestId('login-page')).not.toBeInTheDocument()
  })

  it('navigates to redirect query param on successful login', async () => {
    mockLoginApi.mockResolvedValue({ token: 'jwt-token-123', user: mockUser })

    const user = userEvent.setup()
    renderLoginPage('/login', '?redirect=/some-redirect')

    await user.type(screen.getByLabelText('账号'), 'testuser')
    await user.type(screen.getByLabelText('密码'), 'password123')
    await user.click(getSubmitButton())

    await waitFor(() => {
      expect(screen.getByTestId('redirect-page')).toBeInTheDocument()
    })
  })

  it('shows generic error on unexpected server error', async () => {
    const error = new Error('Server Error') as any
    error.response = {
      status: 500,
      data: { code: 'INTERNAL_ERROR', message: '服务器内部错误' },
    }
    mockLoginApi.mockRejectedValue(error)

    const user = userEvent.setup()
    renderLoginPage()

    await user.type(screen.getByLabelText('账号'), 'testuser')
    await user.type(screen.getByLabelText('密码'), 'password123')
    await user.click(getSubmitButton())

    await waitFor(() => {
      expect(screen.getByText('登录失败，请稍后重试')).toBeInTheDocument()
    })
  })

  it('form submits on Enter key press in password field', async () => {
    mockLoginApi.mockResolvedValue({ token: 'jwt-token-123', user: mockUser })

    const user = userEvent.setup()
    renderLoginPage()

    await user.type(screen.getByLabelText('账号'), 'testuser')
    await user.type(screen.getByLabelText('密码'), 'password123{Enter}')

    await waitFor(() => {
      expect(mockLoginApi).toHaveBeenCalledWith({
        username: 'testuser',
        password: 'password123',
      })
    })
  })

  it('shows loading state on button during login', async () => {
    // Create a promise we can control
    let resolveLogin: (value: any) => void
    mockLoginApi.mockReturnValue(new Promise((resolve) => { resolveLogin = resolve }))

    const user = userEvent.setup()
    renderLoginPage()

    await user.type(screen.getByLabelText('账号'), 'testuser')
    await user.type(screen.getByLabelText('密码'), 'password123')
    await user.click(getSubmitButton())

    // Button should show loading text and be disabled
    expect(getSubmitButton()).toHaveTextContent('登录中...')
    expect(getSubmitButton()).toBeDisabled()

    // Resolve to clean up
    resolveLogin!({ token: 'jwt-token-123', user: mockUser })
    await waitFor(() => {
      expect(screen.queryByText('登录中...')).not.toBeInTheDocument()
    })
  })

  it('uses no antd imports', () => {
    // Verify the component module can be loaded without antd
    const fs = require('fs')
    const path = require('path')
    const source = fs.readFileSync(
      path.resolve(__dirname, 'LoginPage.tsx'),
      'utf-8',
    )
    expect(source).not.toContain('antd')
    expect(source).not.toContain('ant-design')
  })
})
