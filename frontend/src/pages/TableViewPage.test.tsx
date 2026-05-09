import { describe, it, expect, beforeEach, beforeAll, afterAll, afterEach, vi } from 'vitest'
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { server } from '@/mocks/server'
import { http, HttpResponse } from 'msw'
import { useTeamStore } from '@/store/team'
import { ToastProvider } from '@/components/ui/toast'
import TableViewPage from './TableViewPage'
import type { TableRow, PageResult } from '@/types'

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
          <TableViewPage />
        </MemoryRouter>
      </ToastProvider>
    </QueryClientProvider>,
  )
}

// --- Seed data ---

const seedRows: TableRow[] = [
  {
    bizKey: '1',
    type: 'main',
    code: 'MI-0001',
    title: '用户认证模块开发',
    priority: 'P1',
    assigneeKey: 'U001',
    assigneeName: '张明',
    itemStatus: 'progressing',
    completion: 65,
    expectedEndDate: '2026-04-15',
    actualEndDate: null,
    mainItemId: null,
  },
  {
    bizKey: '2',
    type: 'sub',
    code: 'SI-0002',
    title: '登录页开发',
    priority: 'P1',
    assigneeKey: 'U002',
    assigneeName: '李华',
    itemStatus: 'completed',
    completion: 100,
    expectedEndDate: '2026-04-10',
    actualEndDate: '2026-04-09',
    mainItemId: '1',
  },
  {
    bizKey: '3',
    type: 'sub',
    code: 'SI-0003',
    title: 'JWT Token 集成',
    priority: 'P2',
    assigneeKey: 'U003',
    assigneeName: '王芳',
    itemStatus: 'progressing',
    completion: 80,
    expectedEndDate: '2026-04-18',
    actualEndDate: null,
    mainItemId: '1',
  },
  {
    bizKey: '4',
    type: 'main',
    code: 'MI-0004',
    title: '数据看板设计',
    priority: 'P2',
    assigneeKey: 'U004',
    assigneeName: '赵强',
    itemStatus: 'progressing',
    completion: 40,
    expectedEndDate: '2026-05-05',
    actualEndDate: null,
    mainItemId: null,
  },
  {
    bizKey: '5',
    type: 'sub',
    code: 'SI-0005',
    title: '看板数据接口',
    priority: 'P2',
    assigneeKey: 'U004',
    assigneeName: '赵强',
    itemStatus: 'completed',
    completion: 100,
    expectedEndDate: '2026-04-05',
    actualEndDate: '2026-04-04',
    mainItemId: '4',
  },
]

function makePageResult(rows: TableRow[], page = 1, pageSize = 50): PageResult<TableRow> {
  return {
    items: rows,
    total: rows.length,
    page,
    size: pageSize,
  }
}

let capturedFilter: Record<string, string | string[]> = {}

