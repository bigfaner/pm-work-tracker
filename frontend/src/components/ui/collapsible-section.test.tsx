import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { CollapsibleSection } from './collapsible-section'

describe('CollapsibleSection', () => {
  it('renders title and children', () => {
    render(
      <CollapsibleSection title="团队管理">
        <div>Content here</div>
      </CollapsibleSection>,
    )
    expect(screen.getByText('团队管理')).toBeInTheDocument()
    expect(screen.getByText('Content here')).toBeInTheDocument()
  })

  it('shows content when defaultOpen is true', () => {
    render(
      <CollapsibleSection title="Test" defaultOpen>
        <div>Visible content</div>
      </CollapsibleSection>,
    )
    expect(screen.getByText('Visible content')).toBeVisible()
  })

  it('hides content when defaultOpen is false', () => {
    render(
      <CollapsibleSection title="Test" defaultOpen={false}>
        <div>Hidden content</div>
      </CollapsibleSection>,
    )
    // Content exists in DOM but is hidden via max-height-0
    expect(screen.getByText('Hidden content')).toBeInTheDocument()
  })

  it('toggles content visibility on click', async () => {
    const user = userEvent.setup()
    render(
      <CollapsibleSection title="Test" defaultOpen>
        <div>Toggle me</div>
      </CollapsibleSection>,
    )

    const button = screen.getByText('Test')
    expect(screen.getByText('Toggle me')).toBeVisible()

    await user.click(button)
    // After clicking, the content should be hidden
    const content = screen.getByText('Toggle me')
    expect(content).toBeInTheDocument()

    await user.click(button)
    expect(screen.getByText('Toggle me')).toBeVisible()
  })

  it('has rotate-180 class when open', () => {
    render(
      <CollapsibleSection title="Test" defaultOpen>
        <div>Content</div>
      </CollapsibleSection>,
    )
    const svg = document.querySelector('.rotate-180')
    expect(svg).toBeInTheDocument()
  })

  it('has rotate-0 class when closed', async () => {
    const user = userEvent.setup()
    render(
      <CollapsibleSection title="Test" defaultOpen>
        <div>Content</div>
      </CollapsibleSection>,
    )

    await user.click(screen.getByText('Test'))
    const svg = document.querySelector('.rotate-0')
    expect(svg).toBeInTheDocument()
  })
})
