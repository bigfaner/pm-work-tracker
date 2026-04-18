import { describe, it, expect, beforeEach, vi } from 'vitest'
import { render, screen, waitFor, within, fireEvent } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import TeamManagementPage from './TeamManagementPage'
import { useAuthStore } from '@/store/auth'
import { useTeamStore } from '@/store/team'
import type { User, Team, TeamDetailResp, TeamMemberResp } from '@/types'

// --- Mocks ---

const mockGetTeam = vi.fn()
const mockListMembers = vi.fn()
const mockInviteMember = vi.fn()
const mockRemoveMember = vi.fn()
const mockTransferPm = vi.fn()
const mockDeleteTeam = vi.fn()

vi.mock('@/api/teams', () => ({
  getTeamApi: (...args: unknown[]) => mockGetTeam(...args),
  listMembersApi: (...args: unknown[]) => mockListMembers(...args),
  inviteMemberApi: (...args: unknown[]) => mockInviteMember(...args),
  removeMemberApi: (...args: unknown[]) => mockRemoveMember(...args),
  transferPmApi: (...args: unknown[]) => mockTransferPm(...args),
  deleteTeamApi: (...args: unknown[]) => mockDeleteTeam(...args),
}))

// --- Test Data ---

const pmUser: User = {
  id: 1,
  username: 'pmuser',
  display_name: 'PM User',
  is_super_admin: false,
  can_create_team: false,
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
}

const memberUser: User = {
  id: 10,
  username: 'member1',
  display_name: 'Member One',
  is_super_admin: false,
  can_create_team: false,
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
}

const mockTeam: Team = {
  id: 1,
  name: 'Team Alpha',
  description: 'A test team',
  pm_id: 1,
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
}

const mockTeamDetail: TeamDetailResp = {
  id: 1,
  name: 'Team Alpha',
  description: 'A test team',
  pmId: 1,
  pm: { displayName: 'PM User' },
  memberCount: 2,
  mainItemCount: 5,
  createdAt: '2024-01-01T00:00:00Z',
}

const mockMembers: TeamMemberResp[] = [
  { userId: 1, displayName: 'PM User', username: 'pmuser', role: 'pm', joinedAt: '2024-01-01' },
  { userId: 10, displayName: 'Member One', username: 'member1', role: 'member', joinedAt: '2024-01-15' },
]

// --- Helpers ---

function createQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: { retry: false },
    },
  })
}

