import { createFileRoute } from '@tanstack/react-router'
import { useQuery } from '@tanstack/react-query'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'

interface HealthData {
  status: string
  db: string
  latency: number
}

interface HealthResponse {
  success: boolean
  data: HealthData
}

export const Route = createFileRoute('/')({
  component: Dashboard,
})

function Dashboard() {
  const { data: health } = useQuery<HealthResponse>({
    queryKey: ['health'],
    queryFn: () => fetch('/api/v1/health').then((r) => r.json()),
    refetchInterval: 5000,
  })

  const d = health?.data

  return (
    <div className="flex flex-col gap-6">
      <h1 className="text-2xl font-bold">Dashboard</h1>

      <div className="grid grid-cols-3 gap-4">
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <span className={`size-2 rounded-full ${d?.status === 'running' ? 'bg-green-500' : 'bg-red-500'}`} />
              API
            </CardTitle>
            <CardDescription>Server status</CardDescription>
          </CardHeader>
          <CardContent>
            <p className="text-lg font-semibold capitalize">{d?.status ?? 'checking...'}</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <span className={`size-2 rounded-full ${d?.db === 'connected' ? 'bg-green-500' : 'bg-red-500'}`} />
              Database
            </CardTitle>
            <CardDescription>PostgreSQL connection</CardDescription>
          </CardHeader>
          <CardContent>
            <p className="text-lg font-semibold capitalize">{d?.db ?? 'checking...'}</p>
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
            <p className="text-lg font-semibold">{d ? `${d.latency}ms` : '...'}</p>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
