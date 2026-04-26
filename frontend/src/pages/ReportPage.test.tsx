import { describe, it, expect, beforeEach, vi } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import ReportPage from './ReportPage'
import { useTeamStore } from '@/store/team'
import { useAuthStore } from '@/store/auth'
import type { Team } from '@/types'

// Mock the report API
const mockGetPreview = vi.fn()
const mockExportReport = vi.fn()
vi.mock('@/api/reports', () => ({
  getWeeklyReportPreviewApi: (...args: unknown[]) => mockGetPreview(...args),
  exportWeeklyReportApi: (...args: unknown[]) => mockExportReport(...args),
}))

const mockTeams: Team[] = [
  {
    bizKey: '1',
    name: '产品研发团队',
    description: '',
    code: '',
    pmKey: '1',
  createdAt: '2024-01-01',
  updatedAt: '2024-01-01',
  },
]

const mockPreview = {
  weekStart: '2026-04-14',
  weekEnd: '2026-04-20',
  sections: [
    {
      mainItem: { bizKey: '1', title: '用户认证模块开发', completion: 65, isKeyItem: true },
      subItems: [
        {
          bizKey: '1',
          title: '登录页开发',
          completion: 100,
          achievements: ['登录页开发完成，已通过联调测试'],
          blockers: [],
        },
        {
          bizKey: '2',
          title: 'JWT Token 集成',
          completion: 80,
          achievements: ['Token 签发完成，黑名单机制联调中'],
          blockers: ['旧接口 Token 格式兼容问题待修复'],
        },
      ],
    },
  ],
}

function renderReportPage() {
  return render(
    <MemoryRouter initialEntries={[{ pathname: '/report' }]}>
      <Routes>
        <Route path="/report" element={<ReportPage />} />
      </Routes>
    </MemoryRouter>,
  )
}

describe('ReportPage', () => {
  beforeEach(() => {
    useTeamStore.getState().setCurrentTeam('1')
    useTeamStore.getState().setTeams(mockTeams)
    useAuthStore.getState().setPermissions({
      isSuperAdmin: false,
      teamPermissions: { 1: ['report:export'] },
    })
    mockGetPreview.mockReset()
    mockExportReport.mockReset()
  })

  it('renders page with data-testid', async () => {
    const { default: ReportPage } = await import('./ReportPage')
    render(<ReportPage />)
    expect(screen.getByTestId('report-page')).toBeInTheDocument()
  })

  it('renders page title 周报导出', () => {
    renderReportPage()
    expect(screen.getByText('周报导出')).toBeInTheDocument()
  })

  it('renders week selector (WeekPicker)', () => {
    renderReportPage()
    expect(screen.getByLabelText('prev week')).toBeInTheDocument()
    expect(screen.getByLabelText('next week')).toBeInTheDocument()
  })

  it('renders generate preview button', () => {
    renderReportPage()
    expect(screen.getByRole('button', { name: '生成预览' })).toBeInTheDocument()
  })

  it('calls preview API when generate button clicked', async () => {
    mockGetPreview.mockResolvedValue(mockPreview)
    const user = userEvent.setup()
    renderReportPage()

    await user.click(screen.getByRole('button', { name: '生成预览' }))

    await waitFor(() => {
      expect(mockGetPreview).toHaveBeenCalled()
    })
  })

  it('renders preview content after successful API call', async () => {
    mockGetPreview.mockResolvedValue(mockPreview)
    const user = userEvent.setup()
    renderReportPage()

    await user.click(screen.getByRole('button', { name: '生成预览' }))

    await waitFor(() => {
      expect(screen.getByText(/用户认证模块开发/)).toBeInTheDocument()
    })
  })

  it('renders export buttons in preview card header', async () => {
    mockGetPreview.mockResolvedValue(mockPreview)
    const user = userEvent.setup()
    renderReportPage()

    await user.click(screen.getByRole('button', { name: '生成预览' }))

    await waitFor(() => {
      expect(screen.getByTestId('export-personal-btn')).toBeInTheDocument()
      expect(screen.getByTestId('export-full-btn')).toBeInTheDocument()
    })
  })

  it('downloads markdown when full export button clicked', async () => {
    mockGetPreview.mockResolvedValue(mockPreview)

    const mockUrl = 'blob:mock-url'
    URL.createObjectURL = vi.fn(() => mockUrl)
    URL.revokeObjectURL = vi.fn()
    const mockAnchor = { click: vi.fn(), remove: vi.fn(), href: '', download: '' }
    const originalCreateElement = document.createElement.bind(document)
    vi.spyOn(document, 'createElement').mockImplementation((tag: string) => {
      if (tag === 'a') return mockAnchor as any
      return originalCreateElement(tag)
    })

    const user = userEvent.setup()
    renderReportPage()

    await user.click(screen.getByRole('button', { name: '生成预览' }))
    await waitFor(() => {
      expect(screen.getByTestId('export-full-btn')).toBeInTheDocument()
    })

    await user.click(screen.getByTestId('export-full-btn'))

    await waitFor(() => {
      expect(mockAnchor.click).toHaveBeenCalled()
      expect(mockAnchor.download).toMatch(/weekly-report/)
    })

    vi.restoreAllMocks()
  })

  it('shows error when no team is selected', async () => {
    useTeamStore.getState().setCurrentTeam(null)
    const user = userEvent.setup()
    renderReportPage()

    await user.click(screen.getByRole('button', { name: '生成预览' }))

    await waitFor(() => {
      expect(screen.getByText(/请先选择团队/)).toBeInTheDocument()
    })
  })

  it('uses no antd imports', () => {
    const fs = require('fs')
    const path = require('path')
    const source = fs.readFileSync(
      path.resolve(__dirname, 'ReportPage.tsx'),
      'utf-8',
    )
    expect(source).not.toContain('antd')
    expect(source).not.toContain('ant-design')
  })
})
