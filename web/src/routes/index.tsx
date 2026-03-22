import { useSuspenseQuery } from '@tanstack/react-query'
import { createFileRoute } from '@tanstack/react-router'
import { api } from '#/lib/api'

const options = api.queryOptions('get', '/home')

export const Route = createFileRoute('/')({
  loader: async ({ context }) => {
    return await context.queryClient.ensureQueryData(options)
  },
  component: App,
})

function App() {
  return (
    <div className="container mx-auto">
      <FeaturedAnime />
    </div>
  )
}

function FeaturedAnime() {
  const { data, error } = useSuspenseQuery(options)

  if (error || !data.featuredAnime) {
    return <div>Error: {error?.error ?? 'No Featured Anime'}</div>
  }

  return (
    <div>
      <h2>Featured Anime</h2>
      <div>{data.featuredAnime.jname}</div>
      <img src={data.featuredAnime.metadata?.mainPictureUrl} alt={data.featuredAnime.jname} />
    </div>
  )
}
