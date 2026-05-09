import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, it, expect, vi } from 'vitest'
import { WeekPicker } from './WeekPicker'

describe('WeekPicker', () => {
  const weekStart = '2026-04-13'
  const maxWeek = '2026-04-20'

  it('renders the formatted week label', () => {
    render(<WeekPicker weekStart={weekStart} onChange={vi.fn()} maxWeek={maxWeek} />)
    expect(screen.getByText(/2026年第16周/)).toBeInTheDocument()
    expect(screen.getByText(/04\/13 ~ 04\/19/)).toBeInTheDocument()
  })

  it('calls onChange with previous week when prev button clicked', async () => {
    const onChange = vi.fn()
    render(<WeekPicker weekStart={weekStart} onChange={onChange} maxWeek={maxWeek} />)
    await userEvent.click(screen.getByRole('button', { name: /prev/i }))
    expect(onChange).toHaveBeenCalledWith('2026-04-06')
  })

  it('calls onChange with next week when next button clicked', async () => {
    const onChange = vi.fn()
    render(<WeekPicker weekStart={weekStart} onChange={onChange} maxWeek={maxWeek} />)
    await userEvent.click(screen.getByRole('button', { name: /next/i }))
    expect(onChange).toHaveBeenCalledWith('2026-04-20')
  })

  it('disables next button when weekStart equals maxWeek', async () => {
    const onChange = vi.fn()
    render(<WeekPicker weekStart={maxWeek} onChange={onChange} maxWeek={maxWeek} />)
    const nextBtn = screen.getByRole('button', { name: /next/i })
    expect(nextBtn).toBeDisabled()
    await userEvent.click(nextBtn)
    expect(onChange).not.toHaveBeenCalled()
  })

  it('enables next button when weekStart is before maxWeek', () => {
    render(<WeekPicker weekStart={weekStart} onChange={vi.fn()} maxWeek={maxWeek} />)
    expect(screen.getByRole('button', { name: /next/i })).not.toBeDisabled()
  })
})
