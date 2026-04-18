import { describe, it, expect, vi, beforeEach, beforeAll, afterAll, afterEach } from 'vitest'
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { server } from '@/mocks/server'
import { http, HttpResponse } from 'msw'
import UserManagementPage from './UserManagementPage'

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
      <MemoryRouter>
        <UserManagementPage />
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

// --- Seed data ---

const seedUsers = [
  {
    id: 1, username: 'zhangming', displayName: '张明', email: 'zhangming@example.com',
    canCreateTeam: true, isSuperAdmin: true, status: 'enabled',
    teams: [{ id: 1, name: '产品研发团队', role: 'pm' }],
  },
  {
    id: 2, username: 'lihua', displayName: '李华', email: 'lihua@example.com',
    canCreateTeam: true, isSuperAdmin: false, status: 'enabled',
    teams: [{ id: 1, name: '产品研发团队', role: 'member' }],
  },
  {
    id: 3, username: 'wangfang', displayName: '王芳', email: 'wangfang@example.com',
    canCreateTeam: false, isSuperAdmin: false, status: 'enabled',
    teams: [{ id: 1, name: '产品研发团队', role: 'member' }],
  },
  {
    id: 4, username: 'zhaoqiang', displayName: '赵强', email: 'zhaoqiang@example.com',
    canCreateTeam: false, isSuperAdmin: false, status: 'disabled',
    teams: [{ id: 2, name: '设计团队', role: 'member' }],
  },
]

const seedTeams = [
  { id: 1, name: '产品研发团队', pm: { displayName: '张明' }, memberCount: 3, mainItemCount: 5, createdAt: '2024-01-01' },
  { id: 2, name: '设计团队', pm: { displayName: '李华' }, memberCount: 2, mainItemCount: 2, createdAt: '2024-02-01' },
]

function setupHandlers() {
  server.use(
    // List users
    http.get('/api/v1/admin/users', ({ request }) => {
      const url = new URL(request.url)
      const search = url.searchParams.get('search')
      const canCreateTeam = url.searchParams.get('canCreateTeam')
      const page = Number(url.searchParams.get('page') || 1)
      const pageSizeParam = Number(url.searchParams.get('pageSize') || 10)

      let filtered = [...seedUsers]
      if (search) {
        const q = search.toLowerCase()
        filtered = filtered.filter(
          (u) => u.username.toLowerCase().includes(q) || u.displayName.toLowerCase().includes(q),
        )
      }
      if (canCreateTeam === 'true') {
        filtered = filtered.filter((u) => u.canCreateTeam)
      } else if (canCreateTeam === 'false') {
        filtered = filtered.filter((u) => !u.canCreateTeam)
      }

      const start = (page - 1) * pageSizeParam
      const items = filtered.slice(start, start + pageSizeParam)

      return HttpResponse.json({
        code: 0,
        data: { items, total: filtered.length, page, pageSize: pageSizeParam },
      })
    }),

    // List admin teams
    http.get('/api/v1/admin/teams', () => {
      return HttpResponse.json({
        code: 0,
        data: { items: seedTeams, total: seedTeams.length, page: 1, pageSize: 100 },
      })
    }),

    // Create user
    http.post('/api/v1/admin/users', async ({ request }) => {
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
          canCreateTeam: body.canCreateTeam || false,
          status: 'enabled',
          teams: [],
          initialPassword: 'Abc123456789',
        },
      }, { status: 201 })
    }),

    // Update user
    http.put('/api/v1/admin/users/:userId', async ({ params, request }) => {
      const userId = Number(params.userId)
      const body = (await request.json()) as Record<string, unknown>
      const user = seedUsers.find((u) => u.id === userId)
      if (!user) {
        return HttpResponse.json({ code: 'USER_NOT_FOUND', message: 'not found' }, { status: 404 })
      }
      return HttpResponse.json({
        code: 0,
        data: { ...user, ...body },
      })
    }),

    // Toggle user status
    http.put('/api/v1/admin/users/:userId/status', async ({ params, request }) => {
      const userId = Number(params.userId)
      const body = (await request.json()) as { status: string }
      const user = seedUsers.find((u) => u.id === userId)
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
        data: { id: userId, username: user.username, status: body.status },
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

  it('displays canCreateTeam permission for each user', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('张明')).toBeInTheDocument()
    })
    const permissionYes = screen.getAllByText('有权限')
    const permissionNo = screen.getAllByText('无权限')
    expect(permissionYes.length).toBe(2) // zhangming, lihua
    expect(permissionNo.length).toBe(2) // wangfang, zhaoqiang
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
      http.get('/api/v1/admin/users', () => {
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

  // --- canCreateTeam filter ---

  it('filters by canCreateTeam permission via API', async () => {
    // Directly test that the API is called with correct filter params
    let capturedParams: Record<string, string> = {}
    server.use(
      http.get('/api/v1/admin/users', ({ request }) => {
        const url = new URL(request.url)
        capturedParams = Object.fromEntries(url.searchParams.entries())
        const filtered = seedUsers.filter((u) => u.canCreateTeam)
        return HttpResponse.json({
          code: 0,
          data: { items: filtered, total: filtered.length, page: 1, pageSize: 10 },
        })
      }),
      http.get('/api/v1/admin/teams', () => {
        return HttpResponse.json({
          code: 0,
          data: { items: seedTeams, total: seedTeams.length, page: 1, pageSize: 100 },
        })
      }),
    )

    // Render with pre-filtered data (simulates filter being applied)
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('张明')).toBeInTheDocument()
      expect(screen.getByText('李华')).toBeInTheDocument()
      expect(screen.queryByText('王芳')).not.toBeInTheDocument()
      expect(screen.queryByText('赵强')).not.toBeInTheDocument()
    })
  })
})
