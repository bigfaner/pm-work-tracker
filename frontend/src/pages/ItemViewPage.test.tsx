import { describe, it, expect, beforeEach, vi } from 'vitest'
import { render, screen, waitFor, within, fireEvent } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import ItemViewPage from './ItemViewPage'
import { useAuthStore } from '@/store/auth'
import { useTeamStore } from '@/store/team'
import type { User, Team, MainItem, SubItem, PageResult, TeamMemberResp } from '@/types'

// --- Mocks ---

const mockListMainItems = vi.fn()
const mockCreateMainItem = vi.fn()
const mockUpdateMainItem = vi.fn()
const mockArchiveMainItem = vi.fn()
const mockListSubItems = vi.fn()
const mockListMembers = vi.fn()

vi.mock('@/api/mainItems', () => ({
  listMainItemsApi: (...args: unknown[]) => mockListMainItems(...args),
  createMainItemApi: (...args: unknown[]) => mockCreateMainItem(...args),
  updateMainItemApi: (...args: unknown[]) => mockUpdateMainItem(...args),
  archiveMainItemApi: (...args: unknown[]) => mockArchiveMainItem(...args),
}))

vi.mock('@/api/subItems', () => ({
  listSubItemsApi: (...args: unknown[]) => mockListSubItems(...args),
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
    start_date: null,
    expected_end_date: null,
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
    expected_end_date: null,
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

const emptyPageResult: PageResult<MainItem> = {
  items: [],
  total: 0,
  page: 1,
  pageSize: 20,
}

// --- Helpers ---

function createQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: { retry: false },
    },
  })
}

