import { createFileRoute, Link } from '@tanstack/react-router'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getUsers, deleteUser, type User } from '@/api/users'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'

export const Route = createFileRoute('/')({
  component: UsersList,
})

function UsersList() {
  const queryClient = useQueryClient()

  const { data: users, isLoading } = useQuery<User[]>({
    queryKey: ['users'],
    queryFn: getUsers,
  })

  const deleteMutation = useMutation({
    mutationFn: deleteUser,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['users'] }),
  })

  if (isLoading) return <p className="text-muted-foreground">Loading...</p>

  return (
    <Card>
      <CardHeader>
        <CardTitle>Users</CardTitle>
      </CardHeader>
      <CardContent>
        <ul className="space-y-2">
          {users?.map((u: User) => (
            <li key={u.id} className="flex items-center justify-between rounded-lg border p-3">
              <span>
                <span className="font-medium">{u.name}</span>
                <span className="text-muted-foreground ml-2">- {u.email}</span>
              </span>
              <span className="flex gap-2">
                <Button asChild variant="outline" size="sm">
                  <Link to="/users/$userId/edit" params={{ userId: String(u.id) }}>
                    Edit
                  </Link>
                </Button>
                <Button
                  variant="destructive"
                  size="sm"
                  onClick={() => deleteMutation.mutate(u.id)}
                >
                  Delete
                </Button>
              </span>
            </li>
          ))}
        </ul>
      </CardContent>
    </Card>
  )
}
