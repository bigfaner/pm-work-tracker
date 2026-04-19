import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { Card, CardHeader, CardContent, CardFooter } from './card'

describe('Card', () => {
  it('renders Card with children', () => {
    render(<Card>Card content</Card>)
    expect(screen.getByText('Card content')).toBeInTheDocument()
  })

  it('renders full card composition', () => {
    render(
      <Card>
        <CardHeader>Header</CardHeader>
        <CardContent>Body</CardContent>
        <CardFooter>Footer</CardFooter>
      </Card>
    )
    expect(screen.getByText('Header')).toBeInTheDocument()
    expect(screen.getByText('Body')).toBeInTheDocument()
    expect(screen.getByText('Footer')).toBeInTheDocument()
  })

  it('Card has border and rounded styles', () => {
    render(<Card data-testid="card">Test</Card>)
    const card = screen.getByTestId('card')
    expect(card.className).toContain('rounded-xl')
    expect(card.className).toContain('border-border')
  })

  it('CardHeader has border-bottom', () => {
    render(<CardHeader data-testid="header">H</CardHeader>)
    expect(screen.getByTestId('header').className).toContain('border-b')
  })

  it('CardFooter has bg-alt', () => {
    render(<CardFooter data-testid="footer">F</CardFooter>)
    expect(screen.getByTestId('footer').className).toContain('bg-bg-alt')
  })
})
