import { describe, it, expect, vi, beforeEach, beforeAll, afterAll, afterEach } from 'vitest'
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { server } from '@/mocks/server'
import { http, HttpResponse } from 'msw'
import { ToastProvider } from '@/components/ui/toast'

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
          <TeamManagementPage />
        </MemoryRouter>
      </ToastProvider>
    </QueryClientProvider>,
  )
}

// Lazy import to ensure MSW is ready
const { default: TeamManagementPage } = await import('./TeamManagementPage')
const { useAuthStore } = await import('@/store/auth')

// --- Seed data ---

const seedTeams = [
  {
    bizKey: '1',
    name: '产品研发团队',
    code: 'PROD',
    description: '负责核心产品的研发与迭代',
    pmKey: '1',
    createdAt: '2026-03-01T00:00:00Z',
    updatedAt: '2026-03-01T00:00:00Z',
  },
  {
    bizKey: '2',
    name: '设计团队',
    code: 'DESIGN',
    description: '负责 UI/UX 设计与交互体验优化',
    pmKey: '2',
    createdAt: '2026-03-15T00:00:00Z',
    updatedAt: '2026-03-15T00:00:00Z',
  },
  {
    bizKey: '3',
    name: '基础架构团队',
    code: 'INFRA',
    description: '负责底层基础设施与运维体系建设',
    pmKey: '3',
    createdAt: '2026-04-01T00:00:00Z',
    updatedAt: '2026-04-01T00:00:00Z',
  },
]

function setupHandlers() {
  server.use(
    // List teams — now returns paginated response
    http.get('/v1/teams', ({ request }) => {
      const url = new URL(request.url)
      const search = url.searchParams.get('search') || ''
      const page = Number(url.searchParams.get('page') || 1)
      const pageSize = Number(url.searchParams.get('pageSize') || 20)
      const filtered = search
        ? seedTeams.filter((t) => t.name.includes(search) || t.code.includes(search))
        : seedTeams
      const start = (page - 1) * pageSize
      const items = filtered.slice(start, start + pageSize)
      return HttpResponse.json({
        code: 0,
        data: { items, total: filtered.length, page, pageSize },
      })
    }),

    // Create team
    http.post('/v1/teams', async ({ request }) => {
      const body = (await request.json()) as { name: string; code?: string; description?: string }
      if (!body.name || !body.name.trim()) {
        return HttpResponse.json(
          { code: 'VALIDATION_ERROR', message: '团队名称不能为空' },
          { status: 422 },
        )
      }
      if (body.code === 'DUPE') {
        return HttpResponse.json(
          { code: 'TEAM_CODE_DUPLICATE', message: '该编码已被使用' },
          { status: 422 },
        )
      }
      return HttpResponse.json({
        code: 0,
        data: {
          id: 100,
          name: body.name,
          code: body.code || '',
          description: body.description || '',
          pmId: 1,
          createdAt: new Date().toISOString(),
          updatedAt: new Date().toISOString(),
        },
      }, { status: 201 })
    }),
  )
}

