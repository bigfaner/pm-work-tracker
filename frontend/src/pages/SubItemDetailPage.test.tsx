import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'

// TODO: Full tests will be written when the page is re-implemented with shadcn/ui

describe('SubItemDetailPage', () => {
  it('renders page with data-testid', async () => {
    const { default: SubItemDetailPage } = await import('./SubItemDetailPage')
    render(<SubItemDetailPage />)
    expect(screen.getByTestId('sub-item-detail-page')).toBeInTheDocument()
  })
})
