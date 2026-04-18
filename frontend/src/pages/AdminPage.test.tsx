import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'

// TODO: Full tests will be written when the page is re-implemented with shadcn/ui

describe('AdminPage', () => {
  it('renders page with data-testid', async () => {
    const { default: AdminPage } = await import('./AdminPage')
    render(<AdminPage />)
    expect(screen.getByTestId('admin-page')).toBeInTheDocument()
  })
})
