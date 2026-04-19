import { describe, it, expect, vi, beforeEach, beforeAll, afterAll, afterEach } from 'vitest'
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { server } from '@/mocks/server'
import { http, HttpResponse } from 'msw'

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
        <TeamManagementPage />
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

// Lazy import to ensure MSW is ready
const { default: TeamManagementPage } = await import('./TeamManagementPage')

// --- Seed data ---

const seedTeams = [
  {
    id: 1,
    name: '产品研发团队',
    description: '负责核心产品的研发与迭代',
    pmId: 1,
    createdAt: '2026-03-01T00:00:00Z',
    updatedAt: '2026-03-01T00:00:00Z',
  },
  {
    id: 2,
    name: '设计团队',
    description: '负责 UI/UX 设计与交互体验优化',
    pmId: 2,
    createdAt: '2026-03-15T00:00:00Z',
    updatedAt: '2026-03-15T00:00:00Z',
  },
  {
    id: 3,
    name: '基础架构团队',
    description: '负责底层基础设施与运维体系建设',
    pmId: 3,
    createdAt: '2026-04-01T00:00:00Z',
    updatedAt: '2026-04-01T00:00:00Z',
  },
]

function setupHandlers() {
  server.use(
    // List teams
    http.get('/api/v1/teams', () => {
      return HttpResponse.json({ code: 0, data: seedTeams })
    }),

    // Create team
    http.post('/api/v1/teams', async ({ request }) => {
      const body = (await request.json()) as { name: string; description?: string }
      if (!body.name || !body.name.trim()) {
        return HttpResponse.json(
          { code: 'VALIDATION_ERROR', message: '团队名称不能为空' },
          { status: 422 },
        )
      }
      return HttpResponse.json({
        code: 0,
        data: {
          id: 100,
          name: body.name,
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

  // --- Empty state ---

  it('shows empty state when no teams exist', async () => {
    server.use(
      http.get('/api/v1/teams', () => {
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
      http.get('/api/v1/teams', async () => {
        await new Promise((resolve) => setTimeout(resolve, 500))
        return HttpResponse.json({ code: 0, data: seedTeams })
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
})
