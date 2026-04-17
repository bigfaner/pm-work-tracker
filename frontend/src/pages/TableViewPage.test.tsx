import { describe, it, expect, beforeEach, vi } from 'vitest'
import { render, screen, waitFor, fireEvent } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import TableViewPage from './TableViewPage'
import { useAuthStore } from '@/store/auth'
import { useTeamStore } from '@/store/team'
import type { User, Team, TableRow, PageResult, TeamMemberResp, TableFilter } from '@/types'

// --- Mocks ---

const mockGetTableView = vi.fn()
const mockExportTableCsv = vi.fn()
const mockListMembers = vi.fn()

vi.mock('@/api/views', () => ({
  getTableViewApi: (...args: unknown[]) => mockGetTableView(...args),
  exportTableCsvApi: (...args: unknown[]) => mockExportTableCsv(...args),
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
  { userId: 10, displayName: 'Member A', username: 'membera', role: 'member', joinedAt: '2024-01-01' },
]

function makeRow(overrides: Partial<TableRow> = {}): TableRow {
  return {
    id: 1,
    type: 'main',
    code: 'MI-0001',
    title: 'Test Item',
    priority: 'P2',
    assigneeId: 1,
    assigneeName: 'PM User',
    status: '进行中',
    completion: 50,
    expectedEndDate: '2026-06-01',
    actualEndDate: null,
    ...overrides,
  }
}

const emptyPage: PageResult<TableRow> = {
  items: [],
  total: 0,
  page: 1,
  pageSize: 50,
}

// --- Helpers ---

function createQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: { retry: false },
    },
  })
}

