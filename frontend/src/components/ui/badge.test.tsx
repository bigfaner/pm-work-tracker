import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { Badge } from './badge'

describe('Badge', () => {
  it('renders default variant', () => {
    render(<Badge>Default</Badge>)
    expect(screen.getByText('Default')).toBeInTheDocument()
  })

  it('renders status-planning variant', () => {
    render(<Badge variant="status-planning">Planning</Badge>)
    expect(screen.getByText('Planning').className).toContain('bg-slate-100')
  })

  it('renders status-in-progress variant', () => {
    render(<Badge variant="status-in-progress">In Progress</Badge>)
    expect(screen.getByText('In Progress').className).toContain('bg-blue-50')
  })

  it('renders status-completed variant', () => {
    render(<Badge variant="status-completed">Completed</Badge>)
    expect(screen.getByText('Completed').className).toContain('bg-emerald-50')
  })

  it('renders status-on-hold variant', () => {
    render(<Badge variant="status-on-hold">On Hold</Badge>)
    expect(screen.getByText('On Hold').className).toContain('bg-amber-50')
  })

  it('renders status-cancelled variant', () => {
    render(<Badge variant="status-cancelled">Cancelled</Badge>)
    expect(screen.getByText('Cancelled').className).toContain('bg-red-50')
  })

  it('renders status-overdue variant', () => {
    render(<Badge variant="status-overdue">Overdue</Badge>)
    expect(screen.getByText('Overdue').className).toContain('bg-red-50')
  })

  it('renders status-pending variant', () => {
    render(<Badge variant="status-pending">Pending</Badge>)
    expect(screen.getByText('Pending').className).toContain('bg-yellow-50')
  })

  it('renders priority-high variant', () => {
    render(<Badge variant="priority-high">High</Badge>)
    expect(screen.getByText('High').className).toContain('bg-red-50')
  })

  it('renders priority-medium variant', () => {
    render(<Badge variant="priority-medium">Medium</Badge>)
    expect(screen.getByText('Medium').className).toContain('bg-amber-50')
  })

  it('renders priority-low variant', () => {
    render(<Badge variant="priority-low">Low</Badge>)
    expect(screen.getByText('Low').className).toContain('bg-slate-100')
  })

  it('merges custom className', () => {
    render(<Badge className="extra">Test</Badge>)
    expect(screen.getByText('Test').className).toContain('extra')
  })
})
