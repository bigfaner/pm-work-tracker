import { describe, it, expect, beforeEach, vi } from 'vitest'
import { render, screen, waitFor, within, fireEvent } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import MainItemDetailPage from './MainItemDetailPage'
import { useAuthStore } from '@/store/auth'
import { useTeamStore } from '@/store/team'
import type { User, Team, MainItem, SubItem, TeamMemberResp, PageResult } from '@/types'

// --- Mocks ---

const mockGetMainItem = vi.fn()
const mockListSubItems = vi.fn()
const mockCreateSubItem = vi.fn()
const mockUpdateSubItem = vi.fn()
const mockChangeSubItemStatus = vi.fn()
const mockAssignSubItem = vi.fn()
const mockListMembers = vi.fn()

vi.mock('@/api/mainItems', () => ({
  getMainItemApi: (...args: unknown[]) => mockGetMainItem(...args),
}))

vi.mock('@/api/subItems', () => ({
  listSubItemsApi: (...args: unknown[]) => mockListSubItems(...args),
  createSubItemApi: (...args: unknown[]) => mockCreateSubItem(...args),
  updateSubItemApi: (...args: unknown[]) => mockUpdateSubItem(...args),
  changeSubItemStatusApi: (...args: unknown[]) => mockChangeSubItemStatus(...args),
  assignSubItemApi: (...args: unknown[]) => mockAssignSubItem(...args),
}))

vi.mock('@/api/teams', () => ({
  listMembersApi: (...args: unknown[]) => mockListMembers(...args),
}))

// --- Test Data ---

const mockUser: User = {
  id: 1,
  username: 'pmuser',
  display_name: 'PM User',
  is_super_admin: false,
  can_create_team: false,
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
}

const memberUser: User = {
  ...mockUser,
  id: 10,
  username: 'member',
  display_name: 'Member User',
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
    id: 1,
    team_id: 1,
    main_item_id: 1,
    title: 'Test Sub Item',
    description: '',
    priority: 'P2',
    assignee_id: 1,
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

// --- Helpers ---

function createQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: { retry: false },
    },
  })
}

