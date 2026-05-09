import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import StatusTransitionDropdown from './StatusTransitionDropdown'

// Mock API functions
vi.mock('@/api/mainItems', () => ({
  getMainItemTransitionsApi: vi.fn(),
  changeMainItemStatusApi: vi.fn(),
}))

vi.mock('@/api/subItems', () => ({
  getSubItemTransitionsApi: vi.fn(),
  changeSubItemStatusApi: vi.fn(),
}))

import {
  getMainItemTransitionsApi,
  changeMainItemStatusApi,
} from '@/api/mainItems'
import {
  getSubItemTransitionsApi,
  changeSubItemStatusApi,
} from '@/api/subItems'

function renderWithQueryClient(ui: React.ReactElement) {
  const qc = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={qc}>{ui}</QueryClientProvider>,
  )
}

describe('StatusTransitionDropdown', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders StatusBadge with current status', () => {
    renderWithQueryClient(
      <StatusTransitionDropdown
        currentStatus="pending"
        itemType="main"
        teamId="t1"
        itemId="i10"
        onStatusChanged={() => {}}
      />,
    )
    expect(screen.getByText('待开始')).toBeInTheDocument()
  })

  it('fetches main item transitions when dropdown opens', async () => {
    vi.mocked(getMainItemTransitionsApi).mockResolvedValue(['progressing', 'completed'])
    const user = userEvent.setup()
    renderWithQueryClient(
      <StatusTransitionDropdown
        currentStatus="pending"
        itemType="main"
        teamId="t1"
        itemId="i10"
        onStatusChanged={() => {}}
      />,
    )
    // Click the badge to open dropdown
    await user.click(screen.getByText('待开始'))
    await waitFor(() => {
      expect(getMainItemTransitionsApi).toHaveBeenCalledWith('t1', 'i10')
    })
  })

  it('fetches sub item transitions when dropdown opens', async () => {
    vi.mocked(getSubItemTransitionsApi).mockResolvedValue(['progressing'])
    const user = userEvent.setup()
    renderWithQueryClient(
      <StatusTransitionDropdown
        currentStatus="pending"
        itemType="sub"
        teamId="t1"
        itemId="i20"
        onStatusChanged={() => {}}
      />,
    )
    await user.click(screen.getByText('待开始'))
    await waitFor(() => {
      expect(getSubItemTransitionsApi).toHaveBeenCalledWith('t1', 'i20')
    })
  })

  it('renders transition options in dropdown', async () => {
    vi.mocked(getMainItemTransitionsApi).mockResolvedValue(['progressing', 'completed'])
    const user = userEvent.setup()
    renderWithQueryClient(
      <StatusTransitionDropdown
        currentStatus="pending"
        itemType="main"
        teamId="t1"
        itemId="i10"
        onStatusChanged={() => {}}
      />,
    )
    await user.click(screen.getByText('待开始'))
    await waitFor(() => {
      expect(screen.getByText('进行中')).toBeInTheDocument()
      expect(screen.getByText('已完成')).toBeInTheDocument()
    })
  })

  it('calls change API and onStatusChanged for non-terminal status', async () => {
    vi.mocked(getMainItemTransitionsApi).mockResolvedValue(['progressing'])
    vi.mocked(changeMainItemStatusApi).mockResolvedValue({} as any)
    const onStatusChanged = vi.fn()
    const user = userEvent.setup()
    renderWithQueryClient(
      <StatusTransitionDropdown
        currentStatus="pending"
        itemType="main"
        teamId="t1"
        itemId="i10"
        onStatusChanged={onStatusChanged}
      />,
    )
    await user.click(screen.getByText('待开始'))
    await waitFor(() => screen.getByText('进行中'))
    await user.click(screen.getByText('进行中'))

    await waitFor(() => {
      expect(changeMainItemStatusApi).toHaveBeenCalledWith('t1', 'i10', { status: 'progressing' })
    })
  })

  it('shows confirmation dialog for terminal status', async () => {
    vi.mocked(getMainItemTransitionsApi).mockResolvedValue(['completed'])
    const user = userEvent.setup()
    renderWithQueryClient(
      <StatusTransitionDropdown
        currentStatus="progressing"
        itemType="main"
        teamId="t1"
        itemId="i10"
        onStatusChanged={() => {}}
      />,
    )
    await user.click(screen.getByText('进行中'))
    await waitFor(() => screen.getByText('已完成'))
    await user.click(screen.getByText('已完成'))

    // Confirmation dialog should appear
    expect(screen.getByText('确认变更状态')).toBeInTheDocument()
    expect(screen.getByText(/确认将状态变更为「已完成」/)).toBeInTheDocument()
  })

  it('calls change API on confirm in confirmation dialog', async () => {
    vi.mocked(getMainItemTransitionsApi).mockResolvedValue(['completed'])
    vi.mocked(changeMainItemStatusApi).mockResolvedValue({} as any)
    const onStatusChanged = vi.fn()
    const user = userEvent.setup()
    renderWithQueryClient(
      <StatusTransitionDropdown
        currentStatus="progressing"
        itemType="main"
        teamId="t1"
        itemId="i10"
        onStatusChanged={onStatusChanged}
      />,
    )
    await user.click(screen.getByText('进行中'))
    await waitFor(() => screen.getByText('已完成'))
    await user.click(screen.getByText('已完成'))

    // Confirm in dialog
    await user.click(screen.getByRole('button', { name: '确认' }))

    await waitFor(() => {
      expect(changeMainItemStatusApi).toHaveBeenCalledWith('t1', 'i10', { status: 'completed' })
    })
  })

  it('calls onBeforeTerminalStatus before showing confirmation', async () => {
    vi.mocked(getSubItemTransitionsApi).mockResolvedValue(['completed'])
    const onBeforeTerminalStatus = vi.fn().mockResolvedValue(true)
    const user = userEvent.setup()
    renderWithQueryClient(
      <StatusTransitionDropdown
        currentStatus="progressing"
        itemType="sub"
        teamId="t1"
        itemId="i20"
        onStatusChanged={() => {}}
        onBeforeTerminalStatus={onBeforeTerminalStatus}
      />,
    )
    await user.click(screen.getByText('进行中'))
    await waitFor(() => screen.getByText('已完成'))
    await user.click(screen.getByText('已完成'))

    // onBeforeTerminalStatus should be called
    expect(onBeforeTerminalStatus).toHaveBeenCalledWith('completed')
  })

  it('skips confirmation when onBeforeTerminalStatus returns false', async () => {
    vi.mocked(getSubItemTransitionsApi).mockResolvedValue(['completed'])
    const onBeforeTerminalStatus = vi.fn().mockResolvedValue(false)
    const user = userEvent.setup()
    renderWithQueryClient(
      <StatusTransitionDropdown
        currentStatus="progressing"
        itemType="sub"
        teamId="t1"
        itemId="i20"
        onStatusChanged={() => {}}
        onBeforeTerminalStatus={onBeforeTerminalStatus}
      />,
    )
    await user.click(screen.getByText('进行中'))
    await waitFor(() => screen.getByText('已完成'))
    await user.click(screen.getByText('已完成'))

    // Wait for async callback
    await waitFor(() => {
      expect(onBeforeTerminalStatus).toHaveBeenCalledWith('completed')
    })
    // Confirmation dialog should NOT appear
    expect(screen.queryByText('确认变更状态')).not.toBeInTheDocument()
  })

  it('shows tooltip when no transitions available', async () => {
    vi.mocked(getMainItemTransitionsApi).mockResolvedValue([])
    const user = userEvent.setup()
    renderWithQueryClient(
      <StatusTransitionDropdown
        currentStatus="completed"
        itemType="main"
        teamId="t1"
        itemId="i10"
        onStatusChanged={() => {}}
      />,
    )
    await user.click(screen.getByText('已完成'))
    // Wait for query to resolve and tooltip to appear
    await waitFor(() => {
      expect(getMainItemTransitionsApi).toHaveBeenCalled()
    })
    // The useEffect sets showTip=true after transitions resolve empty
    await waitFor(() => {
      expect(screen.getByText('暂无可用流转')).toBeInTheDocument()
    })
  })

  it('does not render button when disabled', () => {
    renderWithQueryClient(
      <StatusTransitionDropdown
        currentStatus="pending"
        itemType="main"
        teamId="t1"
        itemId="i10"
        onStatusChanged={() => {}}
        disabled
      />,
    )
    // Badge should render but not as a button (no cursor-pointer)
    expect(screen.getByText('待开始')).toBeInTheDocument()
    // There should be no button triggering the dropdown
    expect(screen.queryByRole('button')).not.toBeInTheDocument()
  })

  it('uses sub item API for itemType sub', async () => {
    vi.mocked(getSubItemTransitionsApi).mockResolvedValue(['progressing'])
    vi.mocked(changeSubItemStatusApi).mockResolvedValue(undefined)
    const user = userEvent.setup()
    renderWithQueryClient(
      <StatusTransitionDropdown
        currentStatus="pending"
        itemType="sub"
        teamId="t1"
        itemId="i20"
        onStatusChanged={() => {}}
      />,
    )
    // Verify sub item transitions API is called
    await user.click(screen.getByText('待开始'))
    await waitFor(() => {
      expect(getSubItemTransitionsApi).toHaveBeenCalledWith('t1', 'i20')
    })
    // Verify the transition option renders
    await waitFor(() => {
      expect(screen.getByText('进行中')).toBeInTheDocument()
    })
    // Select the status and verify change API is called
    await user.click(screen.getByText('进行中'))
    await waitFor(() => {
      expect(changeSubItemStatusApi).toHaveBeenCalledWith('t1', 'i20', { status: 'progressing' })
    }, { timeout: 3000 })
  })
})
