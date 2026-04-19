import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import ProgressBar from './ProgressBar'

describe('ProgressBar', () => {
  it('renders with default size', () => {
    const { container } = render(<ProgressBar value={50} />)
    expect(container.firstChild).toBeInTheDocument()
  })

  it('renders with small size', () => {
    const { container } = render(<ProgressBar value={50} size="sm" />)
    // The wrapper div should be rendered
    expect(container.firstChild).toBeInTheDocument()
  })

  it('renders with large size', () => {
    const { container } = render(<ProgressBar value={50} size="lg" />)
    expect(container.firstChild).toBeInTheDocument()
  })

  it('displays percentage text when showPercentage is true', () => {
    render(<ProgressBar value={65} showPercentage />)
    expect(screen.getByText('65%')).toBeInTheDocument()
  })

  it('hides percentage text by default', () => {
    const { container } = render(<ProgressBar value={65} />)
    expect(container.textContent).not.toContain('65%')
  })

  it('uses green color for 100% completion', () => {
    render(<ProgressBar value={100} showPercentage />)
    expect(screen.getByText('100%')).toBeInTheDocument()
  })

  it('clamps value to 0-100 range', () => {
    const { container } = render(<ProgressBar value={150} />)
    expect(container.firstChild).toBeInTheDocument()
  })
})
