import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Toast, ToastProvider, useToast } from './toast'

describe('Toast', () => {
  it('renders default toast', () => {
    render(<Toast>Message</Toast>)
    expect(screen.getByText('Message')).toBeInTheDocument()
    expect(screen.getByRole('alert')).toBeInTheDocument()
  })

  it('renders success variant', () => {
    render(<Toast variant="success">Success!</Toast>)
    expect(screen.getByText('Success!').className).toContain('bg-success-bg')
  })

  it('renders error variant', () => {
    render(<Toast variant="error">Error!</Toast>)
    expect(screen.getByText('Error!').className).toContain('bg-error-bg')
  })

  it('renders warning variant', () => {
    render(<Toast variant="warning">Warning!</Toast>)
    expect(screen.getByText('Warning!').className).toContain('bg-warning-bg')
  })
})

describe('ToastProvider + useToast', () => {
  function TestComponent() {
    const { addToast } = useToast()
    return (
      <button onClick={() => addToast('Hello toast', 'success')}>
        Show Toast
      </button>
    )
  }

  it('shows toast via addToast', async () => {
    render(
      <ToastProvider>
        <TestComponent />
      </ToastProvider>
    )
    await userEvent.click(screen.getByText('Show Toast'))
    expect(screen.getByText('Hello toast')).toBeInTheDocument()
  })

  it('removes toast on click', async () => {
    render(
      <ToastProvider>
        <TestComponent />
      </ToastProvider>
    )
    await userEvent.click(screen.getByText('Show Toast'))
    const toast = screen.getByText('Hello toast')
    await userEvent.click(toast)
    expect(screen.queryByText('Hello toast')).not.toBeInTheDocument()
  })
})
