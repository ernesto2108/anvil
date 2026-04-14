/**
 * Tests para RunsView — vista principal de ejecuciones del dashboard.
 * Se mockea el hook useRunsPolling para aislar el componente de la capa de datos.
 */
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { RunsView } from '../runs-view'
import type { RunDTO } from '@/lib/wails'

// --- Mock del hook de polling -------------------------------------------
vi.mock('@/hooks/use-runs-polling', () => ({
  useRunsPolling: vi.fn(),
}))

import { useRunsPolling } from '@/hooks/use-runs-polling'
const mockUseRunsPolling = vi.mocked(useRunsPolling)

// --- Fixture de datos ---------------------------------------------------
const makeRun = (overrides: Partial<RunDTO> = {}): RunDTO => ({
  id: 'run-abc-001',
  taskId: 'task-1',
  taskDesc: 'Tarea de prueba',
  status: 'success',
  complexity: 'medium',
  provider: 'anthropic',
  project: 'test-project',
  startedAt: '2024-01-15T10:00:00Z',
  endedAt: '2024-01-15T10:05:00Z',
  durationMs: 300_000,
  filesCount: 3,
  agentsCount: 2,
  ...overrides,
})

const RUNS_FIXTURE: RunDTO[] = [
  makeRun({ id: 'run-001', status: 'success' }),
  makeRun({ id: 'run-002', status: 'failed' }),
  makeRun({ id: 'run-003', status: 'running', durationMs: 0 }),
]

// --- Tests ---------------------------------------------------------------
describe('RunsView', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('muestra skeleton de carga cuando runs es null', () => {
    mockUseRunsPolling.mockReturnValue({ runs: null, error: false, refresh: vi.fn() })
    render(<RunsView onRowClick={vi.fn()} />)
    // El skeleton contiene divs con la clase animate-pulse
    const skeletons = document.querySelectorAll('.animate-pulse')
    expect(skeletons.length).toBeGreaterThan(0)
  })

  it('muestra estado vacío cuando no hay runs y no hay filtros', () => {
    mockUseRunsPolling.mockReturnValue({ runs: [], error: false, refresh: vi.fn() })
    render(<RunsView onRowClick={vi.fn()} />)
    expect(screen.getByText('Sin ejecuciones todavía')).toBeInTheDocument()
  })

  it('renderiza filas de la tabla con datos correctos', () => {
    mockUseRunsPolling.mockReturnValue({ runs: RUNS_FIXTURE, error: false, refresh: vi.fn() })
    render(<RunsView onRowClick={vi.fn()} />)
    // Los tres IDs deben aparecer en la tabla
    expect(screen.getByText('run-001')).toBeInTheDocument()
    expect(screen.getByText('run-002')).toBeInTheDocument()
    expect(screen.getByText('run-003')).toBeInTheDocument()
  })

  it('aplica clase bg-running-bg/20 a filas con status running', () => {
    mockUseRunsPolling.mockReturnValue({ runs: RUNS_FIXTURE, error: false, refresh: vi.fn() })
    render(<RunsView onRowClick={vi.fn()} />)
    const runningCell = screen.getByText('run-003')
    // RunRow wraps content in a div (outer) > button (inner); bg class is on the outer div
    const row = runningCell.closest('div.group')
    expect(row).toHaveClass('bg-running-bg/20')
  })

  it('llama a onRowClick con el id correcto al hacer click en una fila', () => {
    const onRowClick = vi.fn()
    mockUseRunsPolling.mockReturnValue({ runs: RUNS_FIXTURE, error: false, refresh: vi.fn() })
    render(<RunsView onRowClick={onRowClick} />)
    const cell = screen.getByText('run-001')
    const row = cell.closest('button')!
    fireEvent.click(row)
    expect(onRowClick).toHaveBeenCalledWith('run-001')
  })

  it('muestra mensaje de filtro vacío cuando hay filtros activos pero no hay resultados', () => {
    // Simular que el hook ya recibió el filtro y devuelve array vacío
    // Pero necesitamos que el estado de filtro interno tenga un valor.
    // Como el estado de filtro es interno al componente, lo manipulamos
    // cambiando el mock para que devuelva [] y luego disparamos un cambio de filtro.
    mockUseRunsPolling.mockReturnValue({ runs: [], error: false, refresh: vi.fn() })
    render(<RunsView onRowClick={vi.fn()} />)

    // Cambiar el select de estado para activar un filtro
    const select = screen.getByRole('combobox')
    fireEvent.change(select, { target: { value: 'failed' } })

    expect(
      screen.getByText('Sin resultados para los filtros seleccionados'),
    ).toBeInTheDocument()
  })
})