function renderPage(user: User = mockUser, mainItemId: string = '1') {
  useAuthStore.getState().setAuth('token', user)
  const qc = createQueryClient()
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter initialEntries={[`/items/${mainItemId}`]}>
        <Routes>
          <Route path="/items" element={<div data-testid="item-view">Item View</div>} />
          <Route path="/items/:mainItemId" element={<MainItemDetailPage />} />
          <Route path="/items/:mainItemId/sub/:subItemId" element={<div data-testid="sub-detail">Sub Detail</div>} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

async function openAndSelectOption(selectTestId: string, optionText: string) {
  const selectEl = screen.getByTestId(selectTestId)
  const selector = selectEl.querySelector('.ant-select-selector')!
  fireEvent.mouseDown(selector)
  await waitFor(() => {
    const options = document.querySelectorAll('.ant-select-item-option')
    const match = Array.from(options).find((el) => el.textContent === optionText)
    expect(match).toBeTruthy()
  })
  const options = document.querySelectorAll('.ant-select-item-option')
  const match = Array.from(options).find((el) => el.textContent === optionText)!
  fireEvent.click(match)
}

// --- Tests ---

describe('MainItemDetailPage', () => {
  beforeEach(() => {
    useAuthStore.getState().clearAuth()
    useAuthStore.getState().setAuth('token', mockUser)
    useTeamStore.getState().setTeams([mockTeam])
    useTeamStore.getState().setCurrentTeam(1)

    mockGetMainItem.mockResolvedValue({
      ...makeMainItem(),
      subItems: [],
    })
    mockListSubItems.mockResolvedValue({ items: [], total: 0, page: 1, pageSize: 20 })
    mockListMembers.mockResolvedValue(mockMembers)
    mockCreateSubItem.mockReset()
    mockUpdateSubItem.mockReset()
    mockChangeSubItemStatus.mockReset()
    mockAssignSubItem.mockReset()
  })

  // --- Breadcrumb ---

  it('renders breadcrumb with 事项视图 and item title', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('事项视图')).toBeInTheDocument()
    })
    // Title appears in breadcrumb and header
    expect(screen.getAllByText('Test Main Item').length).toBeGreaterThanOrEqual(1)
  })

  // --- Detail header card ---

  it('renders detail header card with code, title, priority tag, status tag', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByTestId('detail-header')).toBeInTheDocument()
    })
    expect(screen.getByText('MI-0001')).toBeInTheDocument()
    // Title appears in both breadcrumb and header, so use getAllByText
    expect(screen.getAllByText('Test Main Item').length).toBeGreaterThanOrEqual(1)
    expect(screen.getAllByText('P2').length).toBeGreaterThanOrEqual(1)
    expect(screen.getAllByText('进行中').length).toBeGreaterThanOrEqual(1)
  })

  it('renders completion progress bar in header', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByTestId('detail-header')).toBeInTheDocument()
    })
    expect(screen.getByTestId('header-progress')).toBeInTheDocument()
  })

  it('renders start date and expected end date', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByTestId('detail-header')).toBeInTheDocument()
    })
    expect(screen.getByText('2024-06-01')).toBeInTheDocument()
    expect(screen.getByText('2024-07-01')).toBeInTheDocument()
  })

  it('shows overdue highlight when expected end date is past and status is active', async () => {
    const pastDate = new Date()
    pastDate.setDate(pastDate.getDate() - 5)
    mockGetMainItem.mockResolvedValue({
      ...makeMainItem({ expected_end_date: pastDate.toISOString(), status: '进行中' }),
      subItems: [],
    })
    renderPage()
    await waitFor(() => {
      expect(screen.getByTestId('overdue-date')).toBeInTheDocument()
    })
  })

  it('does not show overdue highlight when item is completed', async () => {
    const pastDate = new Date()
    pastDate.setDate(pastDate.getDate() - 5)
    mockGetMainItem.mockResolvedValue({
      ...makeMainItem({ expected_end_date: pastDate.toISOString(), status: '已完成' }),
      subItems: [],
    })
    renderPage()
    await waitFor(() => {
      expect(screen.getByTestId('detail-header')).toBeInTheDocument()
    })
    expect(screen.queryByTestId('overdue-date')).not.toBeInTheDocument()
  })

  // --- Sub-item filter bar ---

  it('renders sub-item filter bar with priority, status, and assignee selects', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByTestId('detail-header')).toBeInTheDocument()
    })
    expect(screen.getByTestId('sub-filter-priority')).toBeInTheDocument()
    expect(screen.getByTestId('sub-filter-status')).toBeInTheDocument()
    expect(screen.getByTestId('sub-filter-assignee')).toBeInTheDocument()
  })

  it('filters sub-items by priority', async () => {
    const subs = [makeSubItem({ id: 1, title: 'Sub 1', priority: 'P1' })]
    mockListSubItems.mockResolvedValue({ items: subs, total: 1, page: 1, pageSize: 20 })
    renderPage()
    await waitFor(() => {
      expect(screen.getByTestId('sub-filter-priority')).toBeInTheDocument()
    })
    await openAndSelectOption('sub-filter-priority', 'P1')
    await waitFor(() => {
      expect(mockListSubItems).toHaveBeenCalledWith(1, 1, expect.objectContaining({ priority: 'P1' }))
    })
  })

  // --- Sub-item table ---

  it('renders sub-item table with correct columns', async () => {
    const subs = [makeSubItem({ id: 1, title: 'Sub Task 1', priority: 'P2', assignee_id: 1, status: '进行中', completion: 60 })]
    mockGetMainItem.mockResolvedValue({
      ...makeMainItem(),
      subItems: subs,
    })
    mockListSubItems.mockResolvedValue({ items: subs, total: 1, page: 1, pageSize: 20 })
    renderPage()

    await waitFor(() => {
      expect(screen.getByText('Sub Task 1')).toBeInTheDocument()
    })
  })

  it('sub-item title links to sub-item detail page', async () => {
    const subs = [makeSubItem({ id: 1, title: 'Linked Sub' })]
    mockGetMainItem.mockResolvedValue({ ...makeMainItem(), subItems: subs })
    mockListSubItems.mockResolvedValue({ items: subs, total: 1, page: 1, pageSize: 20 })
    renderPage()

    await waitFor(() => {
      const link = screen.getByTestId('sub-item-link-1')
      expect(link).toBeInTheDocument()
      expect(link.getAttribute('href')).toBe('/items/1/sub/1')
    })
  })

  // --- Skeleton loading state ---

  it('shows skeleton while loading main item', () => {
    let resolvePromise!: (v: unknown) => void
    mockGetMainItem.mockReturnValue(new Promise((resolve) => { resolvePromise = resolve }))
    renderPage()
    expect(screen.getByTestId('detail-skeleton')).toBeInTheDocument()
    resolvePromise({ ...makeMainItem(), subItems: [] })
  })

  // --- Empty state ---

  it('shows empty state when no sub-items', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByTestId('sub-items-empty')).toBeInTheDocument()
    })
    expect(screen.getByText('暂无子事项')).toBeInTheDocument()
  })

  // --- Create sub-item button ---

  it('shows create sub-item button for PM user', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByTestId('create-sub-item-btn')).toBeInTheDocument()
    })
  })

  it('shows create sub-item button for team member', async () => {
    renderPage(memberUser)
    await waitFor(() => {
      expect(screen.getByTestId('create-sub-item-btn')).toBeInTheDocument()
    })
  })

  // --- Create sub-item modal ---

  it('opens create sub-item modal on button click', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByTestId('create-sub-item-btn')).toBeInTheDocument()
    })
    await user.click(screen.getByTestId('create-sub-item-btn'))
    expect(screen.getByTestId('sub-item-modal')).toBeInTheDocument()
  })

  it('create modal has required form fields', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByTestId('create-sub-item-btn')).toBeInTheDocument()
    })
    await user.click(screen.getByTestId('create-sub-item-btn'))

    expect(screen.getByTestId('sub-form-title')).toBeInTheDocument()
    expect(screen.getByTestId('sub-form-description')).toBeInTheDocument()
    expect(screen.getByTestId('sub-form-priority')).toBeInTheDocument()
    expect(screen.getByTestId('sub-form-assignee')).toBeInTheDocument()
    expect(screen.getByTestId('sub-form-start-date')).toBeInTheDocument()
    expect(screen.getByTestId('sub-form-expected-end-date')).toBeInTheDocument()
  })

  it('submits create sub-item form and calls API', async () => {
    const user = userEvent.setup()
    mockCreateSubItem.mockResolvedValue(makeSubItem({ id: 99 }))
    renderPage()
    await waitFor(() => {
      expect(screen.getByTestId('create-sub-item-btn')).toBeInTheDocument()
    })
    await user.click(screen.getByTestId('create-sub-item-btn'))

    // Fill required fields
    await user.type(screen.getByTestId('sub-form-title'), 'New Sub Item')
    await openAndSelectOption('sub-form-priority', 'P1')
    await openAndSelectOption('sub-form-assignee', 'PM User')

    // Use waitFor to handle potential async timing
    await waitFor(async () => {
      const submitBtn = screen.getByTestId('sub-modal-submit-btn')
      await user.click(submitBtn)
    }, { timeout: 3000 })

    await waitFor(() => {
      expect(mockCreateSubItem).toHaveBeenCalledWith(1, 1, expect.objectContaining({
        title: 'New Sub Item',
        priority: 'P1',
        assigneeId: 1,
      }))
    }, { timeout: 5000 })
  }, 10000)

  it('shows validation error for missing required fields', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByTestId('create-sub-item-btn')).toBeInTheDocument()
    })
    await user.click(screen.getByTestId('create-sub-item-btn'))

    await user.click(screen.getByTestId('sub-modal-submit-btn'))

    await waitFor(() => {
      expect(screen.getByText('请输入标题')).toBeInTheDocument()
    })
  })

  // --- Actions column ---

  it('renders actions column with edit button for PM or assignee', async () => {
    const subs = [makeSubItem({ id: 1, title: 'Sub 1', assignee_id: 1 })]
    mockGetMainItem.mockResolvedValue({ ...makeMainItem(), subItems: subs })
    mockListSubItems.mockResolvedValue({ items: subs, total: 1, page: 1, pageSize: 20 })
    renderPage()

    await waitFor(() => {
      expect(screen.getByTestId('sub-item-edit-1')).toBeInTheDocument()
    })
  })

  it('renders status change button for PM or assignee', async () => {
    const subs = [makeSubItem({ id: 1, title: 'Sub 1', assignee_id: 1 })]
    mockGetMainItem.mockResolvedValue({ ...makeMainItem(), subItems: subs })
    mockListSubItems.mockResolvedValue({ items: subs, total: 1, page: 1, pageSize: 20 })
    renderPage()

    await waitFor(() => {
      expect(screen.getByTestId('sub-item-status-change-1')).toBeInTheDocument()
    })
  })

  it('renders assignee button only for PM', async () => {
    const subs = [makeSubItem({ id: 1, title: 'Sub 1', assignee_id: 1 })]
    mockGetMainItem.mockResolvedValue({ ...makeMainItem(), subItems: subs })
    mockListSubItems.mockResolvedValue({ items: subs, total: 1, page: 1, pageSize: 20 })
    renderPage()

    await waitFor(() => {
      expect(screen.getByTestId('sub-item-assign-1')).toBeInTheDocument()
    })
  })

  it('hides assignee button for non-PM member', async () => {
    const subs = [makeSubItem({ id: 1, title: 'Sub 1', assignee_id: 1 })]
    mockGetMainItem.mockResolvedValue({ ...makeMainItem(), subItems: subs })
    mockListSubItems.mockResolvedValue({ items: subs, total: 1, page: 1, pageSize: 20 })
    renderPage(memberUser)

    await waitFor(() => {
      expect(screen.queryByTestId('sub-item-assign-1')).not.toBeInTheDocument()
    })
  })

  // --- Status change ---

  it('opens status change modal and calls changeSubItemStatusApi', async () => {
    const user = userEvent.setup()
    const subs = [makeSubItem({ id: 1, title: 'Sub 1', assignee_id: 1, status: '进行中' })]
    mockGetMainItem.mockResolvedValue({ ...makeMainItem(), subItems: subs })
    mockListSubItems.mockResolvedValue({ items: subs, total: 1, page: 1, pageSize: 20 })
    mockChangeSubItemStatus.mockResolvedValue(undefined)
    renderPage()

    await waitFor(() => {
      expect(screen.getByTestId('sub-item-status-change-1')).toBeInTheDocument()
    })
    await user.click(screen.getByTestId('sub-item-status-change-1'))

    await waitFor(() => {
      expect(screen.getByTestId('status-change-modal')).toBeInTheDocument()
    })
  })

  // --- Assignee change ---

  it('opens assignee modal and calls assignSubItemApi for PM', async () => {
    const user = userEvent.setup()
    const subs = [makeSubItem({ id: 1, title: 'Sub 1', assignee_id: 1 })]
    mockGetMainItem.mockResolvedValue({ ...makeMainItem(), subItems: subs })
    mockListSubItems.mockResolvedValue({ items: subs, total: 1, page: 1, pageSize: 20 })
    mockAssignSubItem.mockResolvedValue(undefined)
    renderPage()

    await waitFor(() => {
      expect(screen.getByTestId('sub-item-assign-1')).toBeInTheDocument()
    })
    await user.click(screen.getByTestId('sub-item-assign-1'))

    await waitFor(() => {
      expect(screen.getByTestId('assign-modal')).toBeInTheDocument()
    })
  })

  // --- Edit sub-item ---

  it('opens edit modal with pre-filled data', async () => {
    const user = userEvent.setup()
    const subs = [makeSubItem({ id: 1, title: 'Edit Me Sub', assignee_id: 1, priority: 'P2' })]
    mockGetMainItem.mockResolvedValue({ ...makeMainItem(), subItems: subs })
    mockListSubItems.mockResolvedValue({ items: subs, total: 1, page: 1, pageSize: 20 })
    renderPage()

    await waitFor(() => {
      expect(screen.getByTestId('sub-item-edit-1')).toBeInTheDocument()
    })
    await user.click(screen.getByTestId('sub-item-edit-1'))

    await waitFor(() => {
      expect(screen.getByTestId('sub-item-modal')).toBeInTheDocument()
    })
    const titleInput = screen.getByTestId('sub-form-title') as HTMLInputElement
    expect(titleInput.value).toBe('Edit Me Sub')
  })
})
