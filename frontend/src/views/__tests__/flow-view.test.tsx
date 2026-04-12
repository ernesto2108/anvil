/**
 * Tests para FlowView (Session Detail) — vista de detalle de sesión.
 */
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import type { SessionDetailDTO, RunDTO } from '@/lib/wails'

// --- Mock de @/lib/wails ---------------------------------------------------
vi.mock('@/lib/wails', () => ({
  getSessionDetail: vi.fn(),
}))

import { getSessionDetail } from '@/lib/wails'
import { FlowView } from '../flow-view'

const mockGetSessionDetail = vi.mocked(getSessionDetail)

// --- Fixtures -------------------------------------------------------------
const makeRun = (overrides: Partial<RunDTO> = {}): RunDTO => ({
  id: 'run-001',
  taskId: 'task-1',
  taskDesc: 'Tarea de prueba',
  status: 'success',
  complexity: 'medium',
  provider: 'anthropic',
  project: 'test-project',
  startedAt: '2024-01-15T10:00:00Z',
  endedAt: '2024-01-15T10:05:00Z',
  durationMs: 300_000,
  filesCount: 2,
  agentsCount: 1,
  ...overrides,
})

const makeDetail = (overrides: Partial<SessionDetailDTO> = {}): SessionDetailDTO => ({
  run: makeRun(),
  files: [
    { path: 'src/index.ts', action: 'modify', diff: '+const x = 1' },
  ],
  agents: [
    { id: 'agent-1', role: 'developer', status: 'success', durationMs: 120_000, output: 'Done' },
  ],
  ...overrides,
})

// --- Tests ----------------------------------------------------------------
describe('FlowView (Session Detail)', () => {
  const defaultProps = {
    runId: 'run-001',
    onBack: vi.fn(),
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('muestra spinner de carga', () => {
    mockGetSessionDetail.mockReturnValue(new Promise(() => {}))
    render(<FlowView {...defaultProps} />)
    expect(document.querySelector('.animate-spin')).toBeInTheDocument()
  })

  it('muestra estado vacío cuando no se encuentra la sesión', async () => {
    mockGetSessionDetail.mockResolvedValue(null)
    render(<FlowView {...defaultProps} />)
    await waitFor(() => {
      expect(screen.getByText('Sesión no encontrada.')).toBeInTheDocument()
    })
  })

  it('muestra error cuando getSessionDetail rechaza', async () => {
    mockGetSessionDetail.mockRejectedValue(new Error('Error de red'))
    render(<FlowView {...defaultProps} />)
    await waitFor(() => {
      expect(screen.getByText('Error al cargar la sesión')).toBeInTheDocument()
    })
  })

  it('llama a onBack al hacer click en Volver', () => {
    mockGetSessionDetail.mockReturnValue(new Promise(() => {}))
    render(<FlowView {...defaultProps} />)
    fireEvent.click(screen.getByRole('button', { name: /volver/i }))
    expect(defaultProps.onBack).toHaveBeenCalledTimes(1)
  })

  it('muestra archivos y agentes cuando hay datos', async () => {
    mockGetSessionDetail.mockResolvedValue(makeDetail())
    render(<FlowView {...defaultProps} />)
    await waitFor(() => {
      expect(screen.getByText('src/index.ts')).toBeInTheDocument()
      expect(screen.getByText('developer')).toBeInTheDocument()
    })
  })
})
