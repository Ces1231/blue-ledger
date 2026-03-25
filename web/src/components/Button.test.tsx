import { render, screen, fireEvent } from '@testing-library/react'
import { Button } from './Button'

describe('Button', () => {
  it('renders children text', () => {
    render(<Button>Click Me</Button>)
    expect(screen.getByText('Click Me')).toBeInTheDocument()
  })

  it('is enabled by default', () => {
    render(<Button>Submit</Button>)
    expect(screen.getByRole('button')).not.toBeDisabled()
  })

  it('is disabled when disabled prop is set', () => {
    render(<Button disabled>Submit</Button>)
    expect(screen.getByRole('button')).toBeDisabled()
  })

  it('is disabled when loading is true', () => {
    render(<Button loading>Submitting...</Button>)
    expect(screen.getByRole('button')).toBeDisabled()
  })

  it('calls onClick when clicked', () => {
    const handler = vi.fn()
    render(<Button onClick={handler}>Click</Button>)
    fireEvent.click(screen.getByRole('button'))
    expect(handler).toHaveBeenCalledTimes(1)
  })

  it('does not call onClick when disabled', () => {
    const handler = vi.fn()
    render(<Button disabled onClick={handler}>Disabled</Button>)
    fireEvent.click(screen.getByRole('button'))
    expect(handler).not.toHaveBeenCalled()
  })

  it('applies variant class', () => {
    render(<Button variant="danger">Delete</Button>)
    const btn = screen.getByRole('button')
    expect(btn.className).toContain('btn-danger')
  })

  it('applies gold variant class', () => {
    render(<Button variant="gold">Gold</Button>)
    expect(screen.getByRole('button').className).toContain('btn-gold')
  })

  it('applies sm size class', () => {
    render(<Button size="sm">Small</Button>)
    expect(screen.getByRole('button').className).toContain('btn-sm')
  })
})
