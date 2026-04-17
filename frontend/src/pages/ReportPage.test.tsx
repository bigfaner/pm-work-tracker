import { describe, it, expect, beforeEach, vi } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import ReportPage from './ReportPage'
import { useAuthStore } from '@/store/auth'
import { useTeamStore } from '@/store/team'
import type { User, Team, ReportPreviewResp } from '@/types'

// --- Mocks ---

const mockGetPreview = vi.fn()
const mockExportMarkdown = vi.fn()

vi.mock('@/api/reports', () => ({
  getWeeklyReportPreviewApi: (...args: unknown[]) => mockGetPreview(...args),
  exportWeeklyReportApi: (...args: unknown[]) => mockExportMarkdown(...args),
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

const mockPreview: ReportPreviewResp = {
  weekStart: '2026-04-13',
  weekEnd: '2026-04-19',
  sections: [
    {
      mainItem: { id: 1, title: 'Main Item A', completion: 60, isKeyItem: true },
      subItems: [
        {
          id: 10,
          title: 'Sub Item A1',
          completion: 80,
          achievements: ['完成了 SDK 初始化', '联调通过'],
          blockers: ['正式环境证书申请中'],
        },
      ],
    },
    {
      mainItem: { id: 2, title: 'Main Item B', completion: 30 },
      subItems: [
        {
          id: 20,
          title: 'Sub Item B1',
          completion: 50,
          achievements: [],
          blockers: [],
        },
      ],
    },
  ],
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
      <MemoryRouter initialEntries={['/report']}>
        <Routes>
          <Route path="/report" element={<ReportPage />} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

// --- Tests ---

describe('ReportPage', () => {
  beforeEach(() => {
    useAuthStore.getState().clearAuth()
    useAuthStore.getState().setAuth('token', mockUser)
    useTeamStore.getState().setTeams([mockTeam])
    useTeamStore.getState().setCurrentTeam(1)

    mockGetPreview.mockReset()
    mockExportMarkdown.mockReset()
  })

  // --- Basic rendering ---

  it('renders page with data-testid', () => {
    renderPage()
    expect(screen.getByTestId('report-page')).toBeInTheDocument()
  })

  it('renders page title 周报导出', () => {
    renderPage()
    expect(screen.getByText('周报导出')).toBeInTheDocument()
  })

  // --- Week picker ---

  it('renders a week picker', () => {
    renderPage()
    expect(screen.getByTestId('week-picker')).toBeInTheDocument()
  })

  // --- Preview button ---

  it('renders 生成预览 button', () => {
    renderPage()
    expect(screen.getByTestId('preview-btn')).toBeInTheDocument()
  })

  // --- Export button ---

  it('renders 导出 Markdown button in preview card header', () => {
    renderPage()
    expect(screen.getByTestId('export-btn')).toBeInTheDocument()
  })

  // --- Empty initial state ---

  it('shows empty state with placeholder text on initial load', () => {
    renderPage()
    expect(screen.getByTestId('preview-empty')).toBeInTheDocument()
    expect(screen.getByText('请选择周次后点击预览')).toBeInTheDocument()
  })

  // --- Preview: loading state ---

  it('shows skeleton loading while fetching preview', async () => {
    let resolvePromise!: (v: unknown) => void
    mockGetPreview.mockReturnValue(new Promise((resolve) => { resolvePromise = resolve }))

    const user = userEvent.setup()
    renderPage()

    await user.click(screen.getByTestId('preview-btn'))

    await waitFor(() => {
      expect(screen.getByTestId('preview-loading')).toBeInTheDocument()
    })

    resolvePromise(mockPreview)
  })

  // --- Preview: success renders structured content ---

  it('renders preview sections with main item titles and completion', async () => {
    mockGetPreview.mockResolvedValue(mockPreview)

    const user = userEvent.setup()
    renderPage()

    await user.click(screen.getByTestId('preview-btn'))

    await waitFor(() => {
      expect(screen.getByText('Main Item A')).toBeInTheDocument()
    })
    expect(screen.getByText('Main Item B')).toBeInTheDocument()
    expect(screen.getByText(/60%/)).toBeInTheDocument()
    expect(screen.getByText(/30%/)).toBeInTheDocument()
  })

  it('renders sub-items within sections', async () => {
    mockGetPreview.mockResolvedValue(mockPreview)

    const user = userEvent.setup()
    renderPage()

    await user.click(screen.getByTestId('preview-btn'))

    await waitFor(() => {
      expect(screen.getByText('Sub Item A1')).toBeInTheDocument()
    })
    expect(screen.getByText('Sub Item B1')).toBeInTheDocument()
  })

  it('renders achievements as bullet list', async () => {
    mockGetPreview.mockResolvedValue(mockPreview)

    const user = userEvent.setup()
    renderPage()

    await user.click(screen.getByTestId('preview-btn'))

    await waitFor(() => {
      expect(screen.getByText('完成了 SDK 初始化')).toBeInTheDocument()
    })
    expect(screen.getByText('联调通过')).toBeInTheDocument()
  })

  it('renders blockers as bullet list', async () => {
    mockGetPreview.mockResolvedValue(mockPreview)

    const user = userEvent.setup()
    renderPage()

    await user.click(screen.getByTestId('preview-btn'))

    await waitFor(() => {
      expect(screen.getByText('正式环境证书申请中')).toBeInTheDocument()
    })
  })

  // --- Key item tag ---

  it('shows 重点 Tag for key items', async () => {
    mockGetPreview.mockResolvedValue(mockPreview)

    const user = userEvent.setup()
    renderPage()

    await user.click(screen.getByTestId('preview-btn'))

    await waitFor(() => {
      expect(screen.getByText('重点')).toBeInTheDocument()
    })
  })

  // --- NO_DATA response: warning ---

  it('shows warning and empty state when API returns NO_DATA', async () => {
    mockGetPreview.mockRejectedValue({
      response: { status: 422, data: { code: 'NO_DATA', message: '所选周暂无数据' } },
    })

    const user = userEvent.setup()
    renderPage()

    await user.click(screen.getByTestId('preview-btn'))

    await waitFor(() => {
      expect(screen.getByTestId('preview-empty')).toBeInTheDocument()
    })
  })

  // --- Export: loading state ---

  it('shows loading on export button while exporting', async () => {
    mockGetPreview.mockResolvedValue(mockPreview)
    let resolveExport!: (v: unknown) => void
    mockExportMarkdown.mockReturnValue(new Promise((resolve) => { resolveExport = resolve }))

    // Mock URL.createObjectURL
    const mockCreateObjectURL = vi.fn(() => 'blob:test')
    const mockRevokeObjectURL = vi.fn()
    const origCreate = globalThis.URL.createObjectURL
    const origRevoke = globalThis.URL.revokeObjectURL
    globalThis.URL.createObjectURL = mockCreateObjectURL
    globalThis.URL.revokeObjectURL = mockRevokeObjectURL

    const user = userEvent.setup()
    renderPage()

    // First preview
    await user.click(screen.getByTestId('preview-btn'))
    await waitFor(() => {
      expect(screen.getByText('Main Item A')).toBeInTheDocument()
    })

    // Then export
    const exportBtn = screen.getByTestId('export-btn')
    await user.click(exportBtn)

    await waitFor(() => {
      expect(exportBtn).toHaveClass('ant-btn-loading')
    })

    resolveExport(new Blob(['# Report'], { type: 'text/markdown' }))

    // Cleanup
    globalThis.URL.createObjectURL = origCreate
    globalThis.URL.revokeObjectURL = origRevoke
  })

  // --- Export: success triggers download ---

  it('triggers file download on successful export', async () => {
    mockGetPreview.mockResolvedValue(mockPreview)
    const mdBlob = new Blob(['# Weekly Report'], { type: 'text/markdown' })
    mockExportMarkdown.mockResolvedValue(mdBlob)

    const mockCreateObjectURL = vi.fn(() => 'blob:test')
    const mockRevokeObjectURL = vi.fn()
    const origCreate = globalThis.URL.createObjectURL
    const origRevoke = globalThis.URL.revokeObjectURL
    globalThis.URL.createObjectURL = mockCreateObjectURL
    globalThis.URL.revokeObjectURL = mockRevokeObjectURL

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
    renderPage()

    // First preview
    await user.click(screen.getByTestId('preview-btn'))
    await waitFor(() => {
      expect(screen.getByText('Main Item A')).toBeInTheDocument()
    })

    // Then export
    await user.click(screen.getByTestId('export-btn'))

    await waitFor(() => {
      expect(mockExportMarkdown).toHaveBeenCalled()
    })
    expect(mockCreateObjectURL).toHaveBeenCalled()

    // Cleanup
    globalThis.URL.createObjectURL = origCreate
    globalThis.URL.revokeObjectURL = origRevoke
    createSpy.mockRestore()
  })

  // --- Export: NO_DATA shows warning ---

  it('shows warning when exporting with NO_DATA', async () => {
    mockGetPreview.mockResolvedValue(mockPreview)
    mockExportMarkdown.mockRejectedValue({
      response: { status: 422, data: { code: 'NO_DATA', message: '所选周暂无数据' } },
    })

    const user = userEvent.setup()
    renderPage()

    // First preview
    await user.click(screen.getByTestId('preview-btn'))
    await waitFor(() => {
      expect(screen.getByText('Main Item A')).toBeInTheDocument()
    })

    // Then export
    await user.click(screen.getByTestId('export-btn'))

    await waitFor(() => {
      expect(mockExportMarkdown).toHaveBeenCalled()
    })
  })

  // --- Preview called with correct weekStart ---

  it('calls preview API with weekStart when preview button is clicked', async () => {
    mockGetPreview.mockResolvedValue(mockPreview)

    const user = userEvent.setup()
    renderPage()

    await user.click(screen.getByTestId('preview-btn'))

    await waitFor(() => {
      expect(mockGetPreview).toHaveBeenCalledWith(1, expect.any(String))
    })
    const weekStartArg = mockGetPreview.mock.calls[0][1] as string
    // Should be a Monday ISO date
    expect(weekStartArg).toMatch(/^\d{4}-\d{2}-\d{2}$/)
  })
})
