import { ref, onMounted } from 'vue'
import type { Ref } from 'vue'
import type { ChainType, TokenType } from '~/types'

/**
 * Composable to manage token list filters and initialize them from URL query parameters.
 *
 * @returns Reactive refs for filters and a function to clear them.
 */
export default function () {
  const route = useRoute()

  // Reactive refs for filter state
  const chainTypeFilter: Ref<ChainType | null> = ref(null)
  const tokenTypeFilter: Ref<TokenType | null> = ref(null)

  // Initialize filters from URL query parameters on component mount
  onMounted(() => {
    // Use snake_case from URL query
    const initialChainType = route.query.chain_type
    const initialTokenType = route.query.token_type

    if (
      initialChainType &&
      typeof initialChainType === 'string' &&
      isValidChainType(initialChainType)
    ) {
      chainTypeFilter.value = initialChainType
    }

    if (
      initialTokenType &&
      typeof initialTokenType === 'string' &&
      isValidTokenType(initialTokenType)
    ) {
      tokenTypeFilter.value = initialTokenType
    }
  })

  /**
   * Resets both filters to null.
   */
  function clearFilters() {
    chainTypeFilter.value = null
    tokenTypeFilter.value = null
    // Note: Does not update the URL directly. The page component should watch these refs.
  }

  // --- Type Guards (Basic validation) ---
  // TODO: Potentially use enums or a more robust validation source
  function isValidChainType(value: string): value is ChainType {
    return ['ethereum', 'polygon', 'base'].includes(value)
  }

  function isValidTokenType(value: string): value is TokenType {
    return ['erc20', 'erc721', 'erc1155', 'native'].includes(value)
  }

  return {
    chainTypeFilter,
    tokenTypeFilter,
    clearFilters,
  }
} 