import { computed } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import type { Ref } from 'vue';

/**
 * Represents the return type of the usePagination composable.
 */
export interface UsePaginationReturn {
  limit: Ref<number>;
  offset: Ref<number>;
  setLimit: (newLimit: number) => void;
  previousPage: () => void;
  nextPage: () => void;
}

/**
 * Composable for managing pagination state based on route query parameters.
 * 
 * Provides reactive `limit` and `offset` derived from the current route's query,
 * along with functions to navigate between pages and change the limit.
 *
 * @param defaultLimit - The default number of items per page (defaults to 10).
 * @returns An object containing reactive `limit`, `offset`, and navigation functions.
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
   * Reactive offset based on route query parameter 'offset'.
   * Ensures the offset is a non-negative number, defaulting to 0.
   */
  const offset = computed(() => {
    const queryOffset = route.query.offset ? Number(route.query.offset) : 0;
    // Ensure offset is non-negative
    return isNaN(queryOffset) || queryOffset < 0 ? 0 : queryOffset; 
  });

  /**
   * Updates the route query to set a new limit and resets the offset.
   * @param newLimit - The new number of items per page.
   */
  const setLimit = (newLimit: number): void => {
    // Ensure the new limit is valid before pushing
    const validLimit = isNaN(newLimit) || newLimit <= 0 ? defaultLimit : newLimit;
    router.push({ query: { ...route.query, limit: validLimit, offset: 0 } });
  };

  /**
   * Updates the route query to navigate to the previous page.
   * Calculates the new offset, ensuring it doesn't go below 0.
   */
  const previousPage = (): void => {
    const newOffset = Math.max(0, offset.value - limit.value);
    router.push({ query: { ...route.query, offset: newOffset } });
  };

  /**
   * Updates the route query to navigate to the next page.
   * Calculates the new offset based on the current limit.
   */
  const nextPage = (): void => {
    const newOffset = offset.value + limit.value;
    router.push({ query: { ...route.query, offset: newOffset } });
  };

  return {
    limit,
    offset,
    setLimit,
    previousPage,
    nextPage,
  };
} 