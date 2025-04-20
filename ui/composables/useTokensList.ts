import { computed } from 'vue'
import type { Ref } from 'vue'
import type { IPagedResponse, IToken } from '~/types'

/**
 * Composable for fetching a list of tokens filtered by chainType with token-based pagination.
 *
 * @param chainType - Reactive ref for the chain type to filter tokens.
 * @param limit - Reactive ref for the number of tokens per page.
 * @param nextToken - Reactive ref for the pagination token.
 * @returns Reactive state including tokens, loading status, errors, refresh function, and next token.
 */
export default function (
  chainType: Ref<string | undefined>,
  limit: Ref<number>,
  nextToken: Ref<string | undefined>
) {
  const { $api } = useNuxtApp()

  const {
    data: tokensData,
    status,
    error,
    refresh
  } = useAsyncData<IPagedResponse<IToken>>(
    'tokensList',
    () => $api.token.listTokens(chainType.value, undefined, limit.value, nextToken.value),
    {
      watch: [chainType, limit, nextToken],
      default: () => ({ items: [], limit: limit.value, nextToken: undefined })
    }
  )

  const tokens = computed<IToken[]>(() => tokensData.value?.items || [])
  const nextPageToken = computed<string | undefined>(() => tokensData.value?.nextToken)
  const isLoading = computed<boolean>(() => status.value === 'pending')

  return {
    tokens,
    nextPageToken,
    isLoading,
    error,
    refresh,
  }
} 