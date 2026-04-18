import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'

// TODO: Full tests will be written when the page is re-implemented with shadcn/ui

describe('ItemPoolPage', () => {
  it('renders page with data-testid', async () => {
    const { default: ItemPoolPage } = await import('./ItemPoolPage')
    render(<ItemPoolPage />)
    expect(screen.getByTestId('item-pool-page')).toBeInTheDocument()
  })
})
