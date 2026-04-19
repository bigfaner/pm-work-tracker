import { describe, it, expect, beforeEach, beforeAll, afterAll, afterEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { server } from '@/mocks/server'
import LoginPage from '@/pages/LoginPage'
import { useAuthStore } from '@/store/auth'
import type { User } from '@/types'

// MSW lifecycle for this test suite only
beforeAll(() => server.listen({ onUnhandledRequest: 'bypass' }))
afterEach(() => server.resetHandlers())
afterAll(() => server.close())

// --- Shared data ---

const mockUser: User = {
  id: 1,
  username: 'testuser',
  displayName: 'Test User',
  isSuperAdmin: false,
  canCreateTeam: false,
}

function getLoginButton() {
  return screen.getByRole('button', { name: /登录/ })
}

// --- Tests ---

describe('Integration: Login flow', () => {
  beforeEach(() => {
    useAuthStore.getState().clearAuth()
  })

  it('login success: fill credentials, submit, redirect to /items', async () => {
    const user = userEvent.setup()
    render(
      <MemoryRouter initialEntries={['/login']}>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route path="/items" element={<div data-testid="items-page">Items Page</div>} />
        </Routes>
      </MemoryRouter>,
    )

    await user.type(screen.getByLabelText('账号'), 'testuser')
    await user.type(screen.getByLabelText('密码'), 'password123')
    await user.click(getLoginButton())

    await waitFor(() => {
      expect(screen.getByTestId('items-page')).toBeInTheDocument()
    })

    // Auth store should have token and user
    const state = useAuthStore.getState()
    expect(state.token).toBe('jwt-token-123')
    expect(state.user?.username).toBe('testuser')
    expect(state.isAuthenticated).toBe(true)
  })

  it('login failure: MSW returns 401, inline error displayed, no redirect', async () => {
    const user = userEvent.setup()
    render(
      <MemoryRouter initialEntries={['/login']}>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route path="/items" element={<div data-testid="items-page">Items Page</div>} />
        </Routes>
      </MemoryRouter>,
    )

    await user.type(screen.getByLabelText('账号'), 'wronguser')
    await user.type(screen.getByLabelText('密码'), 'wrongpass')
    await user.click(getLoginButton())

    await waitFor(() => {
      expect(screen.getByText('账号或密码错误')).toBeInTheDocument()
    })

    // Should stay on login page
    expect(screen.getByTestId('login-page')).toBeInTheDocument()
    expect(screen.queryByTestId('items-page')).not.toBeInTheDocument()

    // Auth store should not have token
    expect(useAuthStore.getState().token).toBeNull()
  })
})
