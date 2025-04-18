import { computed } from 'vue'
import type { Ref } from 'vue'
import type { IPagedResponse, IToken } from '~/types'

/**
 * Composable for fetching a list of tokens filtered by chainType with pagination.
 *
 * @param chainType - Reactive ref for the chain type to filter tokens.
 * @param limit - Reactive ref for the number of tokens per page.
 * @param offset - Reactive ref for the starting offset.
 * @returns Reactive state including tokens, loading status, errors, and refresh function.
 */
export default function (
  chainType: Ref<string | undefined>,
  limit: Ref<number>,
  offset: Ref<number>
) {
  const { $api } = useNuxtApp()

  const {
    data: tokensData,
    status,
    error,
    refresh
  } = useAsyncData<IPagedResponse<IToken>>(
    'tokensList',
    () => $api.token.listTokens(chainType.value, undefined, limit.value, offset.value),
    {
      watch: [chainType, limit, offset],
      default: () => ({ items: [], limit: limit.value, offset: offset.value, hasMore: false })
    }
  )

  const tokens = computed<IToken[]>(() => tokensData.value?.items || [])
  const hasMore = computed<boolean>(() => tokensData.value?.hasMore || false)
  const isLoading = computed<boolean>(() => status.value === 'pending')

  return {
    tokens,
    hasMore,
    isLoading,
    error,
    refresh,
  }
} 