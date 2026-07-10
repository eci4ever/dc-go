import { createFileRoute, useNavigate, useParams } from '@tanstack/react-router'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getUsers, updateUser, type User } from '@/api/users'
import { useState, useEffect } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'

export const Route = createFileRoute('/users/$userId/edit')({
  component: EditUser,
})

function EditUser() {
  const { userId } = useParams({ from: '/users/$userId/edit' })
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [name, setName] = useState('')
  const [email, setEmail] = useState('')

  const { data: users } = useQuery<User[]>({
    queryKey: ['users'],
    queryFn: getUsers,
  })

  useEffect(() => {
    const user = users?.find((u) => u.id === Number(userId))
    if (user) {
      setName(user.name)
      setEmail(user.email)
    }
  }, [users, userId])

  const mutation = useMutation({
    mutationFn: (data: { name: string; email: string }) => updateUser(Number(userId), data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] })
      navigate({ to: '/' })
    },
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    mutation.mutate({ name, email })
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Edit User</CardTitle>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit} className="flex flex-col gap-4">
          <div className="flex flex-col gap-2">
            <Label htmlFor="name">Name</Label>
            <Input id="name" placeholder="Name" value={name} onChange={(e) => setName(e.target.value)} required />
          </div>
          <div className="flex flex-col gap-2">
            <Label htmlFor="email">Email</Label>
            <Input id="email" type="email" placeholder="Email" value={email} onChange={(e) => setEmail(e.target.value)} required />
          </div>
          <Button type="submit" disabled={mutation.isPending}>
            {mutation.isPending ? 'Saving...' : 'Update'}
          </Button>
        </form>
      </CardContent>
    </Card>
  )
}
