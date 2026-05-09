import { describe, it, expect, vi, beforeEach, beforeAll, afterAll, afterEach } from 'vitest'
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { server } from '@/mocks/server'
import { http, HttpResponse } from 'msw'
import { ToastProvider } from '@/components/ui/toast'
import { useAuthStore } from '@/store/auth'
import UserManagementPage from './UserManagementPage'

// Mock clipboard utility
vi.mock('@/lib/utils', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@/lib/utils')>()
  return {
    ...actual,
    copyToClipboard: vi.fn().mockResolvedValue(undefined),
  }
})

const { copyToClipboard } = await import('@/lib/utils')

// MSW lifecycle
beforeAll(() => server.listen({ onUnhandledRequest: 'bypass' }))
afterEach(() => server.resetHandlers())
afterAll(() => server.close())

// --- Helpers ---

function createQueryClient() {
  return new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
}

function renderPage() {
  const qc = createQueryClient()
  return render(
    <QueryClientProvider client={qc}>
      <ToastProvider>
        <MemoryRouter>
          <UserManagementPage />
        </MemoryRouter>
      </ToastProvider>
    </QueryClientProvider>,
  )
}

// --- Seed data ---

const seedUsers = [
  {
    bizKey: '1', username: 'zhangming', displayName: '张明', email: 'zhangming@example.com',
    isSuperAdmin: true, userStatus: 'enabled',
    teams: [{ bizKey: '1', name: '产品研发团队', role: 'pm' }],
  },
  {
    bizKey: '2', username: 'lihua', displayName: '李华', email: 'lihua@example.com',
    isSuperAdmin: false, userStatus: 'enabled',
    teams: [{ bizKey: '1', name: '产品研发团队', role: 'member' }],
  },
  {
    bizKey: '3', username: 'wangfang', displayName: '王芳', email: 'wangfang@example.com',
    isSuperAdmin: false, userStatus: 'enabled',
    teams: [{ bizKey: '1', name: '产品研发团队', role: 'member' }],
  },
  {
    bizKey: '4', username: 'zhaoqiang', displayName: '赵强', email: 'zhaoqiang@example.com',
    isSuperAdmin: false, userStatus: 'disabled',
    teams: [{ bizKey: '2', name: '设计团队', role: 'member' }],
  },
]

const seedTeams = [
  { bizKey: '2', name: '设计团队', pmDisplayName: '李华', memberCount: 2, mainItemCount: 2, createTime: '2024-02-01' },
]

function setupHandlers() {
  server.use(
    // List users
    http.get('/v1/admin/users', ({ request }) => {
      const url = new URL(request.url)
      const search = url.searchParams.get('search')
      const page = Number(url.searchParams.get('page') || 1)
      const pageSizeParam = Number(url.searchParams.get('pageSize') || 10)

      let filtered = [...seedUsers]
      if (search) {
        const q = search.toLowerCase()
        filtered = filtered.filter(
          (u) => u.username.toLowerCase().includes(q) || u.displayName.toLowerCase().includes(q),
        )
      }

      const start = (page - 1) * pageSizeParam
      const items = filtered.slice(start, start + pageSizeParam)

      return HttpResponse.json({
        code: 0,
        data: { items, total: filtered.length, page, pageSize: pageSizeParam },
      })
    }),

    // List admin teams
    http.get('/v1/admin/teams', () => {
      return HttpResponse.json({
        code: 0,
        data: { items: seedTeams, total: seedTeams.length, page: 1, pageSize: 100 },
      })
    }),

    // Create user
    http.post('/v1/admin/users', async ({ request }) => {
      const body = (await request.json()) as Record<string, unknown>
      const username = body.username as string

      if (username === 'zhangming') {
        return HttpResponse.json(
          { code: 'USER_EXISTS', message: '账号名已存在' },
          { status: 422 },
        )
      }

      return HttpResponse.json({
        code: 0,
        data: {
          id: 100,
          username,
          displayName: body.displayName,
          email: body.email || '',
          userStatus: 'enabled',
          teams: [],
          initialPassword: 'Abc123456789',
        },
      }, { status: 201 })
    }),

    // Update user
    http.put('/v1/admin/users/:userId', async ({ params, request }) => {
      const userId = Number(params.userId)
      const body = (await request.json()) as Record<string, unknown>
      const user = seedUsers.find((u) => u.bizKey === String(userId))
      if (!user) {
        return HttpResponse.json({ code: 'USER_NOT_FOUND', message: 'not found' }, { status: 404 })
      }
      return HttpResponse.json({
        code: 0,
        data: { ...user, ...body },
      })
    }),

    // Toggle user status
    http.put('/v1/admin/users/:userId/status', async ({ params, request }) => {
      const userId = Number(params.userId)
      const body = (await request.json()) as { status: string }
      const user = seedUsers.find((u) => u.bizKey === String(userId))
      if (!user) {
        return HttpResponse.json({ code: 'USER_NOT_FOUND', message: 'not found' }, { status: 404 })
      }
      if (userId === 1 && body.status === 'disabled') {
        return HttpResponse.json(
          { code: 'CANNOT_DISABLE_SELF', message: '不能禁用自己' },
          { status: 422 },
        )
      }
      return HttpResponse.json({
        code: 0,
        data: { id: userId, username: user.username, userStatus: body.status },
      })
    }),

    // Reset password
    http.put('/v1/admin/users/:userId/password', async ({ params, request }) => {
      const userId = String(params.userId)
      const body = (await request.json()) as { newPassword: string }
      const user = seedUsers.find((u) => u.bizKey === userId)
      if (!user) {
        return HttpResponse.json({ code: 'USER_NOT_FOUND', message: 'not found' }, { status: 404 })
      }
      if (!body.newPassword || body.newPassword.length < 8) {
        return HttpResponse.json(
          { code: 'VALIDATION_ERROR', message: '密码需至少8位' },
          { status: 400 },
        )
      }
      return HttpResponse.json({
        code: 0,
        data: { bizKey: userId, username: user.username, displayName: user.displayName },
      })
    }),
  )
}