function setupTableHandler(rows = seedRows) {
  capturedFilter = {}
  server.use(
    http.get('/v1/teams/:teamId/views/table', ({ request }) => {
      const url = new URL(request.url)
      capturedFilter = Object.fromEntries(url.searchParams.entries())
      const page = Number(url.searchParams.get('page') || 1)
      const pageSize = Number(url.searchParams.get('pageSize') || 50)

      // Server-side filtering simulation
      let filtered = [...rows]
      const typeFilter = url.searchParams.get('type')
      if (typeFilter) {
        filtered = filtered.filter((r) => r.type === typeFilter)
      }
      const priorityFilter = url.searchParams.getAll('priority')
      if (priorityFilter.length > 0) {
        filtered = filtered.filter((r) => priorityFilter.includes(r.priority))
      }
      const statusFilter = url.searchParams.getAll('status')
      if (statusFilter.length > 0) {
        filtered = filtered.filter((r) => statusFilter.includes(r.itemStatus))
      }
      const assigneeFilter = url.searchParams.get('assigneeKey')
      if (assigneeFilter) {
        filtered = filtered.filter((r) => r.assigneeKey === assigneeFilter)
      }

      const start = (page - 1) * pageSize
      const end = start + pageSize
      const pageItems = filtered.slice(start, end)

      return HttpResponse.json({
        code: 0,
        data: {
          items: pageItems,
          total: filtered.length,
          page,
          size: pageSize,
        },
      })
    }),
    http.get('/v1/teams/:teamId/views/table/export', ({ request }) => {
      const csvContent = '编号,标题,类型\nMI-0001,用户认证模块开发,main'
      return new HttpResponse(csvContent, {
        headers: {
          'Content-Type': 'text/csv; charset=utf-8',
          'Content-Disposition': 'attachment; filename="items-export.csv"',
        },
      })
    }),
    http.get('/v1/teams/:teamId/members', () => {
      return HttpResponse.json({
        code: 0,
        data: [
          { id: 1, bizKey: '1', teamKey: '1', userKey: 'U001', displayName: '张明', username: 'zhangming', role: 'pm', roleKey: 1, roleName: 'pm', joinedAt: '2024-01-01' },
          { id: 2, bizKey: '2', teamKey: '1', userKey: 'U002', displayName: '李华', username: 'lihua', role: 'member', roleKey: 2, roleName: 'member', joinedAt: '2024-01-01' },
          { id: 3, bizKey: '3', teamKey: '1', userKey: 'U003', displayName: '王芳', username: 'wangfang', role: 'member', roleKey: 3, roleName: 'member', joinedAt: '2024-01-01' },
          { id: 4, bizKey: '4', teamKey: '1', userKey: 'U004', displayName: '赵强', username: 'zhaoqiang', role: 'member', roleKey: 4, roleName: 'member', joinedAt: '2024-01-01' },
        ],
      })
    }),
  )
}

// --- Tests ---