function renderPage(user: User = pmUser, teamId: number | null = 1) {
  useAuthStore.getState().setAuth('token', user)
  const qc = createQueryClient()
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter initialEntries={[`/teams/${teamId}/settings`]}>
        <Routes>
          <Route path="/teams/:teamId/settings" element={<TeamManagementPage />} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

// Helper to find a table row by a text it contains
function findTableRowByText(text: string): HTMLElement {
  const cells = screen.getAllByText(text)
  // Find the cell that is inside a td
  const td = cells.find((el) => el.closest('td'))
  if (!td) throw new Error(`No table cell found containing "${text}"`)
  return td.closest('tr')!
}

// --- Tests ---

describe('TeamManagementPage', () => {
  beforeEach(() => {
    useAuthStore.getState().clearAuth()
    useAuthStore.getState().setAuth('token', pmUser)
    useTeamStore.getState().setTeams([mockTeam])
    useTeamStore.getState().setCurrentTeam(1)

    mockGetTeam.mockResolvedValue(mockTeamDetail)
    mockListMembers.mockResolvedValue(mockMembers)
    mockInviteMember.mockReset()
    mockRemoveMember.mockReset()
    mockTransferPm.mockReset()
    mockDeleteTeam.mockReset()
  })

  // --- Basic rendering ---

  it('renders page with data-testid', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByTestId('team-management-page')).toBeInTheDocument()
    })
  })

  it('renders page title 团队管理', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('团队管理')).toBeInTheDocument()
    })
  })

  // --- Loading state ---

  it('shows skeleton loading state while fetching', () => {
    let resolvePromise!: (v: unknown) => void
    mockGetTeam.mockReturnValue(new Promise((resolve) => { resolvePromise = resolve }))
    renderPage()
    expect(screen.getByTestId('team-mgmt-skeleton')).toBeInTheDocument()
    resolvePromise(mockTeamDetail)
  })

  // --- Team info card ---

  it('renders team info card with team name, PM, member count, creation time', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByTestId('team-info-card')).toBeInTheDocument()
    })
    const infoCard = screen.getByTestId('team-info-card')
    expect(within(infoCard).getAllByText('Team Alpha').length).toBeGreaterThan(0)
    // memberCount is rendered as "2"
    expect(within(infoCard).getAllByText('2').length).toBeGreaterThan(0)
  })

  // --- Member table ---

  it('renders member table with correct columns', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('Member One')).toBeInTheDocument()
    })
    // Table headers
    expect(screen.getByText('姓名')).toBeInTheDocument()
    expect(screen.getByText('账号')).toBeInTheDocument()
    expect(screen.getByText('角色')).toBeInTheDocument()
    expect(screen.getByText('加入时间')).toBeInTheDocument()
    expect(screen.getByText('操作')).toBeInTheDocument()
  })

  it('renders PM row with PM tag and no actions', async () => {
    renderPage()
    await waitFor(() => {
      // PM tag should appear in the table (as a Tag component, not just a label)
      const pmTags = screen.getAllByText('PM')
      expect(pmTags.length).toBeGreaterThanOrEqual(1)
    })
    // No actions button for PM row (self)
    expect(screen.queryByTestId('member-actions-1')).not.toBeInTheDocument()
  })

  it('renders member row with member tag', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('Member One')).toBeInTheDocument()
    })
    expect(screen.getByText('成员')).toBeInTheDocument()
    expect(screen.getByText('member1')).toBeInTheDocument()
  })

  it('shows username in account column', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('pmuser')).toBeInTheDocument()
      expect(screen.getByText('member1')).toBeInTheDocument()
    })
  })

  it('shows join date in correct format', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('2024-01-15')).toBeInTheDocument()
    })
  })

  // --- PM-only buttons ---

  it('shows invite button for PM user', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByTestId('invite-member-btn')).toBeInTheDocument()
    })
  })

  it('hides invite button for non-PM member', async () => {
    renderPage(memberUser)
    await waitFor(() => {
      expect(screen.queryByTestId('invite-member-btn')).not.toBeInTheDocument()
    })
  })

  it('shows danger zone with dissolve button for PM', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByTestId('danger-zone')).toBeInTheDocument()
      expect(screen.getByTestId('dissolve-team-btn')).toBeInTheDocument()
    })
  })

  it('hides danger zone for non-PM member', async () => {
    renderPage(memberUser)
    await waitFor(() => {
      expect(screen.queryByTestId('danger-zone')).not.toBeInTheDocument()
    })
  })

  it('shows member actions dropdown for PM user on member rows', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByTestId('member-actions-10')).toBeInTheDocument()
    })
  })

  it('hides member actions dropdown for non-PM member', async () => {
    renderPage(memberUser)
    await waitFor(() => {
      expect(screen.queryByTestId('member-actions-10')).not.toBeInTheDocument()
    })
  })

  // --- Invite modal ---

  it('opens invite modal on button click', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByTestId('invite-member-btn')).toBeInTheDocument()
    })
    await user.click(screen.getByTestId('invite-member-btn'))
    expect(screen.getByTestId('invite-modal')).toBeInTheDocument()
  })

  it('invite modal has username input field', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByTestId('invite-member-btn')).toBeInTheDocument()
    })
    await user.click(screen.getByTestId('invite-member-btn'))
    expect(screen.getByTestId('invite-username')).toBeInTheDocument()
  })

  it('submits invite form and calls invite API', async () => {
    const user = userEvent.setup()
    mockInviteMember.mockResolvedValue(undefined)
    renderPage()
    await waitFor(() => {
      expect(screen.getByTestId('invite-member-btn')).toBeInTheDocument()
    })
    await user.click(screen.getByTestId('invite-member-btn'))

    await user.type(screen.getByTestId('invite-username'), 'newuser')
    await user.click(screen.getByTestId('invite-submit-btn'))

    await waitFor(() => {
      expect(mockInviteMember).toHaveBeenCalledWith(1, { username: 'newuser', role: 'member' })
    })
  })

  it('shows inline error for USER_NOT_FOUND', async () => {
    const user = userEvent.setup()
    const error = { response: { status: 422, data: { code: 'USER_NOT_FOUND', message: '用户不存在' } } }
    mockInviteMember.mockRejectedValue(error)
    renderPage()
    await waitFor(() => {
      expect(screen.getByTestId('invite-member-btn')).toBeInTheDocument()
    })
    await user.click(screen.getByTestId('invite-member-btn'))

    await user.type(screen.getByTestId('invite-username'), 'nonexistent')
    await user.click(screen.getByTestId('invite-submit-btn'))

    await waitFor(() => {
      expect(screen.getByText('用户不存在')).toBeInTheDocument()
    })
  })

  it('shows inline error for ALREADY_MEMBER', async () => {
    const user = userEvent.setup()
    const error = { response: { status: 422, data: { code: 'ALREADY_MEMBER', message: '该用户已是团队成员' } } }
    mockInviteMember.mockRejectedValue(error)
    renderPage()
    await waitFor(() => {
      expect(screen.getByTestId('invite-member-btn')).toBeInTheDocument()
    })
    await user.click(screen.getByTestId('invite-member-btn'))

    await user.type(screen.getByTestId('invite-username'), 'pmuser')
    await user.click(screen.getByTestId('invite-submit-btn'))

    await waitFor(() => {
      expect(screen.getByText('该用户已是团队成员')).toBeInTheDocument()
    })
  })

  // --- Remove member ---

  it('shows remove confirm and calls remove API', async () => {
    const user = userEvent.setup()
    mockRemoveMember.mockResolvedValue(undefined)
    renderPage()
    await waitFor(() => {
      expect(screen.getByTestId('member-actions-10')).toBeInTheDocument()
    })

    await user.click(screen.getByTestId('member-actions-10'))
    const removeBtn = await screen.findByText('移除成员')
    await user.click(removeBtn)

    // Confirm dialog should appear - find the one with "确认移除" content
    const confirmDialog = await screen.findByText('确认移除 Member One？')
    const dialog = confirmDialog.closest('.ant-modal')!
    expect(dialog).toBeInTheDocument()

    const okBtn = within(dialog as HTMLElement).getByRole('button', { name: 'OK' })
    await user.click(okBtn)

    await waitFor(() => {
      expect(mockRemoveMember).toHaveBeenCalledWith(1, 10)
    })
  })

  // --- Transfer PM ---

  it('opens transfer PM modal from dropdown', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByTestId('member-actions-10')).toBeInTheDocument()
    })

    await user.click(screen.getByTestId('member-actions-10'))
    const transferBtn = await screen.findByText('设为 PM')
    await user.click(transferBtn)

    // Confirm dialog should appear with transfer text
    const confirmText = await screen.findByText(/确认将 Member One 设为 PM/)
    expect(confirmText).toBeInTheDocument()
  })

  it('calls transfer PM API on confirm', async () => {
    const user = userEvent.setup()
    mockTransferPm.mockResolvedValue(undefined)
    renderPage()
    await waitFor(() => {
      expect(screen.getByTestId('member-actions-10')).toBeInTheDocument()
    })

    await user.click(screen.getByTestId('member-actions-10'))
    const transferBtn = await screen.findByText('设为 PM')
    await user.click(transferBtn)

    // Find all matching confirm texts and use the last (most recent) one
    const confirmTexts = await screen.findAllByText(/确认将 Member One 设为 PM/)
    const confirmText = confirmTexts[confirmTexts.length - 1]
    const dialog = confirmText.closest('.ant-modal')!

    const okBtn = within(dialog as HTMLElement).getByRole('button', { name: 'OK' })
    await user.click(okBtn)

    await waitFor(() => {
      expect(mockTransferPm).toHaveBeenCalledWith(1, { newPmUserId: 10 })
    })
  })

  // --- Dissolve team ---

  it('opens dissolve modal on button click', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByTestId('dissolve-team-btn')).toBeInTheDocument()
    })
    await user.click(screen.getByTestId('dissolve-team-btn'))
    expect(screen.getByTestId('dissolve-modal')).toBeInTheDocument()
  })

  it('dissolve modal has team name input and confirm button disabled', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByTestId('dissolve-team-btn')).toBeInTheDocument()
    })
    await user.click(screen.getByTestId('dissolve-team-btn'))

    expect(screen.getByTestId('dissolve-name-input')).toBeInTheDocument()
    // Confirm button should be disabled when input is empty
    expect(screen.getByTestId('dissolve-confirm-btn')).toBeDisabled()
  })

  it('enables confirm button when team name matches exactly', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByTestId('dissolve-team-btn')).toBeInTheDocument()
    })
    await user.click(screen.getByTestId('dissolve-team-btn'))

    const input = screen.getByTestId('dissolve-name-input')
    await user.type(input, 'Team Alpha')

    expect(screen.getByTestId('dissolve-confirm-btn')).not.toBeDisabled()
  })

  it('keeps confirm button disabled when team name does not match', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByTestId('dissolve-team-btn')).toBeInTheDocument()
    })
    await user.click(screen.getByTestId('dissolve-team-btn'))

    const input = screen.getByTestId('dissolve-name-input')
    await user.type(input, 'Wrong Name')

    expect(screen.getByTestId('dissolve-confirm-btn')).toBeDisabled()
  })

  it('calls delete team API on dissolve confirm', async () => {
    const user = userEvent.setup()
    mockDeleteTeam.mockResolvedValue(undefined)
    renderPage()
    await waitFor(() => {
      expect(screen.getByTestId('dissolve-team-btn')).toBeInTheDocument()
    })
    await user.click(screen.getByTestId('dissolve-team-btn'))

    const input = screen.getByTestId('dissolve-name-input')
    await user.type(input, 'Team Alpha')

    await user.click(screen.getByTestId('dissolve-confirm-btn'))

    await waitFor(() => {
      expect(mockDeleteTeam).toHaveBeenCalledWith(1, { confirmName: 'Team Alpha' })
    })
  })

  // --- Table refresh after actions ---

  it('refreshes member list after successful invite', async () => {
    const user = userEvent.setup()
    mockInviteMember.mockResolvedValue(undefined)
    const updatedMembers = [
      ...mockMembers,
      { userId: 20, displayName: 'New User', username: 'newuser', role: 'member', joinedAt: '2024-06-01' },
    ]
    mockListMembers.mockResolvedValue(updatedMembers)

    renderPage()
    await waitFor(() => {
      expect(screen.getByTestId('invite-member-btn')).toBeInTheDocument()
    })

    mockListMembers.mockClear()

    await user.click(screen.getByTestId('invite-member-btn'))
    await user.type(screen.getByTestId('invite-username'), 'newuser')
    await user.click(screen.getByTestId('invite-submit-btn'))

    await waitFor(() => {
      expect(mockListMembers).toHaveBeenCalled()
    })
  })

  it('refreshes member list after successful remove', async () => {
    const user = userEvent.setup()
    mockRemoveMember.mockResolvedValue(undefined)
    renderPage()
    await waitFor(() => {
      expect(screen.getByTestId('member-actions-10')).toBeInTheDocument()
    })

    mockListMembers.mockClear()

    await user.click(screen.getByTestId('member-actions-10'))
    const removeBtn = await screen.findByText('移除成员')
    await user.click(removeBtn)

    const confirmTexts = await screen.findAllByText('确认移除 Member One？')
    const confirmText = confirmTexts[confirmTexts.length - 1]
    const dialog = confirmText.closest('.ant-modal')!

    const okBtn = within(dialog as HTMLElement).getByRole('button', { name: 'OK' })
    await user.click(okBtn)

    await waitFor(() => {
      expect(mockListMembers).toHaveBeenCalled()
    })
  })
})
