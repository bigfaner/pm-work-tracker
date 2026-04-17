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
  display_name: 'Test User',
  is_super_admin: false,
  can_create_team: false,
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
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

// antd renders Chinese button text with spaces between chars (e.g. "登 录")
function getSubmitButton() {
  return screen.getByRole('button', { name: /登.*录/ })
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

  it('renders username and password fields', () => {
    renderLoginPage()
    expect(screen.getByLabelText('账号')).toBeInTheDocument()
    expect(screen.getByLabelText('密码')).toBeInTheDocument()
  })

  it('renders submit button with text 登录', () => {
    renderLoginPage()
    expect(getSubmitButton()).toBeInTheDocument()
  })

  it('shows validation errors when submitting empty form', async () => {
    const user = userEvent.setup()
    renderLoginPage()
    await user.click(getSubmitButton())

    await waitFor(() => {
      expect(screen.getByText('请输入账号')).toBeInTheDocument()
      expect(screen.getByText('请输入密码')).toBeInTheDocument()
    })
  })

  it('shows loading state on submit button while request is in flight', async () => {
    let resolveLogin!: (value: unknown) => void
    mockLoginApi.mockReturnValue(
      new Promise((resolve) => {
        resolveLogin = resolve
      }),
    )

    const user = userEvent.setup()
    renderLoginPage()

    await user.type(screen.getByLabelText('账号'), 'testuser')
    await user.type(screen.getByLabelText('密码'), 'password123')
    await user.click(getSubmitButton())

    // Button should be in loading state while request is pending
    expect(getSubmitButton()).toHaveClass('ant-btn-loading')

    // Resolve to clean up (component navigates away after success)
    resolveLogin({ token: 'jwt-token', user: mockUser })
    await waitFor(() => {
      expect(screen.getByTestId('items-page')).toBeInTheDocument()
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

  it('shows inline field errors on 400 validation response', async () => {
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
})
