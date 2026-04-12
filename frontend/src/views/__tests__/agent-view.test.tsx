/**
 * Tests para AgentView — detalle de un agente dentro de un run.
 * Se mockea @/lib/wails para aislar el componente.
 */
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import type { AgentDetailDTO } from '@/lib/wails'

// --- Mock de @/lib/wails --------------------------------------------------
vi.mock('@/lib/wails', () => ({
  getAgent: vi.fn(),
}))

import { getAgent } from '@/lib/wails'
import { AgentView } from '../agent-view'

const mockGetAgent = vi.mocked(getAgent)

// --- Fixtures -------------------------------------------------------------
const makeDetail = (overrides: Partial<AgentDetailDTO> = {}): AgentDetailDTO => ({
  agent: {
    id: 'agent-001',
    name: 'developer',
    status: 'success',
    durationMs: 120_000,
    startedAt: '2024-01-15T10:00:00Z',
    endedAt: '2024-01-15T10:02:00Z',
    errorMsg: '',
  },
  files: [
    { path: 'src/index.ts', action: 'modified' },
    { path: 'src/utils.ts', action: 'created', diff: '+const x = 1\n-const y = 2' },
  ],
  output: Array.from({ length: 5 }, (_, i) => `Línea de output ${i + 1}`).join('\n'),
  ...overrides,
})

const LONG_OUTPUT = Array.from({ length: 25 }, (_, i) => `Línea ${i + 1}`).join('\n')

const defaultProps = {
  runId: 'run-001',
  agentId: 'agent-001',
  onBack: vi.fn(),
}

// --- Tests ----------------------------------------------------------------
describe('AgentView', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('muestra spinner de carga mientras se resuelve la petición', () => {
    mockGetAgent.mockReturnValue(new Promise(() => {}))
    render(<AgentView {...defaultProps} />)
    const spinner = document.querySelector('.animate-spin')
    expect(spinner).toBeInTheDocument()
  })

  it('muestra "Agente no encontrado." cuando getAgent devuelve null', async () => {
    mockGetAgent.mockResolvedValue(null)
    render(<AgentView {...defaultProps} />)
    await waitFor(() => {
      expect(screen.getByText('Agente no encontrado.')).toBeInTheDocument()
    })
  })

  it('muestra estado de error cuando getAgent rechaza', async () => {
    mockGetAgent.mockRejectedValue(new Error('timeout'))
    render(<AgentView {...defaultProps} />)
    await waitFor(() => {
      expect(screen.getByText('Error al cargar el agente')).toBeInTheDocument()
      expect(screen.getByText('timeout')).toBeInTheDocument()
    })
  })

  it('renderiza todas las MetaCards con los datos del agente', async () => {
    mockGetAgent.mockResolvedValue(makeDetail())
    render(<AgentView {...defaultProps} />)
    await waitFor(() => {
      // Las labels de cada MetaCard
      expect(screen.getByText('Estado')).toBeInTheDocument()
      expect(screen.getByText('Duración')).toBeInTheDocument()
      expect(screen.getByText('Inicio')).toBeInTheDocument()
      expect(screen.getByText('Fin')).toBeInTheDocument()
    })
  })

  it('muestra botón Expandir para output con más de 20 líneas y lo colapsa al hacer click', async () => {
    mockGetAgent.mockResolvedValue(makeDetail({ output: LONG_OUTPUT }))
    render(<AgentView {...defaultProps} />)

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /expandir/i })).toBeInTheDocument()
    })

    // Click en Expandir => debe cambiar a Colapsar
    fireEvent.click(screen.getByRole('button', { name: /expandir/i }))
    expect(screen.getByRole('button', { name: /colapsar/i })).toBeInTheDocument()
  })

  it('renderiza la lista de archivos con rutas y badges de acción', async () => {
    mockGetAgent.mockResolvedValue(makeDetail())
    render(<AgentView {...defaultProps} />)

    await waitFor(() => {
      expect(screen.getByText('src/index.ts')).toBeInTheDocument()
      expect(screen.getByText('modified')).toBeInTheDocument()
      expect(screen.getByText('src/utils.ts')).toBeInTheDocument()
      expect(screen.getByText('created')).toBeInTheDocument()
    })
  })

  it('expande el diff de un archivo al hacer click cuando tiene diff', async () => {
    mockGetAgent.mockResolvedValue(makeDetail())
    render(<AgentView {...defaultProps} />)

    await waitFor(() => {
      expect(screen.getByText('src/utils.ts')).toBeInTheDocument()
    })

    // El archivo con diff tiene un botón que al clickearlo muestra el contenido
    const fileButton = screen.getByText('src/utils.ts').closest('button')!
    fireEvent.click(fileButton)

    await waitFor(() => {
      // El diff contiene "+const x = 1"
      expect(screen.getByText('+const x = 1')).toBeInTheDocument()
    })
  })
})
