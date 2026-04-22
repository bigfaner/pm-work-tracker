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
  { userId: 1, displayName: '张明', username: 'zhangming', role: 'pm', roleId: 1, roleName: 'pm', joinedAt: '2026-03-01' },
  { userId: 2, displayName: '李华', username: 'lihua', role: 'member', roleId: 3, roleName: 'member', joinedAt: '2026-03-05' },
  { userId: 3, displayName: '王芳', username: 'wangfang', role: 'member', roleId: 3, roleName: 'member', joinedAt: '2026-03-10' },
  { userId: 4, displayName: '赵强', username: 'zhaoqiang', role: 'member', roleId: 3, roleName: 'member', joinedAt: '2026-03-12' },
  { userId: 5, displayName: '陈静', username: 'chenjing', role: 'member', roleId: 3, roleName: 'member', joinedAt: '2026-03-15' },
]

const seedRoles = [
  { id: 1, name: 'superadmin', description: '超级管理员', isPreset: true, permissionCount: 30, memberCount: 1, createdAt: '2026-01-01T00:00:00Z' },
  { id: 2, name: 'pm', description: '团队管理权限', isPreset: true, permissionCount: 22, memberCount: 5, createdAt: '2026-01-01T00:00:00Z' },
  { id: 3, name: 'member', description: '普通成员', isPreset: true, permissionCount: 10, memberCount: 20, createdAt: '2026-01-01T00:00:00Z' },
  { id: 4, name: 'viewer', description: '只读查看者', isPreset: false, permissionCount: 3, memberCount: 0, createdAt: '2026-04-01T00:00:00Z' },
]

const seedAvailableUsers = [
  { userId: 10, displayName: '刘洋', username: 'liuyang' },
  { userId: 11, displayName: '周磊', username: 'zhoulei' },
]