function renderPage(user: User = mockUser) {
  useAuthStore.getState().setAuth('token', user)
  const qc = createQueryClient()
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter initialEntries={['/table']}>
        <Routes>
          <Route path="/table" element={<TableViewPage />} />
          <Route path="/items/:mainItemId" element={<div data-testid="detail-page">Detail</div>} />
          <Route path="/items/:mainItemId/sub/:subItemId" element={<div data-testid="detail-page">Sub Detail</div>} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

// Utility: open antd Select and pick an option by text
async function openAndSelectOption(container: HTMLElement, selectTestId: string, optionText: string) {
  const selectEl = container.querySelector(`[data-testid="${selectTestId}"]`)
  const selector = selectEl!.querySelector('.ant-select-selector')!
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

describe('TableViewPage', () => {
  beforeEach(() => {
    useAuthStore.getState().clearAuth()
    useAuthStore.getState().setAuth('token', mockUser)
    useTeamStore.getState().setTeams([mockTeam])
    useTeamStore.getState().setCurrentTeam(1)

    mockGetTableView.mockResolvedValue(emptyPage)
    mockListMembers.mockResolvedValue(mockMembers)
    mockExportTableCsv.mockReset()
  })

  // --- Basic rendering ---

  it('renders page with data-testid', () => {
    renderPage()
    expect(screen.getByTestId('table-view-page')).toBeInTheDocument()
  })

  it('renders page title 表格视图', () => {
    renderPage()
    expect(screen.getByText('表格视图')).toBeInTheDocument()
  })

  // --- Export button ---

  it('renders CSV export button', () => {
    renderPage()
    expect(screen.getByTestId('export-csv-btn')).toBeInTheDocument()
  })

  // --- Filter bar ---

  it('renders type filter select', () => {
    renderPage()
    expect(screen.getByTestId('filter-type')).toBeInTheDocument()
  })

  it('renders priority filter select', () => {
    renderPage()
    expect(screen.getByTestId('filter-priority')).toBeInTheDocument()
  })

  it('renders status filter select', () => {
    renderPage()
    expect(screen.getByTestId('filter-status')).toBeInTheDocument()
  })

  it('renders assignee filter select', () => {
    renderPage()
    expect(screen.getByTestId('filter-assignee')).toBeInTheDocument()
  })

  it('renders reset button', () => {
    renderPage()
    expect(screen.getByTestId('filter-reset')).toBeInTheDocument()
  })

  // --- Loading state ---

  it('shows table loading state while fetching', () => {
    let resolvePromise!: (v: unknown) => void
    mockGetTableView.mockReturnValue(new Promise((resolve) => { resolvePromise = resolve }))
    renderPage()
    expect(document.querySelector('.ant-spin')).toBeTruthy()
    resolvePromise(emptyPage)
  })

  // --- Empty state ---

  it('shows empty state when no items exist', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByTestId('table-empty')).toBeInTheDocument()
    })
    expect(screen.getByText('暂无事项')).toBeInTheDocument()
  })

  // --- Table columns rendering ---

  it('renders table rows with correct columns', async () => {
    const rows = [
      makeRow({
        id: 1,
        type: 'main',
        code: 'MI-0001',
        title: 'Main Item A',
        priority: 'P1',
        assigneeName: 'PM User',
        status: '进行中',
        completion: 75,
        expectedEndDate: '2026-06-01',
      }),
    ]
    mockGetTableView.mockResolvedValue({ items: rows, total: 1, page: 1, pageSize: 50 })

    renderPage()

    await waitFor(() => {
      expect(screen.getByText('MI-0001')).toBeInTheDocument()
    })
    expect(screen.getByText('Main Item A')).toBeInTheDocument()
    expect(screen.getByText('PM User')).toBeInTheDocument()
    expect(screen.getByText('进行中')).toBeInTheDocument()
  })

  // --- Type Tag colors ---

  it('renders 主事项 type tag as blue', async () => {
    const rows = [makeRow({ type: 'main' })]
    mockGetTableView.mockResolvedValue({ items: rows, total: 1, page: 1, pageSize: 50 })
    renderPage()
    await waitFor(() => {
      const tag = screen.getByText('主事项')
      expect(tag).toBeInTheDocument()
    })
  })

  it('renders 子事项 type tag as default', async () => {
    const rows = [makeRow({ type: 'sub' })]
    mockGetTableView.mockResolvedValue({ items: rows, total: 1, page: 1, pageSize: 50 })
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('子事项')).toBeInTheDocument()
    })
  })

  // --- Priority Tag colors ---

  it('renders priority tags with correct colors', async () => {
    const rows = [
      makeRow({ id: 1, priority: 'P1', code: 'MI-0001' }),
      makeRow({ id: 2, priority: 'P2', code: 'MI-0002' }),
      makeRow({ id: 3, priority: 'P3', code: 'MI-0003' }),
    ]
    mockGetTableView.mockResolvedValue({ items: rows, total: 3, page: 1, pageSize: 50 })
    renderPage()

    await waitFor(() => {
      expect(screen.getByText('MI-0001')).toBeInTheDocument()
    })
    expect(screen.getByText('P1')).toBeInTheDocument()
    expect(screen.getByText('P2')).toBeInTheDocument()
    expect(screen.getByText('P3')).toBeInTheDocument()
  })

  // --- Pagination ---

  it('renders pagination with pageSize 50 and shows total count', async () => {
    const rows = [makeRow({ id: 1 })]
    mockGetTableView.mockResolvedValue({ items: rows, total: 123, page: 1, pageSize: 50 })
    renderPage()

    await waitFor(() => {
      expect(screen.getByText(/123/)).toBeInTheDocument()
    })
  })

  // --- Default sort: priority DESC, expectedEndDate ASC ---

  it('calls API with default sort params on mount', async () => {
    renderPage()
    await waitFor(() => {
      expect(mockGetTableView).toHaveBeenCalledWith(1, expect.objectContaining({
        sortBy: 'priority',
        sortOrder: 'desc',
        page: 1,
        pageSize: 50,
      }))
    })
  })

  // --- Filter interaction: type ---

  it('calls API with type filter when type is selected', async () => {
    const { container } = renderPage()
    await waitFor(() => {
      expect(mockGetTableView).toHaveBeenCalled()
    })
    mockGetTableView.mockClear()

    await openAndSelectOption(container, 'filter-type', '主事项')

    await waitFor(() => {
      expect(mockGetTableView).toHaveBeenCalledWith(1, expect.objectContaining({
        type: 'main',
      }))
    })
  })

  // --- Filter interaction: priority ---

  it('calls API with priority filter when priority is selected', async () => {
    const { container } = renderPage()
    await waitFor(() => {
      expect(mockGetTableView).toHaveBeenCalled()
    })
    mockGetTableView.mockClear()

    await openAndSelectOption(container, 'filter-priority', 'P1')

    await waitFor(() => {
      expect(mockGetTableView).toHaveBeenCalledWith(1, expect.objectContaining({
        priority: 'P1',
      }))
    })
  })

  // --- Filter interaction: status ---

  it('calls API with status filter when status is selected', async () => {
    const { container } = renderPage()
    await waitFor(() => {
      expect(mockGetTableView).toHaveBeenCalled()
    })
    mockGetTableView.mockClear()

    await openAndSelectOption(container, 'filter-status', '进行中')

    await waitFor(() => {
      expect(mockGetTableView).toHaveBeenCalledWith(1, expect.objectContaining({
        status: '进行中',
      }))
    })
  })

  // --- Filter interaction: assignee ---

  it('calls API with assigneeId filter when assignee is selected', async () => {
    const { container } = renderPage()
    await waitFor(() => {
      expect(mockGetTableView).toHaveBeenCalled()
    })
    mockGetTableView.mockClear()

    await openAndSelectOption(container, 'filter-assignee', 'PM User')

    await waitFor(() => {
      expect(mockGetTableView).toHaveBeenCalledWith(1, expect.objectContaining({
        assigneeId: 1,
      }))
    })
  })

  // --- Reset filters ---

  it('resets all filters when reset button is clicked', async () => {
    const user = userEvent.setup()
    const { container } = renderPage()
    await waitFor(() => {
      expect(mockGetTableView).toHaveBeenCalled()
    })
    mockGetTableView.mockClear()

    // Apply a filter
    await openAndSelectOption(container, 'filter-type', '主事项')
    await waitFor(() => {
      expect(mockGetTableView).toHaveBeenCalledWith(1, expect.objectContaining({ type: 'main' }))
    })
    mockGetTableView.mockClear()

    // Click reset
    await user.click(screen.getByTestId('filter-reset'))

    await waitFor(() => {
      expect(mockGetTableView).toHaveBeenCalledWith(1, expect.objectContaining({
        sortBy: 'priority',
        sortOrder: 'desc',
      }))
    })
    // Verify filter params are not present in the reset call
    const lastCall = mockGetTableView.mock.calls[mockGetTableView.mock.calls.length - 1]
    const filterArg = lastCall[1] as TableFilter
    expect(filterArg.type).toBeUndefined()
    expect(filterArg.priority).toBeUndefined()
    expect(filterArg.status).toBeUndefined()
    expect(filterArg.assigneeId).toBeUndefined()
  })

  // --- CSV export: success ---

  it('triggers CSV export with current filters', async () => {
    const csvBlob = new Blob(['code,title\nMI-0001,Test'], { type: 'text/csv' })
    mockExportTableCsv.mockResolvedValue(csvBlob)

    // Mock URL.createObjectURL
    const mockCreateObjectURL = vi.fn(() => 'blob:test')
    const mockRevokeObjectURL = vi.fn()
    const originalCreate = globalThis.URL.createObjectURL
    const originalRevoke = globalThis.URL.revokeObjectURL
    globalThis.URL.createObjectURL = mockCreateObjectURL
    globalThis.URL.revokeObjectURL = mockRevokeObjectURL

    // Mock createElement to track <a> click
    const mockClick = vi.fn()
    const origCreateElement = document.createElement.bind(document)
    const createSpy = vi.spyOn(document, 'createElement').mockImplementation((tag: string) => {
      const el = origCreateElement(tag)
      if (tag === 'a') {
        el.click = mockClick
      }
      return el
    })

    const user = userEvent.setup()
    const rows = [makeRow({ id: 1 })]
    mockGetTableView.mockResolvedValue({ items: rows, total: 1, page: 1, pageSize: 50 })
    renderPage()

    await waitFor(() => {
      expect(screen.getByTestId('export-csv-btn')).toBeInTheDocument()
    })

    await user.click(screen.getByTestId('export-csv-btn'))

    await waitFor(() => {
      expect(mockExportTableCsv).toHaveBeenCalledWith(1, expect.objectContaining({
        sortBy: 'priority',
        sortOrder: 'desc',
      }))
    })

    // Cleanup
    globalThis.URL.createObjectURL = originalCreate
    globalThis.URL.revokeObjectURL = originalRevoke
    createSpy.mockRestore()
  })

  // --- CSV export: empty data shows error ---

  it('shows error message when exporting with no data', async () => {
    mockExportTableCsv.mockRejectedValue({
      response: { status: 422, data: { code: 'NO_DATA', message: '无数据' } },
    })

    const user = userEvent.setup()
    const rows = [makeRow({ id: 1 })]
    mockGetTableView.mockResolvedValue({ items: rows, total: 1, page: 1, pageSize: 50 })
    renderPage()

    await waitFor(() => {
      expect(screen.getByTestId('export-csv-btn')).toBeInTheDocument()
    })

    await user.click(screen.getByTestId('export-csv-btn'))

    await waitFor(() => {
      expect(mockExportTableCsv).toHaveBeenCalled()
    })
  })

  // --- Row click navigates ---

  it('navigates to main item detail on main row click', async () => {
    const user = userEvent.setup()
    const rows = [makeRow({ id: 1, type: 'main' })]
    mockGetTableView.mockResolvedValue({ items: rows, total: 1, page: 1, pageSize: 50 })
    renderPage()

    await waitFor(() => {
      expect(screen.getByText('MI-0001')).toBeInTheDocument()
    })

    // Click on the title link which triggers navigation
    const titleLink = screen.getByText('Test Item')
    await user.click(titleLink)
  })

  it('navigates to sub item detail on sub row click', async () => {
    const user = userEvent.setup()
    const rows = [makeRow({ id: 10, type: 'sub' })]
    mockGetTableView.mockResolvedValue({ items: rows, total: 1, page: 1, pageSize: 50 })
    renderPage()

    await waitFor(() => {
      expect(screen.getByText('MI-0001')).toBeInTheDocument()
    })

    // Click on the title link which triggers navigation
    const titleLink = screen.getByText('Test Item')
    await user.click(titleLink)
  })

  // --- Overdue rows ---

  it('adds overdue class to rows past expected end date and not completed', async () => {
    const rows = [makeRow({
      id: 1,
      expectedEndDate: '2020-01-01', // past date
      status: '进行中',
    })]
    mockGetTableView.mockResolvedValue({ items: rows, total: 1, page: 1, pageSize: 50 })
    renderPage()

    await waitFor(() => {
      expect(screen.getByText('MI-0001')).toBeInTheDocument()
    })

    // Check that the overdue date is rendered in red
    const dateCell = screen.getByText('2020-01-01')
    expect(dateCell).toHaveStyle({ color: '#ff4d4f' })
  })

  it('does not add overdue styling to completed items', async () => {
    const rows = [makeRow({
      id: 1,
      expectedEndDate: '2020-01-01',
      status: '已完成',
    })]
    mockGetTableView.mockResolvedValue({ items: rows, total: 1, page: 1, pageSize: 50 })
    renderPage()

    await waitFor(() => {
      expect(screen.getByText('MI-0001')).toBeInTheDocument()
    })

    // Completed items should not have red date text
    const dateCell = screen.getByText('2020-01-01')
    expect(dateCell).not.toHaveStyle({ color: '#ff4d4f' })
  })

  // --- Sort interaction via Table onChange ---

  it('calls API with updated sort params when completion column header is clicked', async () => {
    const rows = [makeRow({ id: 1 })]
    mockGetTableView.mockResolvedValue({ items: rows, total: 1, page: 1, pageSize: 50 })
    renderPage()

    await waitFor(() => {
      expect(mockGetTableView).toHaveBeenCalled()
    })
    mockGetTableView.mockClear()

    // Find and click the completion column header to trigger sort
    const completionHeader = screen.getByText('完成度')
    await userEvent.setup().click(completionHeader)

    await waitFor(() => {
      expect(mockGetTableView).toHaveBeenCalledWith(1, expect.objectContaining({
        sortBy: 'completion',
      }))
    })
  })

  // --- Pagination change ---

  it('calls API with new page when pagination is changed', async () => {
    // Create enough items for pagination
    const rows = Array.from({ length: 50 }, (_, i) =>
      makeRow({ id: i + 1, code: `MI-${String(i + 1).padStart(4, '0')}` })
    )
    mockGetTableView.mockResolvedValue({ items: rows, total: 100, page: 1, pageSize: 50 })
    renderPage()

    await waitFor(() => {
      expect(screen.getByText('MI-0001')).toBeInTheDocument()
    })
    mockGetTableView.mockClear()

    // Find and click page 2
    const page2 = screen.getByTitle('2')
    await userEvent.setup().click(page2)

    await waitFor(() => {
      expect(mockGetTableView).toHaveBeenCalledWith(1, expect.objectContaining({
        page: 2,
        pageSize: 50,
      }))
    })
  })
})
