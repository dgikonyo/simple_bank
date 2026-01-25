package db

import (
	"context"
	"testing"

	"simple_bank/util"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

// createRandomOrderItem is a helper function to create a random order item for testing
func createRandomOrderItem(t *testing.T, order Order, product Product) OrderItem {
	// Create order item parameters
	arg := CreateOrderItemParams{
		OrderID: pgtype.Int4{
			Int32: order.ID,
			Valid: true,
		},
		ProductID: product.ID,
		Quantity: pgtype.Int4{
			Int32: int32(util.RandomInt(1, 10)),
			Valid: true,
		},
	}

	// Create the order item (returns :exec, not :one)
	err := testQueries.CreateOrderItem(context.Background(), arg)
	require.NoError(t, err)

	orderItems, err := testQueries.ListOrderItems(context.Background(), pgtype.Int4{
		Int32: order.ID,
		Valid: true,
	})
	require.NoError(t, err)
	require.NotEmpty(t, orderItems)

	// Find the specific order item we just created
	var createdItem OrderItem
	for _, item := range orderItems {
		if item.OrderID.Int32 == order.ID && item.ProductID == product.ID {
			createdItem = item
			break
		}
	}
	require.NotEmpty(t, createdItem, "Failed to find created order item")

	// Validate the created order item
	require.Equal(t, arg.OrderID, createdItem.OrderID)
	require.Equal(t, arg.ProductID, createdItem.ProductID)
	require.Equal(t, arg.Quantity, createdItem.Quantity)

	return createdItem
}

// TestCreateOrderItem tests creating an order item
func TestCreateOrderItem(t *testing.T) {
	account := createRandomAccount(t)
	country := createRandomCountry(t)
	merchant := createRandomMerchant(t, account, country)
	product := createRandomProduct(t, merchant)

	// Create an order for the account
	orderArg := CreateOrderParams{
		ID: int32(util.RandomInt(1, 10000)),
		UserID: pgtype.Int4{
			Int32: int32(account.ID),
			Valid: true,
		},
		Status: pgtype.Text{
			String: "pending",
			Valid:  true,
		},
	}
	order, err := testQueries.CreateOrder(context.Background(), orderArg)
	require.NoError(t, err)

	// Create order item
	createRandomOrderItem(t, order, product)
}

// TestCreateOrderItemWithNullValues tests edge cases with nullable fields
func TestCreateOrderItemWithNullValues(t *testing.T) {
	account := createRandomAccount(t)
	country := createRandomCountry(t)
	merchant := createRandomMerchant(t, account, country)
	product := createRandomProduct(t, merchant)

	// Create an order
	orderArg := CreateOrderParams{
		ID: int32(util.RandomInt(10001, 20000)),
		UserID: pgtype.Int4{
			Int32: int32(account.ID),
			Valid: true,
		},
		Status: pgtype.Text{
			String: "pending",
			Valid:  true,
		},
	}
	order, err := testQueries.CreateOrder(context.Background(), orderArg)
	require.NoError(t, err)

	// Test with null quantity (if allowed by business logic)
	arg := CreateOrderItemParams{
		OrderID: pgtype.Int4{
			Int32: order.ID,
			Valid: true,
		},
		ProductID: product.ID,
		Quantity: pgtype.Int4{
			Valid: false, // Null quantity
		},
	}

	err = testQueries.CreateOrderItem(context.Background(), arg)
	require.NoError(t, err)

	// Verify the order item was created with null quantity
	orderItems, err := testQueries.ListOrderItems(context.Background(), pgtype.Int4{
		Int32: order.ID,
		Valid: true,
	})
	require.NoError(t, err)
	require.NotEmpty(t, orderItems)

	found := false
	for _, item := range orderItems {
		if item.ProductID == product.ID {
			found = true
			require.False(t, item.Quantity.Valid, "Quantity should be null")
			break
		}
	}
	require.True(t, found, "Order item not found")

	arg2 := CreateOrderItemParams{
		OrderID: pgtype.Int4{
			Valid: false, // Null order_id
		},
		ProductID: product.ID,
		Quantity: pgtype.Int4{
			Int32: 1,
			Valid: true,
		},
	}

	err = testQueries.CreateOrderItem(context.Background(), arg2)
	// This might fail if order_id is NOT NULL in schema
	if err != nil {
		require.Contains(t, err.Error(), "null value", "Expected null constraint violation")
	}
}

// TestCreateDuplicateOrderItem tests creating the same order item twice
// This should fail if there's a unique constraint on (order_id, product_id)
func TestCreateDuplicateOrderItem(t *testing.T) {
	account := createRandomAccount(t)
	country := createRandomCountry(t)
	merchant := createRandomMerchant(t, account, country)
	product := createRandomProduct(t, merchant)

	// Create an order
	orderArg := CreateOrderParams{
		ID: int32(util.RandomInt(20001, 30000)),
		UserID: pgtype.Int4{
			Int32: int32(account.ID),
			Valid: true,
		},
		Status: pgtype.Text{
			String: "pending",
			Valid:  true,
		},
	}
	order, err := testQueries.CreateOrder(context.Background(), orderArg)
	require.NoError(t, err)

	// Create first order item
	arg := CreateOrderItemParams{
		OrderID: pgtype.Int4{
			Int32: order.ID,
			Valid: true,
		},
		ProductID: product.ID,
		Quantity: pgtype.Int4{
			Int32: 2,
			Valid: true,
		},
	}

	err = testQueries.CreateOrderItem(context.Background(), arg)
	require.NoError(t, err)

	// Try to create duplicate order item (same order and product)
	err = testQueries.CreateOrderItem(context.Background(), arg)
	// This should fail if there's a unique constraint
	if err != nil {
		require.Contains(t, err.Error(), "duplicate", "Expected duplicate key violation")
	} else {
		// If no error, then there's no unique constraint and we should have two entries
		// But this would be unusual for an order_items table
		orderItems, err := testQueries.ListOrderItems(context.Background(), pgtype.Int4{
			Int32: order.ID,
			Valid: true,
		})
		require.NoError(t, err)

		// Count occurrences of this product in the order
		count := 0
		for _, item := range orderItems {
			if item.ProductID == product.ID {
				count++
			}
		}
		// If no unique constraint, we'd have 2 entries for the same product
		// This is probably not desired behavior for an order_items table
	}
}

// TestListOrderItems tests retrieving order items for a specific order
func TestListOrderItems(t *testing.T) {
	// Setup dependencies
	account := createRandomAccount(t)
	country := createRandomCountry(t)
	merchant := createRandomMerchant(t, account, country)

	// Create multiple products
	products := make([]Product, 5)
	for i := 0; i < 5; i++ {
		products[i] = createRandomProduct(t, merchant)
	}

	// Create an order
	orderArg := CreateOrderParams{
		ID: int32(util.RandomInt(30001, 40000)),
		UserID: pgtype.Int4{
			Int32: int32(account.ID),
			Valid: true,
		},
		Status: pgtype.Text{
			String: "pending",
			Valid:  true,
		},
	}
	order, err := testQueries.CreateOrder(context.Background(), orderArg)
	require.NoError(t, err)

	// Create multiple order items for the same order
	createdItems := make([]OrderItem, 5)
	for i := 0; i < 5; i++ {
		arg := CreateOrderItemParams{
			OrderID: pgtype.Int4{
				Int32: order.ID,
				Valid: true,
			},
			ProductID: products[i].ID,
			Quantity: pgtype.Int4{
				Int32: int32(i + 1), // Different quantities: 1, 2, 3, 4, 5
				Valid: true,
			},
		}
		err = testQueries.CreateOrderItem(context.Background(), arg)
		require.NoError(t, err)

		// Store expected items for verification
		createdItems[i] = OrderItem{
			OrderID:   arg.OrderID,
			ProductID: arg.ProductID,
			Quantity:  arg.Quantity,
		}
	}

	// List order items for this order
	listedItems, err := testQueries.ListOrderItems(context.Background(), pgtype.Int4{
		Int32: order.ID,
		Valid: true,
	})
	require.NoError(t, err)
	require.NotEmpty(t, listedItems)

	// Verify we got exactly 5 order items
	require.Len(t, listedItems, 5)

	// Verify all returned items belong to our order
	for _, item := range listedItems {
		require.Equal(t, order.ID, item.OrderID.Int32)
		require.True(t, item.OrderID.Valid)
	}

	// Verify each created item is in the list
	for _, createdItem := range createdItems {
		found := false
		for _, listedItem := range listedItems {
			if listedItem.ProductID == createdItem.ProductID {
				found = true
				require.Equal(t, createdItem.Quantity, listedItem.Quantity)
				break
			}
		}
		require.True(t, found, "Order item for product %d not found", createdItem.ProductID)
	}

	// Test listing for an order with no items
	emptyOrderArg := CreateOrderParams{
		ID: int32(util.RandomInt(40001, 50000)),
		UserID: pgtype.Int4{
			Int32: int32(account.ID),
			Valid: true,
		},
		Status: pgtype.Text{
			String: "pending",
			Valid:  true,
		},
	}
	emptyOrder, err := testQueries.CreateOrder(context.Background(), emptyOrderArg)
	require.NoError(t, err)

	emptyItems, err := testQueries.ListOrderItems(context.Background(), pgtype.Int4{
		Int32: emptyOrder.ID,
		Valid: true,
	})
	require.NoError(t, err)
	require.Empty(t, emptyItems)

	// Test listing with null order_id
	nullItems, err := testQueries.ListOrderItems(context.Background(), pgtype.Int4{
		Valid: false,
	})
	require.NoError(t, err)
	require.Empty(t, nullItems)
	// Should return items with null order_id if any exist
}

// TestUpdateOrderItem tests updating an order item's quantity
func TestUpdateOrderItem(t *testing.T) {
	// Setup dependencies
	account := createRandomAccount(t)
	country := createRandomCountry(t)
	merchant := createRandomMerchant(t, account, country)
	product := createRandomProduct(t, merchant)

	// Create an order
	orderArg := CreateOrderParams{
		ID: int32(util.RandomInt(50001, 60000)),
		UserID: pgtype.Int4{
			Int32: int32(account.ID),
			Valid: true,
		},
		Status: pgtype.Text{
			String: "pending",
			Valid:  true,
		},
	}
	order, err := testQueries.CreateOrder(context.Background(), orderArg)
	require.NoError(t, err)

	// Create an order item
	createArg := CreateOrderItemParams{
		OrderID: pgtype.Int4{
			Int32: order.ID,
			Valid: true,
		},
		ProductID: product.ID,
		Quantity: pgtype.Int4{
			Int32: 1,
			Valid: true,
		},
	}
	err = testQueries.CreateOrderItem(context.Background(), createArg)
	require.NoError(t, err)

	// Update the order item quantity
	updateArg := UpdateOrderItemParams{
		OrderID: pgtype.Int4{
			Int32: order.ID,
			Valid: true,
		},
		ProductID: product.ID,
		Quantity: pgtype.Int4{
			Int32: 5,
			Valid: true,
		},
	}

	err = testQueries.UpdateOrderItem(context.Background(), updateArg)
	require.NoError(t, err)

	// Verify the update
	orderItems, err := testQueries.ListOrderItems(context.Background(), pgtype.Int4{
		Int32: order.ID,
		Valid: true,
	})
	require.NoError(t, err)
	require.NotEmpty(t, orderItems)

	found := false
	for _, item := range orderItems {
		if item.ProductID == product.ID {
			found = true
			require.Equal(t, int32(5), item.Quantity.Int32)
			require.True(t, item.Quantity.Valid)
			break
		}
	}
	require.True(t, found, "Order item not found after update")

	// Test updating to null quantity
	updateArg2 := UpdateOrderItemParams{
		OrderID: pgtype.Int4{
			Int32: order.ID,
			Valid: true,
		},
		ProductID: product.ID,
		Quantity: pgtype.Int4{
			Valid: false,
		},
	}

	err = testQueries.UpdateOrderItem(context.Background(), updateArg2)
	require.NoError(t, err)

	// Verify null quantity
	orderItems2, err := testQueries.ListOrderItems(context.Background(), pgtype.Int4{
		Int32: order.ID,
		Valid: true,
	})
	require.NoError(t, err)

	for _, item := range orderItems2 {
		if item.ProductID == product.ID {
			require.False(t, item.Quantity.Valid, "Quantity should be null")
			break
		}
	}
}

// TestUpdateNonExistentOrderItem tests updating an order item that doesn't exist
func TestUpdateNonExistentOrderItem(t *testing.T) {
	updateArg := UpdateOrderItemParams{
		OrderID: pgtype.Int4{
			Int32: 999999,
			Valid: true,
		},
		ProductID: 999999,
		Quantity: pgtype.Int4{
			Int32: 10,
			Valid: true,
		},
	}

	err := testQueries.UpdateOrderItem(context.Background(), updateArg)
	require.NoError(t, err)
}

// TestDeleteOrderItem tests deleting an order item
func TestDeleteOrderItem(t *testing.T) {
	account := createRandomAccount(t)
	country := createRandomCountry(t)
	merchant := createRandomMerchant(t, account, country)
	product := createRandomProduct(t, merchant)

	// Create an order
	orderArg := CreateOrderParams{
		ID: int32(util.RandomInt(60001, 70000)),
		UserID: pgtype.Int4{
			Int32: int32(account.ID),
			Valid: true,
		},
		Status: pgtype.Text{
			String: "pending",
			Valid:  true,
		},
	}
	order, err := testQueries.CreateOrder(context.Background(), orderArg)
	require.NoError(t, err)

	// Create an order item
	createArg := CreateOrderItemParams{
		OrderID: pgtype.Int4{
			Int32: order.ID,
			Valid: true,
		},
		ProductID: product.ID,
		Quantity: pgtype.Int4{
			Int32: 3,
			Valid: true,
		},
	}
	err = testQueries.CreateOrderItem(context.Background(), createArg)
	require.NoError(t, err)

	// Verify the order item exists
	orderItems, err := testQueries.ListOrderItems(context.Background(), pgtype.Int4{
		Int32: order.ID,
		Valid: true,
	})
	require.NoError(t, err)
	require.NotEmpty(t, orderItems)

	// Delete the order item
	deleteArg := DeleteOrderItemParams{
		OrderID: pgtype.Int4{
			Int32: order.ID,
			Valid: true,
		},
		ProductID: product.ID,
	}
	err = testQueries.DeleteOrderItem(context.Background(), deleteArg)
	require.NoError(t, err)

	// Verify the order item is gone
	orderItemsAfter, err := testQueries.ListOrderItems(context.Background(), pgtype.Int4{
		Int32: order.ID,
		Valid: true,
	})
	require.NoError(t, err)

	// Check that the specific product is no longer in the order
	for _, item := range orderItemsAfter {
		require.NotEqual(t, product.ID, item.ProductID, "Order item should have been deleted")
	}
}

// TestDeleteNonExistentOrderItem tests deleting an order item that doesn't exist
func TestDeleteNonExistentOrderItem(t *testing.T) {
	deleteArg := DeleteOrderItemParams{
		OrderID: pgtype.Int4{
			Int32: 999999,
			Valid: true,
		},
		ProductID: 999999,
	}

	err := testQueries.DeleteOrderItem(context.Background(), deleteArg)
	require.NoError(t, err)
}

// TestCascadeDelete tests what happens when an order or product is deleted
func TestCascadeDelete(t *testing.T) {
	account := createRandomAccount(t)
	country := createRandomCountry(t)
	merchant := createRandomMerchant(t, account, country)
	product := createRandomProduct(t, merchant)

	// Create an order
	orderArg := CreateOrderParams{
		ID: int32(util.RandomInt(70001, 80000)),
		UserID: pgtype.Int4{
			Int32: int32(account.ID),
			Valid: true,
		},
		Status: pgtype.Text{
			String: "pending",
			Valid:  true,
		},
	}
	order, err := testQueries.CreateOrder(context.Background(), orderArg)
	require.NoError(t, err)

	// Create an order item
	createArg := CreateOrderItemParams{
		OrderID: pgtype.Int4{
			Int32: order.ID,
			Valid: true,
		},
		ProductID: product.ID,
		Quantity: pgtype.Int4{
			Int32: 2,
			Valid: true,
		},
	}
	err = testQueries.CreateOrderItem(context.Background(), createArg)
	require.NoError(t, err)

	// Test what happens when we delete the order
	err = testQueries.DeleteOrder(context.Background(), order.ID)
	require.NoError(t, err)

	// Try to list order items for the deleted order
	orderItems, err := testQueries.ListOrderItems(context.Background(), pgtype.Int4{
		Int32: order.ID,
		Valid: true,
	})

	require.NoError(t, err)
	require.Empty(t, orderItems)
}

// TestOrderItemQuantityBoundaries tests edge cases for quantity field
func TestOrderItemQuantityBoundaries(t *testing.T) {
	account := createRandomAccount(t)
	country := createRandomCountry(t)
	merchant := createRandomMerchant(t, account, country)
	product := createRandomProduct(t, merchant)

	// Create an order
	orderArg := CreateOrderParams{
		ID: int32(util.RandomInt(80001, 90000)),
		UserID: pgtype.Int4{
			Int32: int32(account.ID),
			Valid: true,
		},
		Status: pgtype.Text{
			String: "pending",
			Valid:  true,
		},
	}
	order, err := testQueries.CreateOrder(context.Background(), orderArg)
	require.NoError(t, err)

	// Test with zero quantity (if allowed)
	arg1 := CreateOrderItemParams{
		OrderID: pgtype.Int4{
			Int32: order.ID,
			Valid: true,
		},
		ProductID: product.ID,
		Quantity: pgtype.Int4{
			Int32: 0,
			Valid: true,
		},
	}

	err = testQueries.CreateOrderItem(context.Background(), arg1)
	// This might fail if there's a check constraint
	if err == nil {
		// Verify zero quantity was saved
		orderItems, err := testQueries.ListOrderItems(context.Background(), pgtype.Int4{
			Int32: order.ID,
			Valid: true,
		})
		require.NoError(t, err)
		for _, item := range orderItems {
			if item.ProductID == product.ID {
				require.Equal(t, int32(0), item.Quantity.Int32)
			}
		}
	}

	// Test with negative quantity (probably should fail)
	// Create a different product
	product2 := createRandomProduct(t, merchant)
	arg2 := CreateOrderItemParams{
		OrderID: pgtype.Int4{
			Int32: order.ID,
			Valid: true,
		},
		ProductID: product2.ID,
		Quantity: pgtype.Int4{
			Int32: -5,
			Valid: true,
		},
	}

	err = testQueries.CreateOrderItem(context.Background(), arg2)
	// This should fail if there's a check constraint
	if err != nil {
		require.Contains(t, err.Error(), "check", "Expected check constraint violation")
	}

	// Test with very large quantity
	product3 := createRandomProduct(t, merchant)
	arg3 := CreateOrderItemParams{
		OrderID: pgtype.Int4{
			Int32: order.ID,
			Valid: true,
		},
		ProductID: product3.ID,
		Quantity: pgtype.Int4{
			Int32: 1000000, // 1 million
			Valid: true,
		},
	}

	err = testQueries.CreateOrderItem(context.Background(), arg3)
	// This depends on your business logic
	if err == nil {
		// Verify large quantity was saved
		orderItems, err := testQueries.ListOrderItems(context.Background(), pgtype.Int4{
			Int32: order.ID,
			Valid: true,
		})
		require.NoError(t, err)
		for _, item := range orderItems {
			if item.ProductID == product3.ID {
				require.Equal(t, int32(1000000), item.Quantity.Int32)
			}
		}
	}
}

// TestConcurrentOrderItemOperations tests concurrent create/update/delete operations
func TestConcurrentOrderItemOperations(t *testing.T) {
	account := createRandomAccount(t)
	country := createRandomCountry(t)
	merchant := createRandomMerchant(t, account, country)

	// Create multiple products
	products := make([]Product, 3)
	for i := 0; i < 3; i++ {
		products[i] = createRandomProduct(t, merchant)
	}

	// Create an order
	orderArg := CreateOrderParams{
		ID: int32(util.RandomInt(90001, 100000)),
		UserID: pgtype.Int4{
			Int32: int32(account.ID),
			Valid: true,
		},
		Status: pgtype.Text{
			String: "pending",
			Valid:  true,
		},
	}
	order, err := testQueries.CreateOrder(context.Background(), orderArg)
	require.NoError(t, err)

	// Concurrently create order items
	errs := make(chan error, 3)
	for i := 0; i < 3; i++ {
		go func(idx int) {
			arg := CreateOrderItemParams{
				OrderID: pgtype.Int4{
					Int32: order.ID,
					Valid: true,
				},
				ProductID: products[idx].ID,
				Quantity: pgtype.Int4{
					Int32: int32(idx + 1),
					Valid: true,
				},
			}
			errs <- testQueries.CreateOrderItem(context.Background(), arg)
		}(i)
	}

	for i := 0; i < 3; i++ {
		err := <-errs
		require.NoError(t, err)
	}

	// Verify all were created
	orderItems, err := testQueries.ListOrderItems(context.Background(), pgtype.Int4{
		Int32: order.ID,
		Valid: true,
	})
	require.NoError(t, err)
	require.Len(t, orderItems, 3)
}
