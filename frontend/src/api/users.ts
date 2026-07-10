const BASE = '/api/users'

export interface User {
  id: number
  name: string
  email: string
}

export async function getUsers(): Promise<User[]> {
  const res = await fetch(BASE)
  if (!res.ok) throw new Error('Failed to fetch users')
  return res.json()
}

export async function createUser(data: Omit<User, 'id'>): Promise<User> {
  const res = await fetch(BASE, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  })
  if (!res.ok) throw new Error('Failed to create user')
  return res.json()
}

export async function updateUser(id: number, data: Omit<User, 'id'>): Promise<User> {
  const res = await fetch(`${BASE}/${id}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  })
  if (!res.ok) throw new Error('Failed to update user')
  return res.json()
}

export async function deleteUser(id: number): Promise<void> {
  const res = await fetch(`${BASE}/${id}`, { method: 'DELETE' })
  if (!res.ok) throw new Error('Failed to delete user')
}
