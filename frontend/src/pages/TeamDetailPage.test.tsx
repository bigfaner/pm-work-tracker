import { describe, it, expect, vi, beforeEach, beforeAll, afterAll, afterEach } from 'vitest'
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { server } from '@/mocks/server'
import { http, HttpResponse } from 'msw'
import { ToastProvider } from '@/components/ui/toast'
import TeamDetailPage from './TeamDetailPage'
import { useAuthStore } from '@/store/auth'

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

function renderPage(teamId = '1') {
  const qc = createQueryClient()
  return render(
    <QueryClientProvider client={qc}>
      <ToastProvider>
        <MemoryRouter initialEntries={[`/teams/${teamId}`]}>
          <Routes>
            <Route path="/teams/:teamId" element={<TeamDetailPage />} />
            <Route path="/teams" element={<div data-testid="teams-page">Teams</div>} />
          </Routes>
        </MemoryRouter>
      </ToastProvider>
    </QueryClientProvider>,
  )
}

// --- Seed data ---

const seedTeamDetail = {
  id: 1,
  name: '产品研发团队',
  description: '负责核心产品的研发与迭代',
  pmId: 1,
  pmDisplayName: '张明',
  memberCount: 5,
  mainItemCount: 3,
  createdAt: '2026-03-01T00:00:00Z',
  updatedAt: '2026-03-01T00:00:00Z',
}

const seedMembers = [
  { userId: 1, displayName: '张明', username: 'zhangming', role: 'pm', joinedAt: '2026-03-01' },
  { userId: 2, displayName: '李华', username: 'lihua', role: 'member', joinedAt: '2026-03-05' },
  { userId: 3, displayName: '王芳', username: 'wangfang', role: 'member', joinedAt: '2026-03-10' },
  { userId: 4, displayName: '赵强', username: 'zhaoqiang', role: 'member', joinedAt: '2026-03-12' },
  { userId: 5, displayName: '陈静', username: 'chenjing', role: 'member', joinedAt: '2026-03-15' },
]

const seedAvailableUsers = [
  { userId: 10, displayName: '刘洋', username: 'liuyang' },
  { userId: 11, displayName: '周磊', username: 'zhoulei' },
]

function setupHandlers() {
  server.use(
    // Get team detail
    http.get('/api/v1/teams/:teamId', ({ params }) => {
      const teamId = Number(params.teamId)
      if (teamId === 999) {
        return HttpResponse.json(
          { code: 'NOT_FOUND', message: '团队不存在' },
          { status: 404 },
        )
      }
      return HttpResponse.json({
        code: 0,
        data: { ...seedTeamDetail, id: teamId },
      })
    }),

    // List members
    http.get('/api/v1/teams/:teamId/members', () => {
      return HttpResponse.json({ code: 0, data: seedMembers })
    }),

    // Transfer PM
    http.put('/api/v1/teams/:teamId/pm', async ({ request }) => {
      const body = (await request.json()) as { newPmUserId: number }
      const newPm = seedMembers.find((m) => m.userId === body.newPmUserId)
      return HttpResponse.json({ code: 0, data: { newPmName: newPm?.displayName } })
    }),

    // Remove member
    http.delete('/api/v1/teams/:teamId/members/:userId', () => {
      return HttpResponse.json({ code: 0, data: null })
    }),

    // Disband team
    http.delete('/api/v1/teams/:teamId', async ({ params, request }) => {
      const body = (await request.json()) as { confirmName: string }
      if (body.confirmName !== '产品研发团队') {
        return HttpResponse.json(
          { code: 'NAME_MISMATCH', message: '团队名称不匹配' },
          { status: 422 },
        )
      }
      return HttpResponse.json({ code: 0, data: null })
    }),

    // Invite member - search available users
    http.get('/api/v1/teams/:teamId/available-users', () => {
      return HttpResponse.json({ code: 0, data: seedAvailableUsers })
    }),

    // Invite member
    http.post('/api/v1/teams/:teamId/members', async ({ request }) => {
      const body = (await request.json()) as { username: string; role: string }
      return HttpResponse.json({ code: 0, data: null })
    }),
  )
}

