/** Centralized query keys for the inventory feature. */
export const inventoryQueryKeys = {
  stock: (itemId: string) => ["inventory", "stock", itemId] as const,
};
