import { createFileRoute } from '@tanstack/react-router'
import { useQuery } from '@tanstack/react-query'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'

interface HealthStatus {
  status: string
  db: string
  latency: string
}

export const Route = createFileRoute('/')({
  component: Dashboard,
})

function Dashboard() {
  const { data: health } = useQuery<HealthStatus>({
    queryKey: ['health'],
    queryFn: () => fetch('/api/health').then((r) => r.json()),
    refetchInterval: 5000,
  })

  return (
    <div className="flex flex-col gap-6">
      <h1 className="text-2xl font-bold">Dashboard</h1>

      <div className="grid grid-cols-3 gap-4">
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <span className={`size-2 rounded-full ${health?.status === 'running' ? 'bg-green-500' : 'bg-red-500'}`} />
              API
            </CardTitle>
            <CardDescription>Server status</CardDescription>
          </CardHeader>
          <CardContent>
            <p className="text-lg font-semibold capitalize">{health?.status ?? 'checking...'}</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <span className={`size-2 rounded-full ${health?.db === 'connected' ? 'bg-green-500' : 'bg-red-500'}`} />
              Database
            </CardTitle>
            <CardDescription>PostgreSQL connection</CardDescription>
          </CardHeader>
          <CardContent>
            <p className="text-lg font-semibold capitalize">{health?.db ?? 'checking...'}</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <span className="size-2 rounded-full bg-blue-500" />
              Latency
            </CardTitle>
            <CardDescription>DB response time</CardDescription>
          </CardHeader>
          <CardContent>
            <p className="text-lg font-semibold">{health?.latency ?? '...'}</p>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
