import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor, act } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import StatusTransitionDropdown from './StatusTransitionDropdown'
import axios from 'axios'

// Mock API functions
vi.mock('@/api/mainItems', () => ({
  getMainItemTransitionsApi: vi.fn(),
  changeMainItemStatusApi: vi.fn(),
}))

vi.mock('@/api/subItems', () => ({
  getSubItemTransitionsApi: vi.fn(),
  changeSubItemStatusApi: vi.fn(),
}))

vi.mock('@/api/milestones', () => ({
  getMilestoneMapTransitionsApi: vi.fn(),
  changeMilestoneMapStatusApi: vi.fn(),
  getMilestoneTransitionsApi: vi.fn(),
  changeMilestoneStatusApi: vi.fn(),
}))

import {
  getMainItemTransitionsApi,
  changeMainItemStatusApi,
} from '@/api/mainItems'
import {
  getSubItemTransitionsApi,
  changeSubItemStatusApi,
} from '@/api/subItems'
import {
  getMilestoneMapTransitionsApi,
  changeMilestoneMapStatusApi,
  getMilestoneTransitionsApi,
  changeMilestoneStatusApi,
} from '@/api/milestones'

function renderWithQueryClient(ui: React.ReactElement) {
  const qc = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(<QueryClientProvider client={qc}>{ui}</QueryClientProvider>)
}

/** Query for role="alert" bypassing aria-hidden (Radix overlay hides content). */
function queryAlert(container: HTMLElement) {
  return container.querySelector('[role="alert"]')
}

