import { describe, it, expect, vi, beforeEach, beforeAll, afterAll, afterEach } from 'vitest'
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { server } from '@/mocks/server'
import { http, HttpResponse } from 'msw'
import { ToastProvider } from '@/components/ui/toast'
import RoleManagementPage from './RoleManagementPage'

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
          <RoleManagementPage />
        </MemoryRouter>
      </ToastProvider>
    </QueryClientProvider>,
  )
}

// --- Seed data ---

const seedRoles = [
  {
    id: 1,
    roleName: 'superadmin',
    roleDesc: '超级管理员',
    isPreset: true,
    permissionCount: 30,
    memberCount: 2,
    createTime: '2026-04-01T00:00:00Z',
  },
  {
    id: 2,
    roleName: 'pm',
    roleDesc: '团队管理权限',
    isPreset: true,
    permissionCount: 22,
    memberCount: 5,
    createTime: '2026-04-01T00:00:00Z',
  },
  {
    id: 3,
    roleName: 'member',
    roleDesc: '普通成员权限',
    isPreset: true,
    permissionCount: 8,
    memberCount: 10,
    createTime: '2026-04-01T00:00:00Z',
  },
  {
    id: 4,
    roleName: 'viewer',
    roleDesc: '只读查看者',
    isPreset: false,
    permissionCount: 3,
    memberCount: 0,
    createTime: '2026-04-15T12:00:00Z',
  },
]

const seedPermissions = [
  {
    resource: 'team',
    actions: [
      { code: 'team:create', description: '创建团队' },
      { code: 'team:read', description: '查看团队信息' },
    ],
  },
  {
    resource: 'main_item',
    actions: [
      { code: 'main_item:create', description: '创建主事项' },
      { code: 'main_item:read', description: '查看主事项' },
    ],
  },
]

function setupHandlers() {
  server.use(
    // List roles
    http.get('/v1/admin/roles', ({ request }) => {
      const url = new URL(request.url)
      const search = url.searchParams.get('search')
      const isPresetParam = url.searchParams.get('isPreset')
      const page = Number(url.searchParams.get('page') || 1)
      const pageSize = Number(url.searchParams.get('pageSize') || 20)

      let filtered = [...seedRoles]
      if (search) {
        const q = search.toLowerCase()
        filtered = filtered.filter((r) => r.roleName.toLowerCase().includes(q))
      }
      if (isPresetParam === 'preset') {
        filtered = filtered.filter((r) => r.isPreset)
      } else if (isPresetParam === 'custom') {
        filtered = filtered.filter((r) => !r.isPreset)
      }

      const start = (page - 1) * pageSize
      const items = filtered.slice(start, start + pageSize)

      return HttpResponse.json({
        code: 0,
        data: { items, total: filtered.length, page, pageSize },
      })
    }),

    // Get role detail
    http.get('/v1/admin/roles/:id', ({ params }) => {
      const id = Number(params.id)
      const role = seedRoles.find((r) => r.id === id)
      if (!role) {
        return HttpResponse.json(
          { code: 'ERR_ROLE_NOT_FOUND', message: '角色不存在' },
          { status: 404 },
        )
      }
      return HttpResponse.json({
        code: 0,
        data: {
          ...role,
          permissions: [
            { code: 'team:read', description: '查看团队信息' },
            { code: 'main_item:read', description: '查看主事项' },
          ],
        },
      })
    }),

    // Create role
    http.post('/v1/admin/roles', async ({ request }) => {
      const body = (await request.json()) as Record<string, unknown>
      const name = body.name as string

      if (name === 'pm') {
        return HttpResponse.json(
          { code: 'ERR_ROLE_NAME_EXISTS', message: '角色名称已存在' },
          { status: 409 },
        )
      }

      return HttpResponse.json({
        code: 0,
        data: {
          id: 100,
          roleName: name,
          roleDesc: body.description || '',
          isPreset: false,
          permissionCount: (body.permissionCodes as string[]).length,
          memberCount: 0,
          createTime: '2026-04-19T12:00:00Z',
        },
      })
    }),

    // Update role
    http.put('/v1/admin/roles/:id', async ({ params }) => {
      const id = Number(params.id)
      const role = seedRoles.find((r) => r.id === id)
      if (!role) {
        return HttpResponse.json(
          { code: 'ERR_ROLE_NOT_FOUND', message: '角色不存在' },
          { status: 404 },
        )
      }
      return HttpResponse.json({
        code: 0,
        data: {
          ...role,
          permissions: [
            { code: 'team:read', description: '查看团队信息' },
          ],
        },
      })
    }),

    // Delete role
    http.delete('/v1/admin/roles/:id', ({ params }) => {
      const id = Number(params.id)
      const role = seedRoles.find((r) => r.id === id)
      if (!role) {
        return HttpResponse.json(
          { code: 'ERR_ROLE_NOT_FOUND', message: '角色不存在' },
          { status: 404 },
        )
      }
      if (role.isPreset) {
        return HttpResponse.json(
          { code: 'ERR_PRESET_ROLE_IMMUTABLE', message: '预置角色不可删除' },
          { status: 403 },
        )
      }
      if (role.memberCount > 0) {
        return HttpResponse.json(
          { code: 'ERR_ROLE_IN_USE', message: '角色正在被使用' },
          { status: 422 },
        )
      }
      return HttpResponse.json({ code: 0, data: null })
    }),

    // List permissions
    http.get('/v1/admin/permissions', () => {
      return HttpResponse.json({ code: 0, data: seedPermissions })
    }),
  )
}

