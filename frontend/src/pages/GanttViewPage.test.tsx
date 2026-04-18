import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'

// TODO: Full tests will be written when the page is re-implemented with shadcn/ui

describe('GanttViewPage', () => {
  it('renders page with data-testid', async () => {
    const { default: GanttViewPage } = await import('./GanttViewPage')
    render(<GanttViewPage />)
    expect(screen.getByTestId('gantt-view-page')).toBeInTheDocument()
  })
})