describe('TeamDetailPage', () => {
  beforeEach(() => {
    setupHandlers()
    useAuthStore.getState().setPermissions({
      isSuperadmin: false,
      teamPermissions: { 1: ['team:invite', 'team:remove', 'team:transfer', 'team:delete'] },
    })
  })

  // --- Breadcrumb navigation ---

  it('renders breadcrumb with team management link', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('团队管理')).toBeInTheDocument()
    })
    const link = screen.getByText('团队管理').closest('a')
    expect(link).toHaveAttribute('href', '/teams')
  })

  it('renders breadcrumb with current team name', async () => {
    renderPage()
    await waitFor(() => {
      const instances = screen.getAllByText('产品研发团队')
      expect(instances.length).toBeGreaterThanOrEqual(1)
    })
  })

  // --- Team info card ---

  it('displays team info card with all fields', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getAllByText('产品研发团队').length).toBeGreaterThanOrEqual(1)
    })

    expect(screen.getAllByText('张明').length).toBeGreaterThanOrEqual(1)
    expect(screen.getByText('5')).toBeInTheDocument()
    expect(screen.getAllByText(/2026/).length).toBeGreaterThanOrEqual(1)
    expect(screen.getByText('负责核心产品的研发与迭代')).toBeInTheDocument()
  })

  // --- Member list ---

  it('displays member list with role badges', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('成员列表')).toBeInTheDocument()
    })

    // PM badge
    const pmBadges = screen.getAllByText('PM')
    expect(pmBadges.length).toBeGreaterThanOrEqual(1)

    // Member badge
    const memberBadges = screen.getAllByText('成员')
    expect(memberBadges.length).toBe(4)
  })

  // --- PM row has no action buttons ---

  it('PM row has no action buttons', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('成员列表')).toBeInTheDocument()
    })

    // Find the row containing the PM user 张明
    const rows = screen.getAllByRole('row')
    // First row is header, second row should be PM (张明)
    const pmRow = rows[1]
    expect(within(pmRow).getByText('张明')).toBeInTheDocument()
    expect(within(pmRow).queryByText('设为PM')).not.toBeInTheDocument()
    expect(within(pmRow).queryByText('移除')).not.toBeInTheDocument()
  })

  // --- Member row has action buttons ---

  it('member rows have set PM and remove buttons', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('成员列表')).toBeInTheDocument()
    })

    const setPmButtons = screen.getAllByText('设为PM')
    const removeButtons = screen.getAllByText('移除')
    expect(setPmButtons.length).toBe(4)
    expect(removeButtons.length).toBe(4)
  })

  // --- Transfer PM ---

  it('transfer PM shows confirmation dialog', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('李华')).toBeInTheDocument()
    })

    // Click "设为PM" for 李华
    const setPmButtons = screen.getAllByText('设为PM')
    await user.click(setPmButtons[0])

    await waitFor(() => {
      expect(screen.getByText(/确认将.*设为PM/)).toBeInTheDocument()
    })
  })

  // --- Remove member ---

  it('remove member shows confirmation dialog', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getAllByText('李华').length).toBeGreaterThanOrEqual(1)
    })

    const removeButtons = screen.getAllByText('移除')
    await user.click(removeButtons[0])

    await waitFor(() => {
      expect(screen.getAllByText(/确认移除/).length).toBeGreaterThanOrEqual(1)
    })
  })

  // --- Add member dialog ---

  it('opens add member dialog when clicking add button', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('添加成员')).toBeInTheDocument()
    })

    await user.click(screen.getByText('添加成员'))

    await waitFor(() => {
      expect(screen.getByText('搜索用户')).toBeInTheDocument()
    })
  })

  // --- Danger zone ---

  it('renders danger zone with disband button', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('危险操作')).toBeInTheDocument()
    })
    expect(screen.getAllByText('解散团队').length).toBeGreaterThanOrEqual(1)
  })

  // --- Disband team with name verification ---

  it('disband dialog requires exact team name match', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('危险操作')).toBeInTheDocument()
    })

    // Click disband button in danger zone
    const disbandButtons = screen.getAllByText('解散团队')
    await user.click(disbandButtons[disbandButtons.length - 1])

    await waitFor(() => {
      expect(screen.getByText(/此操作不可恢复/)).toBeInTheDocument()
    })

    // Confirm button should be disabled initially
    const confirmButtons = screen.getAllByText('解散团队')
    const dialogConfirm = confirmButtons[confirmButtons.length - 1]
    expect(dialogConfirm).toBeDisabled()

    // Type wrong name
    const input = screen.getByPlaceholderText(/产品研发团队/)
    await user.type(input, '错误名称')
    expect(dialogConfirm).toBeDisabled()

    // Clear and type correct name
    await user.clear(input)
    await user.type(input, '产品研发团队')

    await waitFor(() => {
      expect(dialogConfirm).not.toBeDisabled()
    })
  })

  // --- Loading state ---

  it('shows loading state', async () => {
    // Delay response to catch loading
    server.use(
      http.get('/api/v1/teams/:teamId', async () => {
        await new Promise((r) => setTimeout(r, 100))
        return HttpResponse.json({ code: 0, data: seedTeamDetail })
      }),
      http.get('/api/v1/teams/:teamId/members', async () => {
        await new Promise((r) => setTimeout(r, 100))
        return HttpResponse.json({ code: 0, data: seedMembers })
      }),
    )

    renderPage()
    expect(screen.getByText('加载中...')).toBeInTheDocument()
    await waitFor(() => {
      expect(screen.queryByText('加载中...')).not.toBeInTheDocument()
    })
  })

  // --- No antd imports ---

  it('does not import antd', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getAllByText('产品研发团队').length).toBeGreaterThanOrEqual(1)
    })
    const antdElements = document.querySelectorAll('[class*="ant-"]')
    expect(antdElements.length).toBe(0)
  })

  // --- Member search/filter ---

  it('has member search filter', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('成员列表')).toBeInTheDocument()
    })
    expect(screen.getByPlaceholderText('搜索姓名...')).toBeInTheDocument()
  })

  // --- Member list shows join dates ---

  it('shows join dates for members', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('成员列表')).toBeInTheDocument()
    })
    expect(screen.getByText('2026/03/05')).toBeInTheDocument()
    expect(screen.getByText('2026/03/10')).toBeInTheDocument()
  })
})
