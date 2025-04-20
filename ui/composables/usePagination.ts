import { computed } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import type { Ref } from 'vue';

/**
 * Represents the return type of the usePagination composable.
 */
export interface UsePaginationReturn {
  limit: Ref<number>;
  nextToken: Ref<string | undefined>;
  setLimit: (newLimit: number) => void;
  previousPage: () => void;
  nextPage: (newToken?: string) => void;
}

/**
 * Composable for managing token-based pagination state based on route query parameters.
 * 
 * Provides reactive `limit` and `nextToken` derived from the current route's query,
 * along with functions to navigate between pages and change the limit.
 *
 * @param defaultLimit - The default number of items per page (defaults to 10).
 * @returns An object containing reactive `limit`, `nextToken`, and navigation functions.
 */
export default function (defaultLimit: number = 10): UsePaginationReturn {
  const route = useRoute();
  const router = useRouter();

  /**
   * Reactive limit based on route query parameter 'limit'.
   * Ensures the limit is a positive number, defaulting to `defaultLimit`.
   */
  const limit = computed(() => {
    const queryLimit = route.query.limit ? Number(route.query.limit) : defaultLimit;
    // Ensure limit is a positive number
    return isNaN(queryLimit) || queryLimit <= 0 ? defaultLimit : queryLimit; 
  });

  /**
   * Reactive nextToken based on route query parameter 'next_token'.
   */
  const nextToken = computed(() => {
    return typeof route.query.next_token === 'string' ? route.query.next_token : undefined;
  });

  /**
   * Updates the route query to set a new limit and resets pagination.
   * @param newLimit - The new number of items per page.
   */
  const setLimit = (newLimit: number): void => {
    // Ensure the new limit is valid before pushing
    const validLimit = isNaN(newLimit) || newLimit <= 0 ? defaultLimit : newLimit;
    const query: Record<string, any> = { ...route.query, limit: validLimit };
    
    // Remove next_token to start from the beginning
    delete query.next_token;
    
    router.push({ query });
  };

  /**
   * Updates the route query to navigate to the previous page.
   * For token-based pagination, this typically means removing the token
   * to return to the first page, as we can't easily navigate backwards.
   */
  const previousPage = (): void => {
    // In token-based pagination, we typically can only go back to the beginning
    // as the API doesn't provide "previous" tokens
    const query: Record<string, any> = { ...route.query };
    delete query.next_token;
    router.push({ query });
  };

  /**
   * Updates the route query to navigate to the next page using the provided token.
   * @param newToken - The new pagination token received from the API.
   */
  const nextPage = (newToken?: string): void => {
    if (!newToken) return; // Don't navigate if no token is provided
    
    const query: Record<string, any> = { ...route.query, next_token: newToken };
    router.push({ query });
  };

  return {
    limit,
    nextToken,
    setLimit,
    previousPage,
    nextPage,
  };
} 