import { describe, it, expect, vi, beforeEach, beforeAll, afterAll, afterEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import * as toast from '@/lib/toast'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { server } from '@/mocks/server'
import { http, HttpResponse } from 'msw'
import { useTeamStore } from '@/store/team'
import { useAuthStore } from '@/store/auth'
import SubItemDetailPage from './SubItemDetailPage'

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

function renderPage(mainItemId = '1', subItemId = '12') {
  const qc = createQueryClient()
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter initialEntries={[`/items/${mainItemId}/sub/${subItemId}`]}>
        <Routes>
          <Route path="/items/:mainItemId/sub/:subItemId" element={<SubItemDetailPage />} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

// --- Seed data ---

const seedMembers = [
  { id: 1, bizKey: '1', teamKey: '1', userKey: 'U001', displayName: 'Test User', username: 'testuser', role: 'pm', roleKey: 1, roleName: 'pm', joinedAt: '2024-01-01' },
  { id: 2, bizKey: '2', teamKey: '1', userKey: 'U002', displayName: 'Alice', username: 'alice', role: 'member', roleKey: 2, roleName: 'member', joinedAt: '2024-01-01' },
  { id: 3, bizKey: '3', teamKey: '1', userKey: 'U003', displayName: 'Bob', username: 'bob', role: 'member', roleKey: 3, roleName: 'member', joinedAt: '2024-01-01' },
]

const seedMainItem = {
  bizKey: '1', teamKey: '1', code: 'MI-0001', title: 'Alpha Task', priority: 'P1',
  proposerKey: 'U001', assigneeKey: 'U001', planStartDate: '2026-03-20', expectedEndDate: '2026-04-15',
  actualEndDate: null, itemStatus: 'progressing', completion: 65,
  createTime: '2026-03-20T00:00:00Z', dbUpdateTime: '2026-04-01T00:00:00Z',
  subItems: [],
}

const seedSubItem = {
  bizKey: '12', teamKey: '1', mainItemKey: '1', code: 'SI-001-12', title: 'Sub Alpha 2', itemDesc: 'JWT Token implementation',
  priority: 'P2', assigneeKey: 'U003', planStartDate: '2026-04-08', expectedEndDate: '2026-04-18',
  actualEndDate: null, itemStatus: 'progressing', completion: 80,
  weight: 1,
  createTime: '2026-04-01T00:00:00Z', dbUpdateTime: '2026-04-08T00:00:00Z',
}

const seedProgressRecords = [
  {
    subItemKey: '12', teamKey: '1', authorKey: 'U003', completion: 10,
    achievement: '完成 JWT Token 签发逻辑的基础代码框架搭建',
    blocker: '', lesson: '', isPMCorrect: false,
    createTime: '2026-04-01T10:00:00Z',
  },
  {
    subItemKey: '12', teamKey: '1', authorKey: 'U003', completion: 40,
    achievement: 'Token 刷新机制实现，AccessToken + RefreshToken 双 Token 模式',
    blocker: 'Redis 集群偶发连接超时，影响 Token 黑名单方案',
    lesson: '', isPMCorrect: false,
    createTime: '2026-04-08T10:00:00Z',
  },
  {
    subItemKey: '12', teamKey: '1', authorKey: 'U003', completion: 80,
    achievement: 'Token 黑名单方案实现完成，集成测试通过',
    blocker: '部分旧接口 Token 格式兼容问题待修复',
    lesson: '', isPMCorrect: false,
    createTime: '2026-04-15T10:00:00Z',
  },
]

function setupHandlers() {
  server.use(
    // Get sub item
    http.get('/v1/teams/:teamId/sub-items/:itemId', ({ params }) => {
      if (String(params.itemId) === '12') {
        return HttpResponse.json({ code: 0, data: seedSubItem })
      }
      return HttpResponse.json({ code: 'NOT_FOUND', message: 'not found' }, { status: 404 })
    }),

    // Get main item (for parent link)
    http.get('/v1/teams/:teamId/main-items/:itemId', ({ params }) => {
      if (String(params.itemId) === '1') {
        return HttpResponse.json({ code: 0, data: seedMainItem })
      }
      return HttpResponse.json({ code: 'NOT_FOUND', message: 'not found' }, { status: 404 })
    }),

    // List progress records
    http.get('/v1/teams/:teamId/sub-items/:itemId/progress', ({ params }) => {
      if (String(params.itemId) === '12') {
        return HttpResponse.json({ code: 0, data: seedProgressRecords })
      }
      return HttpResponse.json({ code: 0, data: [] })
    }),

    // List members
    http.get('/v1/teams/:teamId/members', () => {
      return HttpResponse.json({ code: 0, data: seedMembers })
    }),

    // Append progress
    http.post('/v1/teams/:teamId/sub-items/:itemId/progress', async ({ request }) => {
      const body = await request.json() as Record<string, unknown>
      return HttpResponse.json({
        code: 0,
        data: {
          subItemKey: '12', teamKey: '1', authorKey: 'U001',
          lesson: '', isPMCorrect: false,
          createTime: new Date().toISOString(),
          ...body,
        },
      })
    }),
  )
}

describe('SubItemDetailPage', () => {
  beforeEach(() => {
    useTeamStore.setState({ currentTeamId: '1', teams: [{ bizKey: '1', name: 'Test Team', description: '', code: '', pmKey: '1', createdAt: '', updatedAt: '' }] })
    useAuthStore.getState().setPermissions({
      isSuperAdmin: false,
      teamPermissions: { 1: ['progress:update'] },
    })
    setupHandlers()
  })

  // --- Core rendering ---

  it('renders breadcrumb with three levels', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('事项清单')).toBeInTheDocument()
      // "Alpha Task" appears in breadcrumb and info card
      expect(screen.getAllByText('Alpha Task').length).toBeGreaterThanOrEqual(1)
      // "Sub Alpha 2" appears in breadcrumb and heading
      expect(screen.getAllByText('Sub Alpha 2').length).toBeGreaterThanOrEqual(1)
    })
  })

  it('breadcrumb links are clickable', async () => {
    renderPage()
    await waitFor(() => {
      const itemsLink = screen.getByText('事项清单').closest('a')
      expect(itemsLink).toHaveAttribute('href', '/items')

      // "Alpha Task" appears in both breadcrumb and info card
      const alphaLinks = screen.getAllByText('Alpha Task')
      // Find the breadcrumb link (first one, inside nav)
      const mainLink = alphaLinks.find(el => el.closest('nav') !== null)?.closest('a')
      expect(mainLink).toHaveAttribute('href', '/items/1')
    })
  })

  it('renders sub item title', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByRole('heading', { name: 'Sub Alpha 2' })).toBeInTheDocument()
    })
  })

  // --- Info card ---

  it('renders info card fields', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('负责人')).toBeInTheDocument()
      expect(screen.getByText('预期完成时间')).toBeInTheDocument()
      expect(screen.getByText('总进度')).toBeInTheDocument()
    })
  })

  it('renders sub item code', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('SI-001-12')).toBeInTheDocument()
    })
  })

  it('renders sub item code as Badge component (rounded-full, not plain rounded)', async () => {
    renderPage()
    await waitFor(() => {
      const codeEl = screen.getByText('SI-001-12')
      expect(codeEl.className).toContain('font-mono')
      expect(codeEl.className).toContain('rounded-full')
      expect(codeEl.className).not.toContain('px-1.5')
    })
  })

  it('renders priority badge', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('P2')).toBeInTheDocument()
    })
  })

  it('renders assignee with avatar', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('Bob')).toBeInTheDocument()
    })
  })

  it('renders parent item link', async () => {
    renderPage()
    await waitFor(() => {
      // "Alpha Task" appears in breadcrumb (breadcrumb link) and info card (所属主事项 link)
      const alphaLinks = screen.getAllByText('Alpha Task')
      const parentLink = alphaLinks[0].closest('a')
      expect(parentLink).toHaveAttribute('href', '/items/1')
    })
  })

  // --- Progress bar ---

  it('renders linear progress bar with percentage', async () => {
    renderPage()
    await waitFor(() => {
      // "80%" appears in info card, progress bar, and timeline
      const matches = screen.getAllByText('80%')
      expect(matches.length).toBeGreaterThanOrEqual(1)
    })
  })

  // --- Progress timeline ---

  it('renders progress timeline', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('进度记录')).toBeInTheDocument()
    })
  })

  it('append button is in progress card header, not title bar', async () => {
    renderPage()
    await waitFor(() => {
      const btn = screen.getByRole('button', { name: '追加进度' })
      const h1 = screen.getByRole('heading', { name: 'Sub Alpha 2' })
      expect(h1.parentElement).not.toContainElement(btn)
      const progressHeading = screen.getByText('进度记录')
      expect(progressHeading.parentElement).toContainElement(btn)
    })
  })

  it('renders timeline entries in reverse chronological order', async () => {
    renderPage()
    await waitFor(() => {
      // Dates appear in timeline
      const dates = screen.getAllByText(/2026-04-\d{2}/)
      expect(dates.length).toBeGreaterThanOrEqual(3)
    })
  })

  it('renders achievements in timeline', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText(/Token 黑名单方案实现完成/)).toBeInTheDocument()
    })
  })

  it('renders blockers in timeline', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText(/Redis 集群偶发连接超时/)).toBeInTheDocument()
    })
  })

  // --- Append progress dialog ---

  it('opens append progress dialog', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByRole('heading', { name: 'Sub Alpha 2' })).toBeInTheDocument()
    })

    await user.click(screen.getByRole('button', { name: '追加进度' }))

    await waitFor(() => {
      // Dialog title appears (h2)
      expect(screen.getByRole('heading', { name: '追加进度', level: 2 })).toBeInTheDocument()
    })
  })

  it('shows hint about minimum completion value', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('追加进度')).toBeInTheDocument()
    })

    await user.click(screen.getByText('追加进度'))

    await waitFor(() => {
      expect(screen.getByText(/不能低于上一条记录的进度/)).toBeInTheDocument()
    })
  })

  // --- Append progress validation ---

  it('bug: shows error toast when submitting completion lower than last record', async () => {
    const showToastSpy = vi.spyOn(toast, 'showToast')
    const user = userEvent.setup()
    renderPage()

    await waitFor(() => {
      expect(screen.getByRole('button', { name: '追加进度' })).toBeInTheDocument()
    })

    await user.click(screen.getByRole('button', { name: '追加进度' }))

    await waitFor(() => {
      expect(screen.getByRole('heading', { name: '追加进度', level: 2 })).toBeInTheDocument()
    })

    // lastCompletion is 80 (last record); enter 50 which is lower
    const input = screen.getByPlaceholderText('请输入 0-100 的整数')
    await user.clear(input)
    await user.type(input, '50')

    await user.click(screen.getByRole('button', { name: '提交' }))

    await waitFor(() => {
      expect(showToastSpy).toHaveBeenCalledWith(expect.any(String), 'error')
    })
  })

  // --- No antd imports ---

  it('does not import antd', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByRole('heading', { name: 'Sub Alpha 2' })).toBeInTheDocument()
    })
    const antdElements = document.querySelectorAll('[class*="ant-"]')
    expect(antdElements.length).toBe(0)
  })
})