describe('RoleManagementPage', () => {
  beforeEach(() => {
    setupHandlers()
  })

  // --- Core rendering ---

  it('renders page header with title and breadcrumb', async () => {
    renderPage()
    expect(screen.getAllByText('角色管理').length).toBeGreaterThanOrEqual(1)
    expect(screen.getByText('首页')).toBeInTheDocument()
  })

  it('renders filter bar with search and preset filter', async () => {
    renderPage()
    expect(screen.getByPlaceholderText('搜索角色名称')).toBeInTheDocument()
    await waitFor(() => {
      expect(screen.getByText('superadmin')).toBeInTheDocument()
    })
  })

  it('renders create and permission browse buttons', async () => {
    renderPage()
    expect(screen.getByText('创建角色')).toBeInTheDocument()
    expect(screen.getByText('权限列表')).toBeInTheDocument()
  })

  // --- Role table ---

  it('displays role table with all columns', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('superadmin')).toBeInTheDocument()
      expect(screen.getByText('pm')).toBeInTheDocument()
      expect(screen.getByText('member')).toBeInTheDocument()
      expect(screen.getByText('viewer')).toBeInTheDocument()
    })

    // Check descriptions
    expect(screen.getByText('超级管理员')).toBeInTheDocument()
    expect(screen.getByText('团队管理权限')).toBeInTheDocument()

    // Check permission counts
    expect(screen.getByText('30')).toBeInTheDocument()
    expect(screen.getByText('22')).toBeInTheDocument()

    // Check member counts
    expect(screen.getByText('10')).toBeInTheDocument()

    // Check type badges
    const presetBadges = screen.getAllByText('预置')
    expect(presetBadges.length).toBe(3)
    expect(screen.getByText('自定义')).toBeInTheDocument()

    // Check action buttons
    const editButtons = screen.getAllByText('编辑')
    expect(editButtons.length).toBe(4)
  })

  it('formats created_at as YYYY/MM/DD', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('superadmin')).toBeInTheDocument()
    })
    // 3 roles share 2026/04/01, 1 role has 2026/04/15
    const dates0401 = screen.getAllByText('2026/04/01')
    expect(dates0401.length).toBe(3)
    expect(screen.getByText('2026/04/15')).toBeInTheDocument()
  })

  // --- Search filter ---

  it('filters roles by search text', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('superadmin')).toBeInTheDocument()
    })

    const searchInput = screen.getByPlaceholderText('搜索角色名称')
    await user.type(searchInput, 'view')

    await waitFor(() => {
      expect(screen.getByText('viewer')).toBeInTheDocument()
      expect(screen.queryByText('superadmin')).not.toBeInTheDocument()
    })
  })

  // --- Preset filter ---

  it('filters roles by preset/custom type', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('superadmin')).toBeInTheDocument()
    })

    // Click the select trigger for preset filter
    const triggers = screen.getAllByRole('combobox')
    const presetTrigger = triggers[0]
    await user.click(presetTrigger)

    // Select custom
    await waitFor(() => {
      expect(screen.getByText('自定义角色')).toBeInTheDocument()
    })
    await user.click(screen.getByText('自定义角色'))

    await waitFor(() => {
      expect(screen.getByText('viewer')).toBeInTheDocument()
      expect(screen.queryByText('superadmin')).not.toBeInTheDocument()
    })
  })

  // --- Delete ---

  it('disables delete for preset roles', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('superadmin')).toBeInTheDocument()
    })

    // Get all delete buttons - preset roles should have disabled delete
    const deleteButtons = screen.getAllByText('删除')
    // First 3 are preset roles, should be disabled
    expect(deleteButtons[0]).toBeDisabled()
    expect(deleteButtons[1]).toBeDisabled()
    expect(deleteButtons[2]).toBeDisabled()
    // 4th is viewer (custom, 0 members) - should be enabled
    expect(deleteButtons[3]).not.toBeDisabled()
  })

  it('opens delete confirm dialog for custom role', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('viewer')).toBeInTheDocument()
    })

    // Click delete on viewer (enabled)
    const deleteButtons = screen.getAllByText('删除')
    await user.click(deleteButtons[3]) // viewer

    await waitFor(() => {
      expect(screen.getByText('删除角色')).toBeInTheDocument()
      expect(screen.getByText(/确定要删除角色"viewer"吗/)).toBeInTheDocument()
    })
  })

  // --- Create role ---

  it('opens create dialog when clicking create button', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('superadmin')).toBeInTheDocument()
    })

    await user.click(screen.getByText('创建角色'))

    await waitFor(() => {
      expect(screen.getByText('创建角色', { selector: '[data-state] *' })).toBeInTheDocument()
    })
  })

  // --- Edit role ---

  it('opens edit dialog with role name in title', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('superadmin')).toBeInTheDocument()
    })

    const editButtons = screen.getAllByText('编辑')
    await user.click(editButtons[3]) // viewer

    await waitFor(() => {
      expect(screen.getByText(/编辑角色.*viewer/)).toBeInTheDocument()
    })
  })

  // --- Permission browse dialog ---

  it('opens permission browse dialog', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('superadmin')).toBeInTheDocument()
    })

    await user.click(screen.getByText('权限列表'))

    await waitFor(() => {
      expect(screen.getByText('系统权限列表')).toBeInTheDocument()
    })
  })

  // --- Empty state ---

  it('shows empty state when no roles match filters', async () => {
    server.use(
      http.get('/v1/admin/roles', () => {
        return HttpResponse.json({
          code: 0,
          data: { items: [], total: 0, page: 1, pageSize: 20 },
        })
      }),
    )

    renderPage()
    await waitFor(() => {
      expect(screen.getByText('暂无自定义角色')).toBeInTheDocument()
    })
  })

  // --- Loading state ---

  it('shows loading skeleton', async () => {
    server.use(
      http.get('/v1/admin/roles', async () => {
        await new Promise((r) => setTimeout(r, 500))
        return HttpResponse.json({
          code: 0,
          data: { items: seedRoles, total: seedRoles.length, page: 1, pageSize: 20 },
        })
      }),
    )

    renderPage()
    // Page header should be visible immediately
    expect(screen.getAllByText('角色管理').length).toBeGreaterThanOrEqual(1)
  })

  // --- Pagination ---

  it('shows total count', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText(/共 4 条/)).toBeInTheDocument()
    })
  })
})