describe('TeamManagementPage', () => {
  beforeEach(() => {
    setupHandlers()
    useAuthStore.getState().setPermissions({
      isSuperAdmin: false,
      teamPermissions: { 1: ['team:create'] },
    })
  })

  // --- Core rendering ---

  it('renders page header with title and create button', async () => {
    renderPage()
    expect(screen.getByText('团队管理')).toBeInTheDocument()
    expect(screen.getByText('创建团队')).toBeInTheDocument()
  })

  it('renders team table with all teams', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('产品研发团队')).toBeInTheDocument()
      expect(screen.getByText('设计团队')).toBeInTheDocument()
      expect(screen.getByText('基础架构团队')).toBeInTheDocument()
    })
  })

  it('displays team descriptions', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('负责核心产品的研发与迭代')).toBeInTheDocument()
      expect(screen.getByText('负责 UI/UX 设计与交互体验优化')).toBeInTheDocument()
    })
  })

  it('displays team creation dates', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('产品研发团队')).toBeInTheDocument()
    })
    // Check date formatting for the first team (2026-03-01)
    expect(screen.getByText('2026/03/01')).toBeInTheDocument()
    expect(screen.getByText('2026/03/15')).toBeInTheDocument()
    expect(screen.getByText('2026/04/01')).toBeInTheDocument()
  })

  // --- Team name links ---

  it('team names are links to team detail page', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('产品研发团队')).toBeInTheDocument()
    })
    const link = screen.getByText('产品研发团队').closest('a')
    expect(link).toHaveAttribute('href', '/teams/1')
  })

  // --- Create team dialog ---

  it('opens create team dialog when clicking create button', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('产品研发团队')).toBeInTheDocument()
    })

    await user.click(screen.getByText('创建团队'))

    await waitFor(() => {
      expect(screen.getByText('创建团队', { selector: '[data-state] *' })).toBeInTheDocument()
    })
  })

  it('creates team successfully', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('产品研发团队')).toBeInTheDocument()
    })

    await user.click(screen.getByText('创建团队'))

    await waitFor(() => {
      expect(screen.getByText('创建团队', { selector: '[data-state] *' })).toBeInTheDocument()
    })

    // Fill form
    const nameInput = screen.getByPlaceholderText('请输入团队名称')
    await user.type(nameInput, '新团队')
    const codeInput = screen.getByPlaceholderText('如 FEAT、CORE')
    await user.type(codeInput, 'FEAT')

    // Submit
    await user.click(screen.getByText('确认创建'))

    // Dialog should close
    await waitFor(() => {
      expect(screen.queryByText('确认创建')).not.toBeInTheDocument()
    })
  })

  it('disables create button when name is empty', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('产品研发团队')).toBeInTheDocument()
    })

    await user.click(screen.getByText('创建团队'))

    await waitFor(() => {
      expect(screen.getByText('创建团队', { selector: '[data-state] *' })).toBeInTheDocument()
    })

    // Submit button should be disabled when name is empty
    const submitBtn = screen.getByText('确认创建')
    expect(submitBtn).toBeDisabled()
  })

  // --- Code field ---

  it('shows code input with correct placeholder', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => expect(screen.getByText('产品研发团队')).toBeInTheDocument())

    await user.click(screen.getByText('创建团队'))

    await waitFor(() => {
      expect(screen.getByPlaceholderText('如 FEAT、CORE')).toBeInTheDocument()
    })
  })

  it('shows validation error on blur when code is invalid', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => expect(screen.getByText('产品研发团队')).toBeInTheDocument())

    await user.click(screen.getByText('创建团队'))

    await waitFor(() => {
      expect(screen.getByPlaceholderText('如 FEAT、CORE')).toBeInTheDocument()
    })

    const codeInput = screen.getByPlaceholderText('如 FEAT、CORE')
    await user.type(codeInput, 'A')
    await user.tab()

    await waitFor(() => {
      expect(screen.getByText('CODE须为 2~6 位英文字母')).toBeInTheDocument()
    })
  })

  it('shows duplicate code error from backend', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => expect(screen.getByText('产品研发团队')).toBeInTheDocument())

    await user.click(screen.getByText('创建团队'))

    await waitFor(() => {
      expect(screen.getByPlaceholderText('请输入团队名称')).toBeInTheDocument()
    })

    await user.type(screen.getByPlaceholderText('请输入团队名称'), '新团队')
    await user.type(screen.getByPlaceholderText('如 FEAT、CORE'), 'DUPE')
    await user.click(screen.getByText('确认创建'))

    await waitFor(() => {
      expect(screen.getByText('该CODE已被使用')).toBeInTheDocument()
    })
  })

  it('renders Code column in team table', async () => {
    renderPage()
    await waitFor(() => expect(screen.getByText('产品研发团队')).toBeInTheDocument())

    expect(screen.getByText('Code')).toBeInTheDocument()
    expect(screen.getByText('PROD')).toBeInTheDocument()
    expect(screen.getByText('DESIGN')).toBeInTheDocument()
    expect(screen.getByText('INFRA')).toBeInTheDocument()
  })

  // --- Empty state ---

  it('shows empty state when no teams exist', async () => {
    server.use(
      http.get('/v1/teams', () => {
        return HttpResponse.json({ code: 0, data: [] })
      }),
    )

    renderPage()
    await waitFor(() => {
      expect(screen.getByText('暂无团队')).toBeInTheDocument()
    })
    expect(screen.getByText('点击上方按钮创建第一个团队')).toBeInTheDocument()
  })

  // --- No antd imports ---

  it('does not import antd', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('团队管理')).toBeInTheDocument()
    })
    const antdElements = document.querySelectorAll('[class*="ant-"]')
    expect(antdElements.length).toBe(0)
  })

  // --- Loading state ---

  it('shows loading state', async () => {
    server.use(
      http.get('/v1/teams', async () => {
        await new Promise((resolve) => setTimeout(resolve, 500))
        return HttpResponse.json({ code: 0, data: { items: seedTeams, total: seedTeams.length, page: 1, pageSize: 20 } })
      }),
    )

    renderPage()
    expect(screen.getByText('加载中...')).toBeInTheDocument()
  })

  // --- data-testid ---

  it('renders with data-testid', async () => {
    renderPage()
    expect(screen.getByTestId('team-management-page')).toBeInTheDocument()
  })

  // --- Refresh button ---

  it('renders refresh button in filter bar', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('产品研发团队')).toBeInTheDocument()
    })
    expect(screen.getByTestId('refresh-btn')).toBeInTheDocument()
  })

  it('refresh button refetches data', async () => {
    let callCount = 0
    server.use(
      http.get('/v1/teams', ({ request }) => {
        callCount++
        const url = new URL(request.url)
        const page = Number(url.searchParams.get('page') || 1)
        const pageSize = Number(url.searchParams.get('pageSize') || 10)
        return HttpResponse.json({
          code: 0,
          data: { items: seedTeams, total: seedTeams.length, page, pageSize },
        })
      }),
    )

    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('产品研发团队')).toBeInTheDocument()
    })

    const before = callCount
    await user.click(screen.getByTestId('refresh-btn'))
    await waitFor(() => {
      expect(callCount).toBeGreaterThan(before)
    })
  })

  // --- Pagination ---

  it('always shows pagination bar even with few teams', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('产品研发团队')).toBeInTheDocument()
    })
    expect(screen.getByRole('navigation', { name: 'Pagination' })).toBeInTheDocument()
    expect(screen.getByText('共 3 条')).toBeInTheDocument()
  })

  it('shows pagination when teams exceed page size', async () => {
    const manyTeams = Array.from({ length: 12 }, (_, i) => ({
      id: i + 1,
      name: `团队 ${i + 1}`,
      code: `T${String(i + 1).padStart(2, '0')}`,
      description: '',
      pmId: 1,
      pmDisplayName: '张明',
      createdAt: '2026-03-01T00:00:00Z',
      updatedAt: '2026-03-01T00:00:00Z',
    }))
    server.use(
      http.get('/v1/teams', ({ request }) => {
        const url = new URL(request.url)
        const page = Number(url.searchParams.get('page') || 1)
        const pageSize = Number(url.searchParams.get('pageSize') || 10)
        const start = (page - 1) * pageSize
        return HttpResponse.json({
          code: 0,
          data: { items: manyTeams.slice(start, start + pageSize), total: manyTeams.length, page, pageSize },
        })
      }),
    )

    renderPage()
    await waitFor(() => {
      expect(screen.getByText('团队 1')).toBeInTheDocument()
    })

    // First page shows 10 items
    expect(screen.getByText('团队 10')).toBeInTheDocument()
    expect(screen.queryByText('团队 11')).not.toBeInTheDocument()

    // Pagination nav is visible
    expect(screen.getByRole('navigation', { name: 'Pagination' })).toBeInTheDocument()
  })

  it('navigates to next page when clicking next', async () => {
    const user = userEvent.setup()
    const manyTeams = Array.from({ length: 12 }, (_, i) => ({
      id: i + 1,
      name: `团队 ${i + 1}`,
      code: `T${String(i + 1).padStart(2, '0')}`,
      description: '',
      pmId: 1,
      pmDisplayName: '张明',
      createdAt: '2026-03-01T00:00:00Z',
      updatedAt: '2026-03-01T00:00:00Z',
    }))
    server.use(
      http.get('/v1/teams', ({ request }) => {
        const url = new URL(request.url)
        const page = Number(url.searchParams.get('page') || 1)
        const pageSize = Number(url.searchParams.get('pageSize') || 10)
        const start = (page - 1) * pageSize
        return HttpResponse.json({
          code: 0,
          data: { items: manyTeams.slice(start, start + pageSize), total: manyTeams.length, page, pageSize },
        })
      }),
    )

    renderPage()
    await waitFor(() => {
      expect(screen.getByText('团队 1')).toBeInTheDocument()
    })

    // Go to page 2
    await user.click(screen.getByLabelText('Next page'))

    await waitFor(() => {
      expect(screen.getByText('团队 11')).toBeInTheDocument()
      expect(screen.getByText('团队 12')).toBeInTheDocument()
    })
    expect(screen.queryByText('团队 1')).not.toBeInTheDocument()
  })

  it('filters teams by search input', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('产品研发团队')).toBeInTheDocument()
    })

    const searchInput = screen.getByTestId('team-search-input')
    await user.type(searchInput, '设计')

    await waitFor(() => {
      expect(screen.getByText('设计团队')).toBeInTheDocument()
      expect(screen.queryByText('产品研发团队')).not.toBeInTheDocument()
      expect(screen.queryByText('基础架构团队')).not.toBeInTheDocument()
    })
  })
})
