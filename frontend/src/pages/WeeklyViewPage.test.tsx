import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'

// TODO: Full tests will be written when the page is re-implemented with shadcn/ui

describe('WeeklyViewPage', () => {
  it('renders page with data-testid', async () => {
    const { default: WeeklyViewPage } = await import('./WeeklyViewPage')
    render(<WeeklyViewPage />)
    expect(screen.getByTestId('weekly-view-page')).toBeInTheDocument()
  })
})