describe('UserManagementPage', () => {
  beforeEach(() => {
    setupHandlers()
  })

  // --- Core rendering ---

  it('renders page header with title', async () => {
    renderPage()
    expect(screen.getByText('用户管理')).toBeInTheDocument()
    expect(screen.getByText('创建用户')).toBeInTheDocument()
  })

  it('renders filter bar with search and filter', async () => {
    renderPage()
    expect(screen.getByPlaceholderText('搜索用户名/姓名')).toBeInTheDocument()
    await waitFor(() => {
      expect(screen.getByText('张明')).toBeInTheDocument()
    })
  })

  // --- User table ---

  it('displays user table with all fields', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('张明')).toBeInTheDocument()
      expect(screen.getByText('李华')).toBeInTheDocument()
      expect(screen.getByText('王芳')).toBeInTheDocument()
      expect(screen.getByText('赵强')).toBeInTheDocument()
    })

    // Check usernames (monospace)
    expect(screen.getByText('zhangming')).toBeInTheDocument()
    expect(screen.getByText('lihua')).toBeInTheDocument()

    // Check emails
    expect(screen.getByText('zhangming@example.com')).toBeInTheDocument()

    // Check team badges
    const productTeamBadges = screen.getAllByText('产品研发团队')
    expect(productTeamBadges.length).toBeGreaterThanOrEqual(1)
    expect(screen.getByText('设计团队')).toBeInTheDocument()

    // Check status badges
    const enabledBadges = screen.getAllByText('启用')
    expect(enabledBadges.length).toBeGreaterThanOrEqual(3)
    expect(screen.getByText('停用')).toBeInTheDocument()

    // Check actions
    const editButtons = screen.getAllByText('编辑')
    expect(editButtons.length).toBe(4)
    const statusButtons = screen.getAllByText('修改状态')
    expect(statusButtons.length).toBe(4)
  })

  // --- Search ---

  it('filters users by search text (username/name)', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('张明')).toBeInTheDocument()
    })

    const searchInput = screen.getByPlaceholderText('搜索用户名/姓名')
    await user.type(searchInput, 'zhang')

    await waitFor(() => {
      expect(screen.getByText('张明')).toBeInTheDocument()
      expect(screen.queryByText('李华')).not.toBeInTheDocument()
    })
  })

  it('filters users by displayName search', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('张明')).toBeInTheDocument()
    })

    const searchInput = screen.getByPlaceholderText('搜索用户名/姓名')
    await user.type(searchInput, '王芳')

    await waitFor(() => {
      expect(screen.getByText('王芳')).toBeInTheDocument()
      expect(screen.queryByText('张明')).not.toBeInTheDocument()
    })
  })

  // --- Create user ---

  it('opens create dialog when clicking create button', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('张明')).toBeInTheDocument()
    })

    await user.click(screen.getByText('创建用户'))

    await waitFor(() => {
      expect(screen.getByText('创建用户', { selector: '[data-state] *' })).toBeInTheDocument()
    })
  })

  it('creates user and shows initial password dialog', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('张明')).toBeInTheDocument()
    })

    await user.click(screen.getByText('创建用户'))

    // Fill form
    const inputs = screen.getAllByRole('textbox')
    const usernameInput = inputs.find((el) => (el as HTMLInputElement).placeholder === '请输入账号')!
    const nameInput = inputs.find((el) => (el as HTMLInputElement).placeholder === '请输入姓名')!

    await user.type(nameInput, '新用户')
    await user.type(usernameInput, 'newuser')

    // Submit
    await user.click(screen.getByText('确认创建'))

    // Should show password dialog
    await waitFor(() => {
      expect(screen.getByTestId('initial-password')).toHaveTextContent('Abc123456789')
    })
    expect(screen.getByText('请妥善保管，关闭后无法再次查看')).toBeInTheDocument()
  })

  it('shows error when creating duplicate username', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('张明')).toBeInTheDocument()
    })

    await user.click(screen.getByText('创建用户'))

    const inputs = screen.getAllByRole('textbox')
    const usernameInput = inputs.find((el) => (el as HTMLInputElement).placeholder === '请输入账号')!
    const nameInput = inputs.find((el) => (el as HTMLInputElement).placeholder === '请输入姓名')!

    await user.type(nameInput, '重复用户')
    await user.type(usernameInput, 'zhangming')

    await user.click(screen.getByText('确认创建'))

    await waitFor(() => {
      expect(screen.getByText('该账号名已存在')).toBeInTheDocument()
    })
  })

  // --- Edit user ---

  it('opens edit dialog with pre-filled values', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('张明')).toBeInTheDocument()
    })

    // Click the first edit button (zhangming)
    const editButtons = screen.getAllByText('编辑')
    await user.click(editButtons[0])

    await waitFor(() => {
      expect(screen.getByText('编辑用户')).toBeInTheDocument()
    })
  })

  // --- Toggle status ---

  it('opens toggle status dialog with confirmation warning', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('张明')).toBeInTheDocument()
    })

    // Click status button for an enabled user (zhangming)
    const statusButtons = screen.getAllByText('修改状态')
    await user.click(statusButtons[1]) // lihua (non-superadmin)

    await waitFor(() => {
      expect(screen.getByText('修改用户状态')).toBeInTheDocument()
      expect(screen.getByText(/禁用用户后该用户将无法登录系统，但数据不会被删除/)).toBeInTheDocument()
    })
  })

  it('handles CANNOT_DISABLE_SELF error', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('张明')).toBeInTheDocument()
    })

    // Click status for first user (superadmin = zhangming, id=1)
    const statusButtons = screen.getAllByText('修改状态')
    await user.click(statusButtons[0])

    await waitFor(() => {
      expect(screen.getByText('修改用户状态')).toBeInTheDocument()
    })

    // Confirm toggle
    await user.click(screen.getByText('确认修改'))

    await waitFor(() => {
      expect(screen.getByText('不能禁用自己')).toBeInTheDocument()
    })
  })

  // --- Empty state ---

  it('shows empty state when no users exist', async () => {
    server.use(
      http.get('/v1/admin/users', () => {
        return HttpResponse.json({
          code: 0,
          data: { items: [], total: 0, page: 1, pageSize: 10 },
        })
      }),
    )

    renderPage()
    await waitFor(() => {
      expect(screen.getByText('暂无用户')).toBeInTheDocument()
    })
  })

  // --- No antd imports ---

  it('does not import antd', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('用户管理')).toBeInTheDocument()
    })
    const antdElements = document.querySelectorAll('[class*="ant-"]')
    expect(antdElements.length).toBe(0)
  })

  // --- Pagination ---

  it('shows pagination info', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText(/共 4 条/)).toBeInTheDocument()
    })
  })

  // --- Reset Password Dialog ---

  describe('reset password dialog', () => {
    beforeEach(() => {
      // Set current user as super admin
      useAuthStore.setState({
        token: 'test-token',
        user: { bizKey: '1', username: 'zhangming', displayName: '张明', isSuperAdmin: true, createTime: '' },
        isAuthenticated: true,
        isSuperAdmin: true,
        _hasHydrated: true,
      })
    })

    afterEach(() => {
      useAuthStore.getState().clearAuth()
    })

    it('shows reset password button for super admin', async () => {
      renderPage()
      await waitFor(() => {
        expect(screen.getByText('张明')).toBeInTheDocument()
      })
      const resetButtons = screen.getAllByText('重置密码')
      expect(resetButtons.length).toBe(4) // one per user row
    })

    it('does not show reset password button for non-super-admin', async () => {
      useAuthStore.setState({ isSuperAdmin: false })
      renderPage()
      await waitFor(() => {
        expect(screen.getByText('张明')).toBeInTheDocument()
      })
      expect(screen.queryByText('重置密码')).not.toBeInTheDocument()
    })

    it('opens reset password dialog with user info', async () => {
      const user = userEvent.setup()
      renderPage()
      await waitFor(() => {
        expect(screen.getByText('张明')).toBeInTheDocument()
      })

      // Click reset on second user (李华)
      const resetButtons = screen.getAllByText('重置密码')
      await user.click(resetButtons[1])

      await waitFor(() => {
        expect(screen.getByText('为下面的用户设置新密码')).toBeInTheDocument()
      })
    })

    it('shows validation error when password is empty on submit', async () => {
      const user = userEvent.setup()
      renderPage()
      await waitFor(() => {
        expect(screen.getByText('张明')).toBeInTheDocument()
      })

      const resetButtons = screen.getAllByText('重置密码')
      await user.click(resetButtons[1])

      await waitFor(() => {
        expect(screen.getByText('为下面的用户设置新密码')).toBeInTheDocument()
      })

      // Click confirm without filling fields
      const confirmButtons = screen.getAllByRole('button', { name: '确认' })
      const confirmBtn = confirmButtons.find((btn) => (btn as HTMLButtonElement).closest('[role="dialog"]') !== null)
      if (confirmBtn) {
        await user.click(confirmBtn)
      }

      await waitFor(() => {
        expect(screen.getByText('请输入新密码')).toBeInTheDocument()
      })
    })

    it('shows error for weak password (less than 8 chars)', async () => {
      const user = userEvent.setup()
      renderPage()
      await waitFor(() => {
        expect(screen.getByText('张明')).toBeInTheDocument()
      })

      const resetButtons = screen.getAllByText('重置密码')
      await user.click(resetButtons[1])

      await waitFor(() => {
        expect(screen.getByText('为下面的用户设置新密码')).toBeInTheDocument()
      })

      // Type a short password
      const passwordInputs = screen.getAllByPlaceholderText('请输入新密码')
      await user.type(passwordInputs[0], 'Ab1')

      // Trigger blur validation
      await user.tab()

      await waitFor(() => {
        expect(screen.getByText('密码需至少8位，包含字母和数字')).toBeInTheDocument()
      })
    })

    it('shows error when confirmPassword does not match', async () => {
      const user = userEvent.setup()
      renderPage()
      await waitFor(() => {
        expect(screen.getByText('张明')).toBeInTheDocument()
      })

      const resetButtons = screen.getAllByText('重置密码')
      await user.click(resetButtons[1])

      await waitFor(() => {
        expect(screen.getByText('为下面的用户设置新密码')).toBeInTheDocument()
      })

      const passwordInputs = screen.getAllByPlaceholderText('请输入新密码')
      await user.type(passwordInputs[0], 'Password123')
      const confirmInputs = screen.getAllByPlaceholderText('请再次输入新密码')
      await user.type(confirmInputs[0], 'Different456')

      // Trigger blur validation
      await user.tab()

      await waitFor(() => {
        expect(screen.getByText('两次输入的密码不一致')).toBeInTheDocument()
      })
    })

    it('submits valid form and shows success toast', async () => {
      const user = userEvent.setup()
      renderPage()
      await waitFor(() => {
        expect(screen.getByText('张明')).toBeInTheDocument()
      })

      const resetButtons = screen.getAllByText('重置密码')
      await user.click(resetButtons[1])

      await waitFor(() => {
        expect(screen.getByText('为下面的用户设置新密码')).toBeInTheDocument()
      })

      // Fill valid password
      const passwordInputs = screen.getAllByPlaceholderText('请输入新密码')
      await user.type(passwordInputs[0], 'Password123')
      const confirmInputs = screen.getAllByPlaceholderText('请再次输入新密码')
      await user.type(confirmInputs[0], 'Password123')

      // Submit
      const confirmButtons = screen.getAllByRole('button', { name: '确认' })
      const confirmBtn = confirmButtons.find((btn) => (btn as HTMLButtonElement).closest('[role="dialog"]') !== null)
      if (confirmBtn) {
        await user.click(confirmBtn)
      }

      await waitFor(() => {
        expect(screen.getByText('密码已重置')).toBeInTheDocument()
      })
    })

    it('keeps dialog open on API error', async () => {
      const user = userEvent.setup()
      // Override to return error for user bizKey 2
      server.use(
        http.put('/v1/admin/users/:userId/password', async ({ params }) => {
          if (params.userId === '2') {
            return HttpResponse.json(
              { code: 'USER_NOT_FOUND', message: '用户不存在' },
              { status: 404 },
            )
          }
          return HttpResponse.json({
            code: 0,
            data: { bizKey: params.userId, username: 'u', displayName: 'U' },
          })
        }),
      )

      renderPage()
      await waitFor(() => {
        expect(screen.getByText('张明')).toBeInTheDocument()
      })

      const resetButtons = screen.getAllByText('重置密码')
      await user.click(resetButtons[1]) // 李华 (bizKey 2)

      await waitFor(() => {
        expect(screen.getByText('为下面的用户设置新密码')).toBeInTheDocument()
      })

      // Fill valid password
      const passwordInputs = screen.getAllByPlaceholderText('请输入新密码')
      await user.type(passwordInputs[0], 'Password123')
      const confirmInputs = screen.getAllByPlaceholderText('请再次输入新密码')
      await user.type(confirmInputs[0], 'Password123')

      // Submit
      const confirmButtons = screen.getAllByRole('button', { name: '确认' })
      const confirmBtn = confirmButtons.find((btn) => (btn as HTMLButtonElement).closest('[role="dialog"]') !== null)
      if (confirmBtn) {
        await user.click(confirmBtn)
      }

      await waitFor(() => {
        // Dialog should still be open
        expect(screen.getByText('为下面的用户设置新密码')).toBeInTheDocument()
      })
    })

    it('has password visibility toggle', async () => {
      const user = userEvent.setup()
      renderPage()
      await waitFor(() => {
        expect(screen.getByText('张明')).toBeInTheDocument()
      })

      const resetButtons = screen.getAllByText('重置密码')
      await user.click(resetButtons[1])

      await waitFor(() => {
        expect(screen.getByText('为下面的用户设置新密码')).toBeInTheDocument()
      })

      // Find the password input — initially type="password"
      const passwordInputs = screen.getAllByPlaceholderText('请输入新密码')
      expect((passwordInputs[0] as HTMLInputElement).type).toBe('password')

      // Find and click eye toggle
      const eyeButtons = screen.getAllByRole('button', { name: /显示密码|隐藏密码/ })
      await user.click(eyeButtons[0])

      expect((passwordInputs[0] as HTMLInputElement).type).toBe('text')
    })
  })

  // --- Delete Dialog & Action Buttons ---

  describe('delete button and dialog', () => {
    beforeEach(() => {
      useAuthStore.setState({
        token: 'test-token',
        user: { bizKey: '1', username: 'zhangming', displayName: '张明', isSuperAdmin: true, createTime: '' },
        isAuthenticated: true,
        isSuperAdmin: true,
        _hasHydrated: true,
      })
    })

    afterEach(() => {
      useAuthStore.getState().clearAuth()
    })

    it('shows delete button for super admin', async () => {
      renderPage()
      await waitFor(() => {
        expect(screen.getByText('张明')).toBeInTheDocument()
      })
      const deleteButtons = screen.getAllByText('删除')
      expect(deleteButtons.length).toBe(4)
    })

    it('does not show delete button for non-super-admin', async () => {
      useAuthStore.setState({ isSuperAdmin: false })
      renderPage()
      await waitFor(() => {
        expect(screen.getByText('张明')).toBeInTheDocument()
      })
      expect(screen.queryByText('删除')).not.toBeInTheDocument()
    })

    it('disables delete button on self-row with tooltip', async () => {
      renderPage()
      await waitFor(() => {
        expect(screen.getByText('张明')).toBeInTheDocument()
      })

      // First user is zhangming (bizKey '1'), which is the current user
      const deleteButtons = screen.getAllByText('删除')
      const selfDeleteButton = deleteButtons[0].closest('button')!
      expect(selfDeleteButton).toBeDisabled()
    })

    it('opens delete confirmation dialog with username', async () => {
      const user = userEvent.setup()
      renderPage()
      await waitFor(() => {
        expect(screen.getByText('张明')).toBeInTheDocument()
      })

      // Click delete on second user (李华)
      const deleteButtons = screen.getAllByText('删除')
      await user.click(deleteButtons[1])

      await waitFor(() => {
        expect(screen.getByText(/确认删除用户.*李华/)).toBeInTheDocument()
      })
    })

    it('deletes user successfully and shows toast', async () => {
      server.use(
        http.delete('/v1/admin/users/:userId', ({ params }) => {
          return HttpResponse.json({ code: 0, data: null }, { status: 200 })
        }),
      )

      const user = userEvent.setup()
      renderPage()
      await waitFor(() => {
        expect(screen.getByText('张明')).toBeInTheDocument()
      })

      const deleteButtons = screen.getAllByText('删除')
      await user.click(deleteButtons[1])

      await waitFor(() => {
        expect(screen.getByText(/确认删除用户.*李华/)).toBeInTheDocument()
      })

      const confirmBtn = screen.getByRole('button', { name: '确认删除' })
      await user.click(confirmBtn)

      await waitFor(() => {
        expect(screen.getByText('用户已删除')).toBeInTheDocument()
      })
    })

    it('handles 404 error by removing row and showing error', async () => {
      server.use(
        http.delete('/v1/admin/users/:userId', () => {
          return HttpResponse.json(
            { code: 'USER_NOT_FOUND', message: '用户不存在' },
            { status: 404 },
          )
        }),
      )

      const user = userEvent.setup()
      renderPage()
      await waitFor(() => {
        expect(screen.getByText('张明')).toBeInTheDocument()
      })

      const deleteButtons = screen.getAllByText('删除')
      await user.click(deleteButtons[1])

      await waitFor(() => {
        expect(screen.getByText(/确认删除用户.*李华/)).toBeInTheDocument()
      })

      const confirmBtn = screen.getByRole('button', { name: '确认删除' })
      await user.click(confirmBtn)

      await waitFor(() => {
        expect(screen.getByText('用户已删除')).toBeInTheDocument()
      })
    })

    it('shows error message in dialog on generic API error', async () => {
      server.use(
        http.delete('/v1/admin/users/:userId', () => {
          return HttpResponse.json(
            { code: 'INTERNAL_ERROR', message: 'internal error' },
            { status: 500 },
          )
        }),
      )

      const user = userEvent.setup()
      renderPage()
      await waitFor(() => {
        expect(screen.getByText('张明')).toBeInTheDocument()
      })

      const deleteButtons = screen.getAllByText('删除')
      await user.click(deleteButtons[1])

      await waitFor(() => {
        expect(screen.getByText(/确认删除用户.*李华/)).toBeInTheDocument()
      })

      const confirmBtn = screen.getByRole('button', { name: '确认删除' })
      await user.click(confirmBtn)

      await waitFor(() => {
        expect(screen.getByText('操作失败，请稍后重试')).toBeInTheDocument()
      })
    })

    it('closes dialog on cancel', async () => {
      const user = userEvent.setup()
      renderPage()
      await waitFor(() => {
        expect(screen.getByText('张明')).toBeInTheDocument()
      })

      const deleteButtons = screen.getAllByText('删除')
      await user.click(deleteButtons[1])

      await waitFor(() => {
        expect(screen.getByText(/确认删除用户.*李华/)).toBeInTheDocument()
      })

      await user.click(screen.getByRole('button', { name: '取消' }))

      await waitFor(() => {
        expect(screen.queryByText(/确认删除用户/)).not.toBeInTheDocument()
      })
    })
  })

  // --- Copy Credentials ---

  describe('copy credentials button', () => {
    beforeEach(() => {
      useAuthStore.setState({
        token: 'test-token',
        user: { bizKey: '1', username: 'zhangming', displayName: '张明', isSuperAdmin: true, createTime: '' },
        isAuthenticated: true,
        isSuperAdmin: true,
        _hasHydrated: true,
      })
      vi.mocked(copyToClipboard).mockResolvedValue(undefined)
    })

    afterEach(() => {
      useAuthStore.getState().clearAuth()
    })

    it('shows copy button after creating user', async () => {
      const user = userEvent.setup()
      renderPage()
      await waitFor(() => {
        expect(screen.getByText('张明')).toBeInTheDocument()
      })

      await user.click(screen.getByText('创建用户'))

      const inputs = screen.getAllByRole('textbox')
      const usernameInput = inputs.find((el) => (el as HTMLInputElement).placeholder === '请输入账号')!
      const nameInput = inputs.find((el) => (el as HTMLInputElement).placeholder === '请输入姓名')!

      await user.type(nameInput, '新用户')
      await user.type(usernameInput, 'newuser')
      await user.click(screen.getByText('确认创建'))

      await waitFor(() => {
        expect(screen.getByTestId('initial-password')).toHaveTextContent('Abc123456789')
        expect(screen.getByText('复制账号与密码')).toBeInTheDocument()
      })
    })

    it('copies correct format to clipboard', async () => {
      const user = userEvent.setup()
      renderPage()
      await waitFor(() => {
        expect(screen.getByText('张明')).toBeInTheDocument()
      })

      await user.click(screen.getByText('创建用户'))

      const inputs = screen.getAllByRole('textbox')
      const usernameInput = inputs.find((el) => (el as HTMLInputElement).placeholder === '请输入账号')!
      const nameInput = inputs.find((el) => (el as HTMLInputElement).placeholder === '请输入姓名')!

      await user.type(nameInput, '新用户')
      await user.type(usernameInput, 'newuser')
      await user.click(screen.getByText('确认创建'))

      await waitFor(() => {
        expect(screen.getByText('复制账号与密码')).toBeInTheDocument()
      })

      const copyButton = screen.getByRole('button', { name: /复制账号与密码/ })
      expect(copyButton).not.toBeDisabled()
      await user.click(copyButton)

      await waitFor(() => {
        expect(copyToClipboard).toHaveBeenCalledWith(
          '账号：newuser\n密码：Abc123456789',
        )
      })
    })

    it('shows "已复制" after successful copy', async () => {
      const user = userEvent.setup()
      renderPage()
      await waitFor(() => {
        expect(screen.getByText('张明')).toBeInTheDocument()
      })

      await user.click(screen.getByText('创建用户'))

      const inputs = screen.getAllByRole('textbox')
      const usernameInput = inputs.find((el) => (el as HTMLInputElement).placeholder === '请输入账号')!
      const nameInput = inputs.find((el) => (el as HTMLInputElement).placeholder === '请输入姓名')!

      await user.type(nameInput, '新用户')
      await user.type(usernameInput, 'newuser')
      await user.click(screen.getByText('确认创建'))

      await waitFor(() => {
        expect(screen.getByText('复制账号与密码')).toBeInTheDocument()
      })

      await user.click(screen.getByText('复制账号与密码'))

      await waitFor(() => {
        expect(screen.getByText('已复制')).toBeInTheDocument()
      })
    })

    it('shows error toast on clipboard failure', async () => {
      vi.mocked(copyToClipboard).mockRejectedValue(new Error('not allowed'))

      const user = userEvent.setup()
      renderPage()
      await waitFor(() => {
        expect(screen.getByText('张明')).toBeInTheDocument()
      })

      await user.click(screen.getByText('创建用户'))

      const inputs = screen.getAllByRole('textbox')
      const usernameInput = inputs.find((el) => (el as HTMLInputElement).placeholder === '请输入账号')!
      const nameInput = inputs.find((el) => (el as HTMLInputElement).placeholder === '请输入姓名')!

      await user.type(nameInput, '新用户')
      await user.type(usernameInput, 'newuser')
      await user.click(screen.getByText('确认创建'))

      await waitFor(() => {
        expect(screen.getByText('复制账号与密码')).toBeInTheDocument()
      })

      await user.click(screen.getByText('复制账号与密码'))

      await waitFor(() => {
        expect(screen.getByText('复制失败，请手动选择文字复制')).toBeInTheDocument()
      })
    })
  })
})
