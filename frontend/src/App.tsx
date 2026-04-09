import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'

function App() {
  return (
    <div className="min-h-screen bg-background flex items-center justify-center p-8">
      <div className="w-full max-w-lg space-y-6">
        <div className="space-y-2 text-center">
          <h1 className="text-3xl font-bold tracking-tight">Anvil Dashboard</h1>
          <p className="text-muted-foreground">
            Observabilidad de ejecuciones de agentes
          </p>
        </div>

        <Card>
          <CardHeader>
            <CardTitle>Scaffold listo</CardTitle>
            <CardDescription>Estado del proyecto</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <p className="text-sm text-muted-foreground">
              Scaffold listo. Las vistas de Runs, Flow, Detail y Metrics se
              cargaran en las proximas tareas (DASH-FEAT-005 al 010).
            </p>
            <Button variant="outline" disabled>
              Ver ejecuciones
            </Button>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}

export default App