async function findAlert(container: HTMLElement) {
  return waitFor(() => {
    const el = queryAlert(container)
    expect(el).toBeInTheDocument()
    return el!
  })
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
    vi.mocked(getMainItemTransitionsApi).mockResolvedValue([
      'progressing',
      'completed',
    ])
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
    vi.mocked(getMainItemTransitionsApi).mockResolvedValue([
      'progressing',
      'completed',
    ])
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
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
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
      expect(changeMainItemStatusApi).toHaveBeenCalledWith('t1', 'i10', {
        status: 'progressing',
      })
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
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
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
      expect(changeMainItemStatusApi).toHaveBeenCalledWith('t1', 'i10', {
        status: 'completed',
      })
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

  it('closes dropdown when no transitions available', async () => {
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
    // Wait for query to resolve
    await waitFor(() => {
      expect(getMainItemTransitionsApi).toHaveBeenCalled()
    })
    // No tooltip should appear (old behavior removed)
    expect(screen.queryByText('暂无可用流转')).not.toBeInTheDocument()
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
    await waitFor(
      () => {
        expect(changeSubItemStatusApi).toHaveBeenCalledWith('t1', 'i20', {
          status: 'progressing',
        })
      },
      { timeout: 3000 },
    )
  })

  describe('inline error alert', () => {
    function createAxiosError(message: string, status: number = 422) {
      const error = new axios.AxiosError(
        'Request failed',
        undefined,
        undefined,
        undefined,
        {
          status,
          statusText: '',
          headers: {},
          config: {} as never,
          data: { code: 'INVALID_STATUS', message },
        },
      )
      return error
    }

    it('shows inline Alert with backend error message on transition failure', async () => {
      vi.mocked(getMainItemTransitionsApi).mockResolvedValue(['progressing'])
      vi.mocked(changeMainItemStatusApi).mockRejectedValue(
        createAxiosError('该主事项下还有未完成的子事项，无法关闭'),
      )
      const user = userEvent.setup()
      const { container } = renderWithQueryClient(
        <StatusTransitionDropdown
          currentStatus="pending"
          itemType="main"
          teamId="t1"
          itemId="i10"
          onStatusChanged={() => {}}
        />,
      )
      await user.click(screen.getByText('待开始'))
      await waitFor(() => screen.getByText('进行中'))
      await user.click(screen.getByText('进行中'))

      const alert = await findAlert(container)
      expect(alert).toHaveTextContent('该主事项下还有未完成的子事项，无法关闭')
    })

    it('does not auto-dismiss the error Alert', async () => {
      vi.mocked(getMainItemTransitionsApi).mockResolvedValue(['progressing'])
      vi.mocked(changeMainItemStatusApi).mockRejectedValue(
        createAxiosError('错误消息'),
      )
      const user = userEvent.setup()
      const { container } = renderWithQueryClient(
        <StatusTransitionDropdown
          currentStatus="pending"
          itemType="main"
          teamId="t1"
          itemId="i10"
          onStatusChanged={() => {}}
        />,
      )
      await user.click(screen.getByText('待开始'))
      await waitFor(() => screen.getByText('进行中'))
      await user.click(screen.getByText('进行中'))

      const alert = await findAlert(container)
      expect(alert).toBeInTheDocument()

      // Wait 3 seconds — alert should still be there (no auto-dismiss)
      await act(async () => {
        await new Promise((r) => setTimeout(r, 3000))
      })
      expect(queryAlert(container)).toBeInTheDocument()
    })

    it('hides Alert on manual close', async () => {
      vi.mocked(getMainItemTransitionsApi).mockResolvedValue(['progressing'])
      vi.mocked(changeMainItemStatusApi).mockRejectedValue(
        createAxiosError('错误消息'),
      )
      const user = userEvent.setup()
      const { container } = renderWithQueryClient(
        <StatusTransitionDropdown
          currentStatus="pending"
          itemType="main"
          teamId="t1"
          itemId="i10"
          onStatusChanged={() => {}}
        />,
      )
      await user.click(screen.getByText('待开始'))
      await waitFor(() => screen.getByText('进行中'))
      await user.click(screen.getByText('进行中'))

      await findAlert(container)

      // Dismiss dropdown overlay first (body has pointer-events: none while Radix overlay is active)
      await user.keyboard('{Escape}')

      // Click close button
      const closeBtn = container.querySelector('button[aria-label="关闭错误提示"]')
      expect(closeBtn).toBeInTheDocument()
      await user.click(closeBtn!)
      await waitFor(() => {
        expect(queryAlert(container)).not.toBeInTheDocument()
      })
    })

    it('updates Alert content on repeated transition failure', async () => {
      vi.mocked(getMainItemTransitionsApi).mockResolvedValue(['progressing'])
      vi.mocked(changeMainItemStatusApi)
        .mockRejectedValueOnce(createAxiosError('第一次错误'))
        .mockRejectedValueOnce(createAxiosError('第二次错误'))
      const user = userEvent.setup()
      const { container } = renderWithQueryClient(
        <StatusTransitionDropdown
          currentStatus="pending"
          itemType="main"
          teamId="t1"
          itemId="i10"
          onStatusChanged={() => {}}
        />,
      )

      // First failure
      await user.click(screen.getByText('待开始'))
      await waitFor(() => screen.getByText('进行中'))
      await user.click(screen.getByText('进行中'))
      const alert1 = await findAlert(container)
      expect(alert1).toHaveTextContent('第一次错误')

      // Dismiss dropdown overlay by pressing Escape
      await user.keyboard('{Escape}')

      // Second failure — Alert content should update
      await user.click(screen.getByText('待开始'))
      await waitFor(() => screen.getByText('进行中'))
      await user.click(screen.getByText('进行中'))
      const alert2 = await findAlert(container)
      expect(alert2).toHaveTextContent('第二次错误')
    })

    it('clears Alert on successful transition', async () => {
      vi.mocked(getMainItemTransitionsApi).mockResolvedValue(['progressing'])
      vi.mocked(changeMainItemStatusApi)
        .mockRejectedValueOnce(createAxiosError('错误消息'))
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        .mockResolvedValueOnce({} as any)
      const user = userEvent.setup()
      const { container } = renderWithQueryClient(
        <StatusTransitionDropdown
          currentStatus="pending"
          itemType="main"
          teamId="t1"
          itemId="i10"
          onStatusChanged={() => {}}
        />,
      )

      // First: trigger error
      await user.click(screen.getByText('待开始'))
      await waitFor(() => screen.getByText('进行中'))
      await user.click(screen.getByText('进行中'))
      await findAlert(container)

      // Dismiss dropdown overlay
      await user.keyboard('{Escape}')

      // Then: trigger success
      await user.click(screen.getByText('待开始'))
      await waitFor(() => screen.getByText('进行中'))
      await user.click(screen.getByText('进行中'))
      await waitFor(() => {
        expect(queryAlert(container)).not.toBeInTheDocument()
      })
    })
  })

  describe('milestone-map type', () => {
    it('bug: fetches milestone-map transitions, not main-item transitions', async () => {
      vi.mocked(getMilestoneMapTransitionsApi).mockResolvedValue([
        'reviewed',
        'cancelled',
      ])
      const user = userEvent.setup()
      renderWithQueryClient(
        <StatusTransitionDropdown
          currentStatus="planning"
          itemType="milestone-map"
          teamId="t1"
          itemId="map1"
          onStatusChanged={() => {}}
        />,
      )
      await user.click(screen.getByText('规划中'))
      await waitFor(() => {
        expect(getMilestoneMapTransitionsApi).toHaveBeenCalledWith('t1', 'map1')
      })
      // Should NOT call main-item API
      expect(getMainItemTransitionsApi).not.toHaveBeenCalled()
    })

    it('calls changeMilestoneMapStatusApi for milestone-map type', async () => {
      vi.mocked(getMilestoneMapTransitionsApi).mockResolvedValue(['reviewed'])
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      vi.mocked(changeMilestoneMapStatusApi).mockResolvedValue({} as any)
      const user = userEvent.setup()
      renderWithQueryClient(
        <StatusTransitionDropdown
          currentStatus="planning"
          itemType="milestone-map"
          teamId="t1"
          itemId="map1"
          onStatusChanged={() => {}}
        />,
      )
      await user.click(screen.getByText('规划中'))
      await waitFor(() => screen.getByText('已评审'))
      await user.click(screen.getByText('已评审'))
      await waitFor(() => {
        expect(changeMilestoneMapStatusApi).toHaveBeenCalledWith('t1', 'map1', {
          status: 'reviewed',
        })
      })
    })
  })

  describe('milestone type', () => {
    it('bug: fetches milestone transitions, not main-item transitions', async () => {
      vi.mocked(getMilestoneTransitionsApi).mockResolvedValue([
        'in_progress',
        'cancelled',
      ])
      const user = userEvent.setup()
      renderWithQueryClient(
        <StatusTransitionDropdown
          currentStatus="not_started"
          itemType="milestone"
          teamId="t1"
          itemId="ms1"
          onStatusChanged={() => {}}
        />,
      )
      await user.click(screen.getByText('未开始'))
      await waitFor(() => {
        expect(getMilestoneTransitionsApi).toHaveBeenCalledWith('t1', 'ms1')
      })
      expect(getMainItemTransitionsApi).not.toHaveBeenCalled()
    })

    it('calls changeMilestoneStatusApi for milestone type', async () => {
      vi.mocked(getMilestoneTransitionsApi).mockResolvedValue(['in_progress'])
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      vi.mocked(changeMilestoneStatusApi).mockResolvedValue({} as any)
      const user = userEvent.setup()
      renderWithQueryClient(
        <StatusTransitionDropdown
          currentStatus="not_started"
          itemType="milestone"
          teamId="t1"
          itemId="ms1"
          onStatusChanged={() => {}}
        />,
      )
      await user.click(screen.getByText('未开始'))
      await waitFor(() => screen.getByText('进行中'))
      await user.click(screen.getByText('进行中'))
      await waitFor(() => {
        expect(changeMilestoneStatusApi).toHaveBeenCalledWith('t1', 'ms1', {
          status: 'in_progress',
        })
      })
    })
  })
})
