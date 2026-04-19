import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { Breadcrumb, BreadcrumbItem, BreadcrumbSeparator } from './breadcrumb'

describe('Breadcrumb', () => {
  it('renders breadcrumb with items and separators', () => {
    render(
      <Breadcrumb>
        <BreadcrumbItem href="/">Home</BreadcrumbItem>
        <BreadcrumbSeparator />
        <BreadcrumbItem href="/items">Items</BreadcrumbItem>
        <BreadcrumbSeparator />
        <BreadcrumbItem isCurrent>Detail</BreadcrumbItem>
      </Breadcrumb>
    )
    expect(screen.getByText('Home')).toBeInTheDocument()
    expect(screen.getByText('Items')).toBeInTheDocument()
    expect(screen.getByText('Detail')).toBeInTheDocument()
  })

  it('renders link items as anchors', () => {
    render(
      <Breadcrumb>
        <BreadcrumbItem href="/">Home</BreadcrumbItem>
      </Breadcrumb>
    )
    expect(screen.getByText('Home').tagName).toBe('A')
  })

  it('renders current item as span with text-primary', () => {
    render(
      <Breadcrumb>
        <BreadcrumbItem isCurrent>Current</BreadcrumbItem>
      </Breadcrumb>
    )
    const current = screen.getByText('Current')
    expect(current.tagName).toBe('SPAN')
    expect(current.className).toContain('text-primary')
  })

  it('separator renders slash', () => {
    render(<BreadcrumbSeparator />)
    expect(screen.getByText('/')).toBeInTheDocument()
  })
})
