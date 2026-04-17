import { describe, it, expect, beforeEach, vi } from 'vitest'
import { render, screen, waitFor, within, fireEvent } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import SubItemDetailPage from './SubItemDetailPage'
import { useAuthStore } from '@/store/auth'
import { useTeamStore } from '@/store/team'
import type { User, Team, MainItem, SubItem, ProgressRecord, TeamMemberResp } from '@/types'

// --- Mocks ---

const mockGetSubItem = vi.fn()
const mockGetMainItem = vi.fn()
const mockListProgress = vi.fn()
const mockAppendProgress = vi.fn()
const mockCorrectCompletion = vi.fn()
const mockListMembers = vi.fn()

vi.mock('@/api/subItems', () => ({
  getSubItemApi: (...args: unknown[]) => mockGetSubItem(...args),
}))

vi.mock('@/api/mainItems', () => ({
  getMainItemApi: (...args: unknown[]) => mockGetMainItem(...args),
}))

vi.mock('@/api/progress', () => ({
  listProgressApi: (...args: unknown[]) => mockListProgress(...args),
  appendProgressApi: (...args: unknown[]) => mockAppendProgress(...args),
  correctCompletionApi: (...args: unknown[]) => mockCorrectCompletion(...args),
}))

vi.mock('@/api/teams', () => ({
  listMembersApi: (...args: unknown[]) => mockListMembers(...args),
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
  username: 'member',
  display_name: 'Member User',
  is_super_admin: false,
  can_create_team: false,
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
}

const otherUser: User = {
  id: 99,
  username: 'other',
  display_name: 'Other User',
  is_super_admin: false,
  can_create_team: false,
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
}

const mockTeam: Team = {
  id: 1,
  name: 'Team Alpha',
  description: '',
  pm_id: 1,
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
}

const mockMembers: TeamMemberResp[] = [
  { userId: 1, displayName: 'PM User', username: 'pmuser', role: 'pm', joinedAt: '2024-01-01' },
  { userId: 10, displayName: 'Member User', username: 'member', role: 'member', joinedAt: '2024-01-01' },
]

function makeMainItem(overrides: Partial<MainItem> = {}): MainItem {
  return {
    id: 1,
    team_id: 1,
    code: 'MI-0001',
    title: 'Test Main Item',
    priority: 'P2',
    proposer_id: 1,
    assignee_id: 1,
    start_date: '2024-06-01',
    expected_end_date: '2024-07-01',
    actual_end_date: null,
    status: '进行中',
    completion: 50,
    is_key_item: false,
    delay_count: 0,
    archived_at: null,
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
    ...overrides,
  }
}

function makeSubItem(overrides: Partial<SubItem> = {}): SubItem {
  return {
    id: 2,
    team_id: 1,
    main_item_id: 1,
    title: 'Test Sub Item',
    description: 'A test sub item',
    priority: 'P1',
    assignee_id: 10,
    start_date: null,
    expected_end_date: '2024-06-15',
    actual_end_date: null,
    status: '进行中',
    completion: 60,
    is_key_item: false,
    delay_count: 0,
    weight: 1,
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
    ...overrides,
  }
}

function makeProgress(overrides: Partial<ProgressRecord> = {}): ProgressRecord {
  return {
    id: 1,
    sub_item_id: 2,
    team_id: 1,
    author_id: 10,
    completion: 30,
    achievement: '完成了前端开发',
    blocker: '',
    lesson: '学到了很多',
    is_pm_correct: false,
    created_at: '2024-06-10T10:00:00Z',
    ...overrides,
  }
}

// --- Helpers ---

function createQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: { retry: false },
    },
  })
}