function setupHandlers() {
  server.use(
    // Get team detail
    http.get('/v1/teams/:teamId', ({ params }) => {
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
    http.get('/v1/teams/:teamId/members', () => {
      return HttpResponse.json({ code: 0, data: seedMembers })
    }),

    // List roles
    http.get('/v1/admin/roles', () => {
      return HttpResponse.json({ code: 0, data: { items: seedRoles, total: 4, page: 1, pageSize: 100 } })
    }),

    // Transfer PM
    http.put('/v1/teams/:teamId/pm', async ({ request }) => {
      const body = (await request.json()) as { newPmUserId: number }
      const newPm = seedMembers.find((m) => m.userId === body.newPmUserId)
      return HttpResponse.json({ code: 0, data: { newPmName: newPm?.displayName } })
    }),

    // Remove member
    http.delete('/v1/teams/:teamId/members/:userId', () => {
      return HttpResponse.json({ code: 0, data: null })
    }),

    // Disband team
    http.delete('/v1/teams/:teamId', async ({ params, request }) => {
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
    http.get('/v1/teams/:teamId/available-users', () => {
      return HttpResponse.json({ code: 0, data: seedAvailableUsers })
    }),

    // Invite member (updated: expects roleId)
    http.post('/v1/teams/:teamId/members', async ({ request }) => {
      const body = (await request.json()) as { username: string; roleId: number }
      return HttpResponse.json({ code: 0, data: null })
    }),

    // Change member role
    http.put('/v1/teams/:teamId/members/:memberId/role', async ({ request }) => {
      const body = (await request.json()) as { roleId: number }
      return HttpResponse.json({ code: 0, data: null })
    }),
  )
}

describe('TeamDetailPage', () => {
  beforeEach(() => {
    setupHandlers()
    // Set user and permissions atomically to avoid race conditions
    useAuthStore.setState({
      isAuthenticated: true,
      isSuperAdmin: false,
      token: 'test-token',
      permissions: {
        isSuperAdmin: false,
        teamPermissions: { 1: ['team:invite', 'team:remove', 'team:transfer', 'team:delete'] },
      },
      permissionsLoadedAt: Date.now(),
    })
  })

  afterEach(() => {
    // Clean up auth store to prevent test pollution
    useAuthStore.setState({
      user: null,
      isAuthenticated: false,
      isSuperAdmin: false,
      token: null,
      permissions: null,
      permissionsLoadedAt: null,
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

    // Member badge (roleName field displayed)
    const memberBadges = screen.getAllByText('member')
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
      http.get('/v1/teams/:teamId', async () => {
        await new Promise((r) => setTimeout(r, 100))
        return HttpResponse.json({ code: 0, data: seedTeamDetail })
      }),
      http.get('/v1/teams/:teamId/members', async () => {
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

  // ========================================
  // NEW TESTS: Role-based invite & inline role change
  // ========================================

  // --- Invite dialog: role dropdown populated from API ---

  it('invite dialog shows role dropdown populated from roles API (excluding superadmin)', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('添加成员')).toBeInTheDocument()
    })

    await user.click(screen.getByText('添加成员'))

    await waitFor(() => {
      expect(screen.getByText('搜索用户')).toBeInTheDocument()
    })

    // Open the role select dropdown
    const roleSelect = screen.getByTestId('invite-role-select')
    await user.click(roleSelect)

    // Should show roles from API but NOT superadmin
    await waitFor(() => {
      // Use getAllByText since 'pm' and 'member' exist as badges in the member list
      // The select items have role="option" attribute
      const options = screen.getAllByRole('option')
      const optionTexts = options.map((o) => o.textContent)
      expect(optionTexts).toContain('pm')
      expect(optionTexts).toContain('member')
      expect(optionTexts).toContain('viewer')
      expect(optionTexts).not.toContain('superadmin')
    })
  })

  // --- Invite dialog: default role is member ---

  it('invite dialog defaults role to member', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('添加成员')).toBeInTheDocument()
    })

    await user.click(screen.getByText('添加成员'))

    await waitFor(() => {
      // The select should show "member" as default value
      const roleSelect = screen.getByTestId('invite-role-select')
      expect(roleSelect).toHaveTextContent('member')
    })
  })

  // --- Invite request sends roleId ---

  it('invite request sends roleId instead of role string', async () => {
    let capturedBody: any = null
    server.use(
      http.post('/v1/teams/:teamId/members', async ({ request }) => {
        capturedBody = await request.json()
        return HttpResponse.json({ code: 0, data: null })
      }),
    )

    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('添加成员')).toBeInTheDocument()
    })

    await user.click(screen.getByText('添加成员'))

    await waitFor(() => {
      expect(screen.getByText('搜索用户')).toBeInTheDocument()
    })

    // Type a username
    const input = screen.getByPlaceholderText('输入用户名或姓名搜索...')
    await user.type(input, 'newuser')

    // Submit
    const submitBtn = screen.getByTestId('invite-submit-btn')
    await user.click(submitBtn)

    await waitFor(() => {
      expect(capturedBody).not.toBeNull()
      expect(capturedBody.roleId).toBe(3) // member role id
      expect(capturedBody.role).toBeUndefined() // should NOT send role string
    })
  })

  // --- Member list: shows change role button ---

  it('member rows show change role button for non-PM members', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('成员列表')).toBeInTheDocument()
    })

    const changeButtons = screen.getAllByTestId('change-role-btn')
    // 4 non-PM members (but userId=1 is PM, so they don't get the button)
    // userId 2,3,4,5 are non-PM members
    expect(changeButtons.length).toBe(4)
  })

  // --- PM row: no change role button (PM is current user self) ---

  it('PM row does not show change role button', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('成员列表')).toBeInTheDocument()
    })

    // PM row (张明, userId=1)
    const rows = screen.getAllByRole('row')
    const pmRow = rows[1]
    expect(within(pmRow).queryByTestId('change-role-btn')).not.toBeInTheDocument()
  })

  // --- Inline role change: clicking change shows select ---

  it('clicking change role button shows inline select dropdown', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('成员列表')).toBeInTheDocument()
    })

    const changeButtons = screen.getAllByTestId('change-role-btn')
    await user.click(changeButtons[0])

    await waitFor(() => {
      expect(screen.getByTestId('inline-role-select')).toBeInTheDocument()
    })
  })

  // --- Inline role change: selecting new role auto-submits ---

  it('selecting a new role in inline select submits change role API', async () => {
    let capturedBody: any = null
    let capturedMemberId: string | null = null
    server.use(
      http.put('/v1/teams/:teamId/members/:memberId/role', async ({ params, request }) => {
        capturedMemberId = params.memberId as string
        capturedBody = await request.json()
        return HttpResponse.json({ code: 0, data: null })
      }),
    )

    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('成员列表')).toBeInTheDocument()
    })

    const changeButtons = screen.getAllByTestId('change-role-btn')
    // Click change on 李华 (userId=2)
    await user.click(changeButtons[0])

    await waitFor(() => {
      expect(screen.getByTestId('inline-role-select')).toBeInTheDocument()
    })

    // Open the select and pick a new role
    const inlineSelect = screen.getByTestId('inline-role-select')
    await user.click(inlineSelect)

    // Select "viewer" role
    await waitFor(() => {
      expect(screen.getByText('viewer')).toBeInTheDocument()
    })
    await user.click(screen.getByText('viewer'))

    await waitFor(() => {
      expect(capturedBody).not.toBeNull()
      expect(capturedBody.roleId).toBe(4) // viewer role id
      expect(capturedMemberId).toBe('2') // 李华's userId
    })
  })

  // --- Inline role change: success shows toast ---

  it('successful role change shows toast and reverts to text display', async () => {
    server.use(
      http.put('/v1/teams/:teamId/members/:memberId/role', async () => {
        return HttpResponse.json({ code: 0, data: null })
      }),
    )

    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('成员列表')).toBeInTheDocument()
    })

    const changeButtons = screen.getAllByTestId('change-role-btn')
    await user.click(changeButtons[0])

    await waitFor(() => {
      expect(screen.getByTestId('inline-role-select')).toBeInTheDocument()
    })

    const inlineSelect = screen.getByTestId('inline-role-select')
    await user.click(inlineSelect)

    await waitFor(() => {
      expect(screen.getByText('viewer')).toBeInTheDocument()
    })
    await user.click(screen.getByText('viewer'))

    await waitFor(() => {
      expect(screen.getByText('角色已更新')).toBeInTheDocument()
    })
  })

  // --- Inline role change: failure shows error toast and reverts ---

  it('failed role change shows error toast and reverts to original', async () => {
    server.use(
      http.put('/v1/teams/:teamId/members/:memberId/role', async () => {
        return HttpResponse.json(
          { code: 'ERR_FORBIDDEN', message: '权限不足' },
          { status: 403 },
        )
      }),
    )

    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('成员列表')).toBeInTheDocument()
    })

    const changeButtons = screen.getAllByTestId('change-role-btn')
    await user.click(changeButtons[0])

    await waitFor(() => {
      expect(screen.getByTestId('inline-role-select')).toBeInTheDocument()
    })

    const inlineSelect = screen.getByTestId('inline-role-select')
    await user.click(inlineSelect)

    await waitFor(() => {
      expect(screen.getByText('viewer')).toBeInTheDocument()
    })
    await user.click(screen.getByText('viewer'))

    await waitFor(() => {
      expect(screen.getByText('角色变更失败，请稍后重试')).toBeInTheDocument()
    })

    // Should revert to text mode (no inline select)
    expect(screen.queryByTestId('inline-role-select')).not.toBeInTheDocument()
  })

  // --- Superadmin excluded from role dropdown ---

  it('inline role select excludes superadmin', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('成员列表')).toBeInTheDocument()
    })

    // Wait for change-role buttons to appear (depends on roles API loading)
    await waitFor(() => {
      expect(screen.getAllByTestId('change-role-btn').length).toBeGreaterThan(0)
    })

    const changeButtons = screen.getAllByTestId('change-role-btn')
    await user.click(changeButtons[0])

    await waitFor(() => {
      expect(screen.getByTestId('inline-role-select')).toBeInTheDocument()
    })

    const inlineSelect = screen.getByTestId('inline-role-select')
    await user.click(inlineSelect)

    await waitFor(() => {
      // Use role="option" to select dropdown items specifically
      const options = screen.getAllByRole('option')
      const optionTexts = options.map((o) => o.textContent)
      expect(optionTexts).toContain('pm')
      expect(optionTexts).toContain('member')
      expect(optionTexts).toContain('viewer')
      expect(optionTexts).not.toContain('superadmin')
    })
  })
})
