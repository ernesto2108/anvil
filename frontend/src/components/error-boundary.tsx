import { Component, type ReactNode } from 'react'
import { AlertTriangle, RefreshCw } from 'lucide-react'
import { Button } from '@/components/ui/button'

interface Props {
  children: ReactNode
  fallbackMessage?: string
}

interface State {
  hasError: boolean
  error: Error | null
}

/**
 * Error boundary that catches render crashes in child components.
 * Shows a recovery UI instead of a white screen, with a hint to rebuild
 * the binary when the error likely comes from a frontend/backend mismatch.
 */
export class ErrorBoundary extends Component<Props, State> {
  state: State = { hasError: false, error: null }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error }
  }

  handleRetry = () => {
    this.setState({ hasError: false, error: null })
  }

  render() {
    if (!this.state.hasError) return this.props.children

    const msg = this.state.error?.message ?? 'Error desconocido'
    const isTypeError = this.state.error instanceof TypeError
    const hint = isTypeError
      ? 'Esto suele ocurrir cuando el binario del dashboard está desactualizado. Reconstruye con: wails build'
      : null

    return (
      <div className="flex flex-1 flex-col items-center justify-center gap-4 p-8">
        <div className="rounded-lg border border-fail-border bg-fail-bg px-6 py-5 max-w-lg text-center space-y-3">
          <AlertTriangle size={24} className="text-fail mx-auto" />
          <p className="text-sm font-medium text-fail">
            {this.props.fallbackMessage ?? 'Error al renderizar esta vista'}
          </p>
          <p className="font-mono text-xs text-muted-foreground break-all">{msg}</p>
          {hint && (
            <p className="text-xs text-muted-foreground mt-2">{hint}</p>
          )}
          <Button variant="outline" size="sm" onClick={this.handleRetry} className="gap-1.5 mt-3">
            <RefreshCw size={14} />
            Reintentar
          </Button>
        </div>
      </div>
    )
  }
}
