import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'

// TODO: Full tests will be written when the page is re-implemented with shadcn/ui

describe('ItemViewPage', () => {
  it('renders page with data-testid', async () => {
    const { default: ItemViewPage } = await import('./ItemViewPage')
    render(<ItemViewPage />)
    expect(screen.getByTestId('item-view-page')).toBeInTheDocument()
  })
})