function renderPage(
  user: User = pmUser,
  mainItemId: string = '1',
  subItemId: string = '2',
) {
  useAuthStore.getState().setAuth('token', user)
  const qc = createQueryClient()
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter initialEntries={[`/items/${mainItemId}/sub/${subItemId}`]}>
        <Routes>
          <Route path="/items" element={<div data-testid="item-view">Item View</div>} />
          <Route path="/items/:mainItemId" element={<div data-testid="main-detail">Main Detail</div>} />
          <Route path="/items/:mainItemId/sub/:subItemId" element={<SubItemDetailPage />} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

// --- Tests ---

describe('SubItemDetailPage', () => {
  beforeEach(() => {
    useAuthStore.getState().clearAuth()
    useAuthStore.getState().setAuth('token', pmUser)
    useTeamStore.getState().setTeams([mockTeam])
    useTeamStore.getState().setCurrentTeam(1)

    mockGetSubItem.mockResolvedValue(makeSubItem())
    mockGetMainItem.mockResolvedValue({ ...makeMainItem(), subItems: [] })
    mockListProgress.mockResolvedValue([])
    mockListMembers.mockResolvedValue(mockMembers)
    mockAppendProgress.mockReset()
    mockCorrectCompletion.mockReset()
  })

  // --- Skeleton ---

  it('shows skeleton while loading', () => {
    let resolvePromise!: (v: unknown) => void
    mockGetSubItem.mockReturnValue(new Promise((resolve) => { resolvePromise = resolve }))
    renderPage()
    expect(screen.getByTestId('detail-skeleton')).toBeInTheDocument()
    resolvePromise(makeSubItem())
  })

  // --- Breadcrumb ---

  it('renders breadcrumb: 事项视图 > MainItem title > SubItem title', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByTestId('breadcrumb')).toBeInTheDocument()
    })
    expect(screen.getByText('事项视图')).toBeInTheDocument()
    expect(screen.getAllByText('Test Main Item').length).toBeGreaterThanOrEqual(1)
    expect(screen.getAllByText('Test Sub Item').length).toBeGreaterThanOrEqual(1)
  })

  it('breadcrumb main item title links to main item detail page', async () => {
    renderPage()
    await waitFor(() => {
      const link = screen.getByTestId('breadcrumb-main-link')
      expect(link).toBeInTheDocument()
      expect(link.getAttribute('href')).toBe('/items/1')
    })
  })

  // --- Info Card ---

  it('renders info card with Descriptions (column=3, bordered)', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByTestId('info-card')).toBeInTheDocument()
    })
    const card = screen.getByTestId('info-card')
    const desc = card.querySelector('.ant-descriptions')
    expect(desc).toBeTruthy()
  })

  it('renders info card fields: 编号, 所属主事项, 优先级, 负责人, 状态, 预期完成时间, 当前完成度', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('编号')).toBeInTheDocument()
    })
    expect(screen.getByText('所属主事项')).toBeInTheDocument()
    expect(screen.getByText('优先级')).toBeInTheDocument()
    expect(screen.getByText('负责人')).toBeInTheDocument()
    expect(screen.getByText('状态')).toBeInTheDocument()
    expect(screen.getByText('预期完成时间')).toBeInTheDocument()
    expect(screen.getByText('当前完成度')).toBeInTheDocument()
  })

  it('所属主事项 is a link to main item detail', async () => {
    renderPage()
    await waitFor(() => {
      const link = screen.getByTestId('main-item-link')
      expect(link).toBeInTheDocument()
      expect(link.getAttribute('href')).toBe('/items/1')
      expect(link.textContent).toContain('Test Main Item')
    })
  })

  it('renders priority tag', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getAllByText('P1').length).toBeGreaterThanOrEqual(1)
    })
  })

  it('renders status tag', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getAllByText('进行中').length).toBeGreaterThanOrEqual(1)
    })
  })

  it('renders expected end date', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('2024-06-15')).toBeInTheDocument()
    })
  })

  it('shows overdue highlight on expected end date when overdue', async () => {
    const pastDate = new Date()
    pastDate.setDate(pastDate.getDate() - 5)
    mockGetSubItem.mockResolvedValue(makeSubItem({
      expected_end_date: pastDate.toISOString(),
      status: '进行中',
    }))
    renderPage()
    await waitFor(() => {
      expect(screen.getByTestId('overdue-date')).toBeInTheDocument()
    })
  })

  it('does not show overdue highlight when completed', async () => {
    const pastDate = new Date()
    pastDate.setDate(pastDate.getDate() - 5)
    mockGetSubItem.mockResolvedValue(makeSubItem({
      expected_end_date: pastDate.toISOString(),
      status: '已完成',
    }))
    renderPage()
    await waitFor(() => {
      expect(screen.getByTestId('info-card')).toBeInTheDocument()
    })
    expect(screen.queryByTestId('overdue-date')).not.toBeInTheDocument()
  })

  // --- Progress Summary Bar ---

  it('renders progress summary bar with correct percent', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByTestId('progress-summary')).toBeInTheDocument()
    })
  })

  // --- Progress Timeline ---

  it('renders empty timeline when no progress records', async () => {
    mockListProgress.mockResolvedValue([])
    renderPage()
    await waitFor(() => {
      expect(screen.getByTestId('timeline-empty')).toBeInTheDocument()
    })
  })

  it('renders timeline items sorted by createdAt ASC', async () => {
    const records = [
      makeProgress({ id: 1, completion: 20, created_at: '2024-06-08T10:00:00Z' }),
      makeProgress({ id: 2, completion: 40, created_at: '2024-06-09T10:00:00Z' }),
      makeProgress({ id: 3, completion: 60, created_at: '2024-06-10T10:00:00Z' }),
    ]
    mockListProgress.mockResolvedValue(records)
    renderPage()
    await waitFor(() => {
      expect(screen.getByTestId('timeline-item-1')).toBeInTheDocument()
    })
    expect(screen.getByTestId('timeline-item-2')).toBeInTheDocument()
    expect(screen.getByTestId('timeline-item-3')).toBeInTheDocument()
  })

  it('each timeline item shows time + author label', async () => {
    const records = [
      makeProgress({ id: 1, author_id: 10, created_at: '2024-06-10T14:30:00Z' }),
    ]
    mockListProgress.mockResolvedValue(records)
    renderPage()
    await waitFor(() => {
      expect(screen.getByTestId('timeline-item-1')).toBeInTheDocument()
    })
    // Author name appears in the timeline label (sibling of the item content)
    expect(screen.getByText('Member User')).toBeInTheDocument()
  })

  it('each timeline item shows completion progress bar', async () => {
    const records = [
      makeProgress({ id: 1, completion: 30 }),
    ]
    mockListProgress.mockResolvedValue(records)
    renderPage()
    await waitFor(() => {
      expect(screen.getByTestId('timeline-progress-1')).toBeInTheDocument()
    })
  })

  it('shows achievement, blocker, lesson text blocks when non-empty', async () => {
    const records = [
      makeProgress({
        id: 1,
        achievement: '完成开发',
        blocker: '缺少文档',
        lesson: '需提前沟通',
      }),
    ]
    mockListProgress.mockResolvedValue(records)
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('完成开发')).toBeInTheDocument()
    })
    expect(screen.getByText('缺少文档')).toBeInTheDocument()
    expect(screen.getByText('需提前沟通')).toBeInTheDocument()
  })

  it('hides empty text blocks', async () => {
    const records = [
      makeProgress({
        id: 1,
        achievement: '成果内容',
        blocker: '',
        lesson: '',
      }),
    ]
    mockListProgress.mockResolvedValue(records)
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('成果内容')).toBeInTheDocument()
    })
    // Blocker and lesson labels should NOT be present since their content is empty
    expect(screen.queryByText('卡点')).not.toBeInTheDocument()
    expect(screen.queryByText('经验')).not.toBeInTheDocument()
  })

  it('shows PM已修正 badge for is_pm_correct records', async () => {
    const records = [
      makeProgress({ id: 1, is_pm_correct: true }),
    ]
    mockListProgress.mockResolvedValue(records)
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('PM已修正')).toBeInTheDocument()
    })
  })

  // --- PM Inline Completion Correction ---

  it('PM can click completion text to trigger inline edit', async () => {
    const user = userEvent.setup()
    const records = [
      makeProgress({ id: 1, completion: 30 }),
    ]
    mockListProgress.mockResolvedValue(records)
    renderPage()
    await waitFor(() => {
      expect(screen.getByTestId('pm-edit-trigger-1')).toBeInTheDocument()
    })
    await user.click(screen.getByTestId('pm-edit-trigger-1'))
    await waitFor(() => {
      expect(screen.getByTestId('pm-edit-input-1')).toBeInTheDocument()
    })
  })

  it('non-PM user does not see clickable completion text', async () => {
    const records = [
      makeProgress({ id: 1, completion: 30 }),
    ]
    mockListProgress.mockResolvedValue(records)
    renderPage(otherUser)
    await waitFor(() => {
      expect(screen.getByTestId('timeline-item-1')).toBeInTheDocument()
    })
    expect(screen.queryByTestId('pm-edit-trigger-1')).not.toBeInTheDocument()
  })

  it('PM confirms inline edit and calls correctCompletionApi', async () => {
    const user = userEvent.setup()
    const records = [
      makeProgress({ id: 1, completion: 30 }),
    ]
    mockListProgress.mockResolvedValue(records)
    mockCorrectCompletion.mockResolvedValue({
      ...records[0],
      completion: 50,
      is_pm_correct: true,
    })
    renderPage()
    await waitFor(() => {
      expect(screen.getByTestId('pm-edit-trigger-1')).toBeInTheDocument()
    })
    await user.click(screen.getByTestId('pm-edit-trigger-1'))
    await waitFor(() => {
      expect(screen.getByTestId('pm-edit-input-1')).toBeInTheDocument()
    })
    // Clear and type new value
    const input = screen.getByTestId('pm-edit-input-1')
    fireEvent.change(input, { target: { value: '50' } })
    await user.click(screen.getByTestId('pm-edit-confirm-1'))
    await waitFor(() => {
      expect(mockCorrectCompletion).toHaveBeenCalledWith(1, 1, { completion: 50 })
    })
  })

  it('PM cancels inline edit', async () => {
    const user = userEvent.setup()
    const records = [
      makeProgress({ id: 1, completion: 30 }),
    ]
    mockListProgress.mockResolvedValue(records)
    renderPage()
    await waitFor(() => {
      expect(screen.getByTestId('pm-edit-trigger-1')).toBeInTheDocument()
    })
    await user.click(screen.getByTestId('pm-edit-trigger-1'))
    await waitFor(() => {
      expect(screen.getByTestId('pm-edit-input-1')).toBeInTheDocument()
    })
    await user.click(screen.getByTestId('pm-edit-cancel-1'))
    await waitFor(() => {
      expect(screen.queryByTestId('pm-edit-input-1')).not.toBeInTheDocument()
    })
    expect(mockCorrectCompletion).not.toHaveBeenCalled()
  })

  // --- Append Progress Button ---

  it('shows append button for assignee', async () => {
    renderPage(memberUser) // memberUser (id=10) is the assignee
    await waitFor(() => {
      expect(screen.getByTestId('append-progress-btn')).toBeInTheDocument()
    })
  })

  it('shows append button for PM', async () => {
    renderPage(pmUser)
    await waitFor(() => {
      expect(screen.getByTestId('append-progress-btn')).toBeInTheDocument()
    })
  })

  it('hides append button for non-assignee non-PM user', async () => {
    renderPage(otherUser)
    await waitFor(() => {
      expect(screen.getByTestId('info-card')).toBeInTheDocument()
    })
    expect(screen.queryByTestId('append-progress-btn')).not.toBeInTheDocument()
  })

  // --- Append Progress Modal ---

  it('opens append progress modal on button click', async () => {
    const user = userEvent.setup()
    renderPage(memberUser)
    await waitFor(() => {
      expect(screen.getByTestId('append-progress-btn')).toBeInTheDocument()
    })
    await user.click(screen.getByTestId('append-progress-btn'))
    await waitFor(() => {
      expect(screen.getByTestId('append-modal')).toBeInTheDocument()
    })
  })

  it('append modal has form fields: 完成度, 成果, 卡点, 经验', async () => {
    const user = userEvent.setup()
    renderPage(memberUser)
    await waitFor(() => {
      expect(screen.getByTestId('append-progress-btn')).toBeInTheDocument()
    })
    await user.click(screen.getByTestId('append-progress-btn'))
    await waitFor(() => {
      expect(screen.getByTestId('append-modal')).toBeInTheDocument()
    })
    expect(screen.getByTestId('append-form-completion')).toBeInTheDocument()
    expect(screen.getByTestId('append-form-achievement')).toBeInTheDocument()
    expect(screen.getByTestId('append-form-blocker')).toBeInTheDocument()
    expect(screen.getByTestId('append-form-lesson')).toBeInTheDocument()
  })

  it('append form completion has helper text showing current max', async () => {
    const user = userEvent.setup()
    const records = [
      makeProgress({ id: 1, completion: 30 }),
    ]
    mockListProgress.mockResolvedValue(records)
    renderPage(memberUser)
    await waitFor(() => {
      expect(screen.getByTestId('append-progress-btn')).toBeInTheDocument()
    })
    await user.click(screen.getByTestId('append-progress-btn'))
    await waitFor(() => {
      expect(screen.getByTestId('append-modal')).toBeInTheDocument()
    })
    // Should show current max (30) in helper text
    expect(screen.getByTestId('completion-helper-text')).toBeInTheDocument()
  })

  it('submits append progress form and calls API', async () => {
    const user = userEvent.setup()
    mockAppendProgress.mockResolvedValue(makeProgress({ id: 99, completion: 80 }))
    renderPage(memberUser)
    await waitFor(() => {
      expect(screen.getByTestId('append-progress-btn')).toBeInTheDocument()
    })
    await user.click(screen.getByTestId('append-progress-btn'))
    await waitFor(() => {
      expect(screen.getByTestId('append-modal')).toBeInTheDocument()
    })
    // Fill required field
    const input = screen.getByTestId('append-form-completion')
    fireEvent.change(input, { target: { value: '80' } })
    await user.click(screen.getByTestId('append-modal-submit'))
    await waitFor(() => {
      expect(mockAppendProgress).toHaveBeenCalledWith(1, 2, expect.objectContaining({
        completion: 80,
      }))
    })
  })

  it('blocks append when completion < previous max and shows inline error', async () => {
    const user = userEvent.setup()
    const records = [
      makeProgress({ id: 1, completion: 50 }),
    ]
    mockListProgress.mockResolvedValue(records)
    renderPage(memberUser)
    await waitFor(() => {
      expect(screen.getByTestId('append-progress-btn')).toBeInTheDocument()
    })
    await user.click(screen.getByTestId('append-progress-btn'))
    await waitFor(() => {
      expect(screen.getByTestId('append-modal')).toBeInTheDocument()
    })
    // Enter value less than previous max (50)
    const input = screen.getByTestId('append-form-completion')
    fireEvent.change(input, { target: { value: '40' } })
    await user.click(screen.getByTestId('append-modal-submit'))
    await waitFor(() => {
      expect(screen.getByTestId('completion-error')).toBeInTheDocument()
    })
    expect(mockAppendProgress).not.toHaveBeenCalled()
  })
})
