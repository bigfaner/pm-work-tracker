import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import PriorityBadge from './PriorityBadge'

describe('PriorityBadge', () => {
  it('renders P1 with high priority variant (red)', () => {
    render(<PriorityBadge priority="P1" />)
    const badge = screen.getByText('P1')
    expect(badge).toBeInTheDocument()
    expect(badge.className).toContain('text-error-text')
  })

  it('renders P2 with medium priority variant (warning)', () => {
    render(<PriorityBadge priority="P2" />)
    const badge = screen.getByText('P2')
    expect(badge).toBeInTheDocument()
    expect(badge.className).toContain('text-warning-text')
  })

  it('renders P3 with low priority variant (secondary)', () => {
    render(<PriorityBadge priority="P3" />)
    const badge = screen.getByText('P3')
    expect(badge).toBeInTheDocument()
    expect(badge.className).toContain('text-secondary')
  })

  it('renders unknown priority with default variant', () => {
    render(<PriorityBadge priority="P4" />)
    expect(screen.getByText('P4')).toBeInTheDocument()
  })

  it('passes through extra className', () => {
    render(<PriorityBadge priority="P1" className="extra-class" />)
    const badge = screen.getByText('P1')
    expect(badge.className).toContain('extra-class')
  })
})
