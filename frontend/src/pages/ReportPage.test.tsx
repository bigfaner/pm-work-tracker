import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'

// TODO: Full tests will be written when the page is re-implemented with shadcn/ui

describe('ReportPage', () => {
  it('renders page with data-testid', async () => {
    const { default: ReportPage } = await import('./ReportPage')
    render(<ReportPage />)
    expect(screen.getByTestId('report-page')).toBeInTheDocument()
  })
})
