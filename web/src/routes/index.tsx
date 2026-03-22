import { createFileRoute } from '@tanstack/react-router'
import { useState } from 'react'
import { Button } from '#/components/ui/button'

export const Route = createFileRoute('/')({ component: App })

function App() {
  const [count, setCount] = useState(0)

  return (
    <div className="container mx-auto">
      <h1>Aniways</h1>
      <p>Count: {count}</p>
      <Button onClick={() => setCount((c) => c + 1)}>Increment</Button>
    </div>
  )
}
