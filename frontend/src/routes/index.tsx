import { useEffect } from 'react'
import { createFileRoute } from '@tanstack/react-router'
import { useQuery } from '@tanstack/react-query'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { useAuth } from '@/hooks/use-auth'

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
  const { user, loading, logout } = useAuth()

  useEffect(() => {
    if (!loading && !user) {
      window.location.href = '/login'
    }
  }, [user, loading])

  const { data: health } = useQuery<HealthResponse>({
    queryKey: ['health'],
    queryFn: () => fetch('/api/v1/health').then((r) => r.json()),
    refetchInterval: 5000,
  })

  if (loading) {
    return <p className="text-center text-muted-foreground">Loading...</p>
  }

  if (!user) {
    return null
  }

  const d = health?.data

  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Dashboard</h1>
        <Button variant="outline" onClick={logout}>Sign out</Button>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>{user.name}</CardTitle>
          <CardDescription>{user.email}</CardDescription>
        </CardHeader>
        <CardContent className="text-sm text-muted-foreground">
          Joined {new Date(user.created_at).toLocaleDateString()}
        </CardContent>
      </Card>

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