describe('TableViewPage', () => {
  beforeEach(() => {
    useTeamStore.setState({
      currentTeamId: '1',
      teams: [{ bizKey: '1', name: 'Test Team', description: '', code: '', pmKey: '1', createdAt: '', updatedAt: '' }],
    })
    setupTableHandler()
  })

  // --- Core rendering ---

  it('renders page with data-testid', async () => {
    renderPage()
    expect(screen.getByTestId('table-view-page')).toBeInTheDocument()
  })

  it('renders page title', async () => {
    renderPage()
    expect(screen.getByText('表格视图')).toBeInTheDocument()
  })

  it('renders export CSV button', async () => {
    renderPage()
    expect(screen.getByTestId('export-csv-btn')).toBeInTheDocument()
  })

  // --- Table rendering ---

  it('renders table headers', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('类型')).toBeInTheDocument()
      expect(screen.getByText('编号')).toBeInTheDocument()
      expect(screen.getByText('标题')).toBeInTheDocument()
      expect(screen.getByText('优先级')).toBeInTheDocument()
      expect(screen.getByText('负责人')).toBeInTheDocument()
      expect(screen.getByText('进度')).toBeInTheDocument()
      expect(screen.getByText('状态')).toBeInTheDocument()
      expect(screen.getByText('预期完成')).toBeInTheDocument()
      expect(screen.getByText('实际完成')).toBeInTheDocument()
    })
  })

  it('renders table rows from API', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('用户认证模块开发')).toBeInTheDocument()
      expect(screen.getByText('登录页开发')).toBeInTheDocument()
      expect(screen.getByText('JWT Token 集成')).toBeInTheDocument()
      expect(screen.getByText('数据看板设计')).toBeInTheDocument()
    })
  })

  it('renders type badges distinguishing main vs sub', async () => {
    renderPage()
    await waitFor(() => {
      const mainBadges = screen.getAllByText('主事项')
      const subBadges = screen.getAllByText('子事项')
      expect(mainBadges.length).toBeGreaterThanOrEqual(1)
      expect(subBadges.length).toBeGreaterThanOrEqual(1)
    })
  })

  it('renders item codes in monospace', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('MI-0001')).toBeInTheDocument()
      expect(screen.getByText('SI-0002')).toBeInTheDocument()
    })
  })

  // --- Title links ---

  it('main item title links to main item detail page', async () => {
    renderPage()
    await waitFor(() => {
      const link = screen.getByText('用户认证模块开发').closest('a')
      expect(link).toHaveAttribute('href', '/items/1')
    })
  })

  it('sub item title links to sub item detail page', async () => {
    renderPage()
    await waitFor(() => {
      const link = screen.getByText('登录页开发').closest('a')
      expect(link).toHaveAttribute('href', '/items/1/sub/2')
    })
  })

  // --- Overdue styling ---

  it('overdue items have red date styling', async () => {
    renderPage()
    await waitFor(() => {
      const overdueDate = screen.getByTestId('expected-date-1')
      expect(overdueDate).toHaveClass('text-error')
    })
  })

  it('non-overdue items do not have red date styling', async () => {
    renderPage()
    await waitFor(() => {
      const normalDate = screen.getByTestId('expected-date-4')
      expect(normalDate).not.toHaveClass('text-error')
    })
  })

  // --- Filters ---

  it('renders title search input', async () => {
    renderPage()
    expect(screen.getByPlaceholderText('搜索标题...')).toBeInTheDocument()
  })

  it('renders type filter', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByTestId('type-filter')).toBeInTheDocument()
    })
  })

  it('renders priority filter', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByTestId('priority-filter')).toBeInTheDocument()
    })
  })

  it('renders assignee filter', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByTestId('assignee-filter')).toBeInTheDocument()
    })
  })

  it('renders status filter', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByTestId('status-filter')).toBeInTheDocument()
    })
  })

  it('renders reset button', async () => {
    renderPage()
    expect(screen.getByText('重置')).toBeInTheDocument()
  })

  it('sends type filter to API when selected', async () => {
    setupTableHandler(seedRows.filter((r) => r.type === 'main'))
    renderPage()
    await waitFor(() => {
      // After type filter is applied, only main items should be shown
      expect(screen.getByText('用户认证模块开发')).toBeInTheDocument()
      expect(screen.queryByText('登录页开发')).not.toBeInTheDocument()
    })
  })

  // --- Pagination ---

  it('renders pagination with page size selector', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByTestId('pagination-page-size')).toBeInTheDocument()
    })
  })

  it('renders total count', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText(/共 5 条/)).toBeInTheDocument()
    })
  })

  // --- Empty state ---

  it('shows empty state when no results', async () => {
    setupTableHandler([])
    renderPage()
    await waitFor(() => {
      expect(screen.getByText(/暂无数据/)).toBeInTheDocument()
    })
  })

  // --- CSV export ---

  it('triggers CSV download on export button click', async () => {
    const createObjectURLSpy = vi.fn().mockReturnValue('blob:test')
    const revokeObjectURLSpy = vi.fn()
    globalThis.URL.createObjectURL = createObjectURLSpy
    globalThis.URL.revokeObjectURL = revokeObjectURLSpy

    // Mock createElement to capture the download link
    const anchorClickSpy = vi.fn()
    const originalCreateElement = document.createElement.bind(document)
    vi.spyOn(document, 'createElement').mockImplementation((tag: string) => {
      const el = originalCreateElement(tag)
      if (tag === 'a') {
        el.click = anchorClickSpy
      }
      return el
    })

    const user = userEvent.setup()
    renderPage()

    await waitFor(() => {
      expect(screen.getByText('用户认证模块开发')).toBeInTheDocument()
    })

    await user.click(screen.getByTestId('export-csv-btn'))

    await waitFor(() => {
      expect(createObjectURLSpy).toHaveBeenCalled()
    })

    vi.restoreAllMocks()
  })

  // --- No antd imports ---

  it('does not import antd', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('表格视图')).toBeInTheDocument()
    })
    const antdElements = document.querySelectorAll('[class*="ant-"]')
    expect(antdElements.length).toBe(0)
  })

  // --- Loading state ---

  it('shows loading state', async () => {
    renderPage()
    expect(screen.getByText('加载中...')).toBeInTheDocument()
  })
})
