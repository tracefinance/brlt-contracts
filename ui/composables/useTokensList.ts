import { computed } from 'vue'
import type { Ref } from 'vue'
import type { IPagedResponse, IToken, ChainType, TokenType } from '~/types'

/**
 * Composable for fetching a paginated list of tokens, reacting to filter and pagination changes.
 *
 * @param chainTypeFilter - Reactive ref for the chain type filter.
 * @param tokenTypeFilter - Reactive ref for the token type filter.
 * @param limit - Reactive ref for the number of items per page.
 * @param nextToken - Reactive ref for the pagination token.
 * @returns Reactive state including the token list, loading status, errors, pagination info, and refresh function.
 */
export default function (
  chainTypeFilter: Ref<ChainType | null>,
  tokenTypeFilter: Ref<TokenType | null>,
  limit: Ref<number>,
  nextToken: Ref<string | undefined>,
) {
  const { $api } = useNuxtApp()

  const { data, status, error, refresh } = useAsyncData<IPagedResponse<IToken>>(
    'tokensList',
    () =>
      $api.token.listTokens(
        // Pass filter values directly from refs
        {
          chainType: chainTypeFilter.value ?? undefined,
          tokenType: tokenTypeFilter.value ?? undefined,
        },
        limit.value,       // Pass limit as a separate argument
        nextToken.value,   // Pass nextToken as a separate argument
      ),
    {
      watch: [chainTypeFilter, tokenTypeFilter, limit, nextToken], // Re-fetch when any of these change
      default: () => ({ items: [], limit: limit.value, nextToken: undefined }), // Default empty state
    },
  )

  const tokens = computed<IToken[]>(() => data.value?.items ?? [])
  const currentPageLimit = computed<number>(() => data.value?.limit ?? limit.value)
  const nextPageToken = computed<string | undefined>(
    () => data.value?.nextToken,
  )
  const isLoading = computed<boolean>(() => status.value === 'pending')
  const hasMore = computed<boolean>(() => !!nextPageToken.value)

  return {
    tokens,
    isLoading,
    error,
    refresh,
    // Pagination related:
    limit: currentPageLimit, // Return the actual limit from response or the requested one
    nextToken: nextPageToken,
    hasMore,
  }
} 