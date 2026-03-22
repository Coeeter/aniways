import createFetchClient from 'openapi-fetch'
import createReactQueryClient from 'openapi-react-query'
import type { paths } from './openapi'

const fetchClient = createFetchClient<paths>({
  baseUrl: import.meta.env.VITE_API_BASE_URL,
  credentials: 'include',
})

const reactQueryClient = createReactQueryClient<paths>(fetchClient)

export { reactQueryClient as api, fetchClient as apiFetchClient }