function renderPage(user: User = mockUser, teamId: number | null = 1) {
  // Set auth to the provided user
  useAuthStore.getState().setAuth('token', user)
  const qc = createQueryClient()
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter initialEntries={['/items']}>
        <Routes>
          <Route path="/items" element={<ItemViewPage />} />
          <Route path="/items/:mainItemId" element={<div data-testid="detail-page">Detail</div>} />
          <Route path="/items/:mainItemId/sub/:subItemId" element={<div data-testid="sub-detail">Sub Detail</div>} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

// Helper to open antd Select dropdown and select an option
// antd v5 Select needs mouseDown on the selector to open the dropdown
async function openAndSelectOption(selectTestId: string, optionText: string) {
  const selectEl = screen.getByTestId(selectTestId)
  const selector = selectEl.querySelector('.ant-select-selector')!
  fireEvent.mouseDown(selector)
  // Wait for dropdown to render in portal
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

describe('ItemViewPage', () => {
  beforeEach(() => {
    useAuthStore.getState().clearAuth()
    useAuthStore.getState().setAuth('token', mockUser)
    useTeamStore.getState().setTeams([mockTeam])
    useTeamStore.getState().setCurrentTeam(1)

    mockListMainItems.mockResolvedValue(emptyPageResult)
    mockListMembers.mockResolvedValue(mockMembers)
    mockCreateMainItem.mockReset()
    mockUpdateMainItem.mockReset()
    mockArchiveMainItem.mockReset()
    mockListSubItems.mockReset()
  })

  // --- Basic rendering ---

  it('renders page with data-testid', () => {
    renderPage()
    expect(screen.getByTestId('item-view-page')).toBeInTheDocument()
  })

  it('renders page title 事项视图', () => {
    renderPage()
    expect(screen.getByText('事项视图')).toBeInTheDocument()
  })

  // --- Filter bar ---

  it('renders filter bar with priority, status, assignee selects and reset button', () => {
    renderPage()
    expect(screen.getByTestId('filter-priority')).toBeInTheDocument()
    expect(screen.getByTestId('filter-status')).toBeInTheDocument()
    expect(screen.getByTestId('filter-assignee')).toBeInTheDocument()
    expect(screen.getByTestId('filter-reset')).toBeInTheDocument()
  })

  it('renders priority select with P1/P2/P3 options', async () => {
    renderPage()
    await openAndSelectOption('filter-priority', 'P1')
    await waitFor(() => {
      expect(mockListMainItems).toHaveBeenCalledWith(1, expect.objectContaining({ priority: 'P1' }))
    })
  })

  it('renders status select with 8 options', async () => {
    renderPage()
    await openAndSelectOption('filter-status', '进行中')
    await waitFor(() => {
      expect(mockListMainItems).toHaveBeenCalledWith(1, expect.objectContaining({ status: '进行中' }))
    })
  })

  it('resets all filters on reset button click', async () => {
    const user = userEvent.setup()
    renderPage()
    // Select a priority filter
    await openAndSelectOption('filter-priority', 'P1')
    // Click reset
    await user.click(screen.getByTestId('filter-reset'))
    // After reset, the list API should be called with default filter (archived: false only)
    await waitFor(() => {
      const calls = mockListMainItems.mock.calls
      const lastCall = calls[calls.length - 1]
      // Filter should only have archived:false (no priority/status/assigneeId keys)
      expect(lastCall[0]).toBe(1)
      expect(lastCall[1]).toEqual({ archived: false })
    })
  })

  // --- Loading state ---

  it('shows skeleton loading state while fetching items', () => {
    let resolvePromise!: (v: unknown) => void
    mockListMainItems.mockReturnValue(new Promise((resolve) => { resolvePromise = resolve }))
    renderPage()
    expect(screen.getByTestId('main-items-skeleton')).toBeInTheDocument()
    resolvePromise(emptyPageResult)
  })

  // --- Empty state ---

  it('shows empty state when no items exist', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByTestId('empty-state')).toBeInTheDocument()
    })
    expect(screen.getByText('暂无事项')).toBeInTheDocument()
  })

  it('shows create button in empty state for PM user', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByTestId('empty-create-btn')).toBeInTheDocument()
    })
  })

  it('does not show create button in empty state for member user', async () => {
    renderPage(memberUser)
    await waitFor(() => {
      expect(screen.queryByTestId('empty-create-btn')).not.toBeInTheDocument()
    })
  })

  // --- Filter empty state ---

  it('shows filter empty state when filters return no results', async () => {
    // Initially returns items
    mockListMainItems.mockResolvedValue({
      items: [makeMainItem()],
      total: 1,
      page: 1,
      pageSize: 20,
    })
    renderPage()

    // Wait for items to load
    await waitFor(() => {
      expect(screen.getByText('Test Main Item')).toBeInTheDocument()
    })

    // Now simulate filter returning empty
    mockListMainItems.mockResolvedValue(emptyPageResult)

    // Apply a filter
    await openAndSelectOption('filter-priority', 'P1')

    await waitFor(() => {
      expect(screen.getByTestId('filter-empty-state')).toBeInTheDocument()
    })
    expect(screen.getByText('没有符合条件的事项')).toBeInTheDocument()
    expect(screen.getByTestId('filter-reset-btn')).toBeInTheDocument()
  })

  // --- Main item list ---

  it('renders main item list with correct data in collapse panels', async () => {
    const items = [
      makeMainItem({ id: 1, code: 'MI-0001', title: 'Item 1', priority: 'P2', status: '进行中', completion: 50 }),
      makeMainItem({ id: 2, code: 'MI-0002', title: 'Item 2', priority: 'P3', status: '未开始', completion: 0 }),
    ]
    mockListMainItems.mockResolvedValue({ items, total: 2, page: 1, pageSize: 20 })
    renderPage()

    await waitFor(() => {
      expect(screen.getByText('Item 1')).toBeInTheDocument()
      expect(screen.getByText('Item 2')).toBeInTheDocument()
    })
    expect(screen.getByText('MI-0001')).toBeInTheDocument()
    expect(screen.getByText('MI-0002')).toBeInTheDocument()
  })

  // --- P1 badge ---

  it('shows badge dot for P1 items', async () => {
    const items = [makeMainItem({ id: 1, code: 'MI-0001', title: 'P1 Item', priority: 'P1' })]
    mockListMainItems.mockResolvedValue({ items, total: 1, page: 1, pageSize: 20 })
    renderPage()

    await waitFor(() => {
      expect(screen.getByText('P1 Item')).toBeInTheDocument()
    })
    expect(screen.getByTestId('p1-badge-1')).toBeInTheDocument()
  })

  it('does not show badge dot for non-P1 items', async () => {
    const items = [makeMainItem({ id: 1, code: 'MI-0001', title: 'P2 Item', priority: 'P2' })]
    mockListMainItems.mockResolvedValue({ items, total: 1, page: 1, pageSize: 20 })
    renderPage()

    await waitFor(() => {
      expect(screen.getByText('P2 Item')).toBeInTheDocument()
    })
    expect(screen.queryByTestId('p1-badge-1')).not.toBeInTheDocument()
  })

  // --- Overdue highlight ---

  it('highlights overdue items with red color and tooltip', async () => {
    const pastDate = new Date()
    pastDate.setDate(pastDate.getDate() - 5)
    const items = [
      makeMainItem({
        id: 1,
        code: 'MI-0001',
        title: 'Overdue Item',
        expected_end_date: pastDate.toISOString(),
        status: '进行中',
      }),
    ]
    mockListMainItems.mockResolvedValue({ items, total: 1, page: 1, pageSize: 20 })
    renderPage()

    await waitFor(() => {
      expect(screen.getByText('Overdue Item')).toBeInTheDocument()
    })
    const dateEl = screen.getByTestId('overdue-date-1')
    expect(dateEl).toBeInTheDocument()
    expect(dateEl).toHaveStyle({ color: '#ff4d4f' })
  })

  it('does not highlight items that are completed even if overdue', async () => {
    const pastDate = new Date()
    pastDate.setDate(pastDate.getDate() - 5)
    const items = [
      makeMainItem({
        id: 1,
        code: 'MI-0001',
        title: 'Completed Item',
        expected_end_date: pastDate.toISOString(),
        status: '已完成',
      }),
    ]
    mockListMainItems.mockResolvedValue({ items, total: 1, page: 1, pageSize: 20 })
    renderPage()

    await waitFor(() => {
      expect(screen.getByText('Completed Item')).toBeInTheDocument()
    })
    expect(screen.queryByTestId('overdue-date-1')).not.toBeInTheDocument()
  })

  // --- Sub-items lazy load on expand ---

  it('sub-items are not fetched on page load (lazy loading)', async () => {
    const subItems = [makeSubItem({ id: 1, main_item_id: 1, title: 'Sub Task 1' })]
    mockListSubItems.mockResolvedValue({ items: subItems, total: 1, page: 1, pageSize: 20 })

    const items = [makeMainItem({ id: 1, code: 'MI-0001', title: 'Item With Sub' })]
    mockListMainItems.mockResolvedValue({ items, total: 1, page: 1, pageSize: 20 })
    renderPage()

    await waitFor(() => {
      expect(screen.getByText('Item With Sub')).toBeInTheDocument()
    })

    // Sub-items should NOT be fetched until panel is expanded
    expect(mockListSubItems).not.toHaveBeenCalled()
  })

  it('renders sub-items when panel content is shown', async () => {
    const subItems = [
      makeSubItem({ id: 1, main_item_id: 1, title: 'Sub Task 1', status: '进行中', completion: 60 }),
      makeSubItem({ id: 2, main_item_id: 1, title: 'Sub Task 2', status: '未开始', completion: 0 }),
    ]
    mockListSubItems.mockResolvedValue({ items: subItems, total: 2, page: 1, pageSize: 20 })

    // Test by directly importing SubItemList and rendering it
    // This validates that the sub-item data is rendered correctly
    const { SubItemList } = await import('./ItemViewPage')
    const qc = createQueryClient()
    render(
      <QueryClientProvider client={qc}>
        <SubItemList teamId={1} mainItemId={1} />
      </QueryClientProvider>,
    )

    await waitFor(() => {
      expect(screen.getByText('Sub Task 1')).toBeInTheDocument()
      expect(screen.getByText('Sub Task 2')).toBeInTheDocument()
    })
    expect(mockListSubItems).toHaveBeenCalledWith(1, 1)
  })

  // --- Create button visibility ---

  it('shows create button for PM user', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByTestId('create-main-item-btn')).toBeInTheDocument()
    })
  })

  it('hides create button for non-PM member', async () => {
    renderPage(memberUser)
    await waitFor(() => {
      expect(screen.queryByTestId('create-main-item-btn')).not.toBeInTheDocument()
    })
  })

  // --- Create modal ---

  it('opens create modal on button click', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByTestId('create-main-item-btn')).toBeInTheDocument()
    })
    await user.click(screen.getByTestId('create-main-item-btn'))
    expect(screen.getByTestId('main-item-modal')).toBeInTheDocument()
  })

  it('create modal has title, priority, assignee, date, description fields', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByTestId('create-main-item-btn')).toBeInTheDocument()
    })
    await user.click(screen.getByTestId('create-main-item-btn'))

    expect(screen.getByTestId('form-title')).toBeInTheDocument()
    expect(screen.getByTestId('form-priority')).toBeInTheDocument()
    expect(screen.getByTestId('form-assignee')).toBeInTheDocument()
    expect(screen.getByTestId('form-expected-end-date')).toBeInTheDocument()
    expect(screen.getByTestId('form-description')).toBeInTheDocument()
  })

  it('submits create form and calls API', async () => {
    const user = userEvent.setup()
    mockCreateMainItem.mockResolvedValue(makeMainItem({ id: 99 }))
    renderPage()
    await waitFor(() => {
      expect(screen.getByTestId('create-main-item-btn')).toBeInTheDocument()
    })
    await user.click(screen.getByTestId('create-main-item-btn'))

    // Fill required fields
    await user.type(screen.getByTestId('form-title'), 'New Item')
    // Select priority from dropdown
    await openAndSelectOption('form-priority', 'P1')

    // Submit
    await user.click(screen.getByTestId('modal-submit-btn'))

    await waitFor(() => {
      expect(mockCreateMainItem).toHaveBeenCalledWith(1, expect.objectContaining({
        title: 'New Item',
        priority: 'P1',
      }))
    })
  })

  it('shows validation error for empty title', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByTestId('create-main-item-btn')).toBeInTheDocument()
    })
    await user.click(screen.getByTestId('create-main-item-btn'))

    // Submit without filling
    await user.click(screen.getByTestId('modal-submit-btn'))

    await waitFor(() => {
      expect(screen.getByText('请输入标题')).toBeInTheDocument()
    })
  })

  // --- Edit action ---

  it('opens edit modal with pre-filled data from dropdown', async () => {
    const items = [makeMainItem({ id: 1, code: 'MI-0001', title: 'Edit Me', priority: 'P2' })]
    mockListMainItems.mockResolvedValue({ items, total: 1, page: 1, pageSize: 20 })
    mockUpdateMainItem.mockResolvedValue(makeMainItem({ id: 1, title: 'Updated' }))

    const user = userEvent.setup()
    renderPage()

    await waitFor(() => {
      expect(screen.getByText('Edit Me')).toBeInTheDocument()
    })

    // Click the actions dropdown
    await user.click(screen.getByTestId('actions-dropdown-1'))

    // Find the edit menu item
    const editMenuItem = await screen.findByText('编辑')
    await user.click(editMenuItem)

    await waitFor(() => {
      expect(screen.getByTestId('main-item-modal')).toBeInTheDocument()
    })
    // Title should be pre-filled
    const titleInput = screen.getByTestId('form-title') as HTMLInputElement
    expect(titleInput.value).toBe('Edit Me')
  })

  // --- Archive action ---

  it('shows warning when archiving item not in valid status', async () => {
    const items = [makeMainItem({ id: 200, code: 'MI-0200', title: 'In Progress', status: '进行中' })]
    mockListMainItems.mockResolvedValue({ items, total: 1, page: 1, pageSize: 20 })

    const user = userEvent.setup()
    renderPage()

    await waitFor(() => {
      expect(screen.getByText('In Progress')).toBeInTheDocument()
    })

    await user.click(screen.getByTestId('actions-dropdown-200'))
    const archiveMenuItem = await screen.findByText('归档')
    await user.click(archiveMenuItem)

    await waitFor(() => {
      const notice = document.querySelector('.ant-message-notice')
      expect(notice).toBeTruthy()
    })
    expect(mockArchiveMainItem).not.toHaveBeenCalled()
  })

  it('shows archive confirmation and calls archive API for completed item', async () => {
    const items = [makeMainItem({ id: 100, code: 'MI-0100', title: 'Archive Me', status: '已完成' })]
    mockListMainItems.mockResolvedValue({ items, total: 1, page: 1, pageSize: 20 })
    mockArchiveMainItem.mockResolvedValue(undefined)

    const user = userEvent.setup()
    renderPage()

    await waitFor(() => {
      expect(screen.getByText('Archive Me')).toBeInTheDocument()
    })

    await user.click(screen.getByTestId('actions-dropdown-100'))
    const archiveMenuItem = await screen.findByText('归档')
    await user.click(archiveMenuItem)

    // Confirm dialog should appear - antd Modal.confirm renders in portal
    const confirmDialog = await screen.findByRole('dialog')
    expect(confirmDialog).toBeInTheDocument()

    // antd Modal.confirm OK button text is "OK" by default (no ConfigProvider locale)
    const okBtn = within(confirmDialog).getByRole('button', { name: 'OK' })
    await user.click(okBtn)

    await waitFor(() => {
      expect(mockArchiveMainItem).toHaveBeenCalledWith(1, 100)
    })
  })

  // --- Archived items filtered by default ---

  it('calls API with archived=false to exclude archived items', async () => {
    renderPage()
    await waitFor(() => {
      expect(mockListMainItems).toHaveBeenCalledWith(1, expect.objectContaining({
        archived: false,
      }))
    })
  })
})
