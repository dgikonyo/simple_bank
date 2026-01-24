package db

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
	"simple_bank/util"
)

// createRandomOrder is a helper function to create a random order for testing
func createRandomOrder(t *testing.T, account Account) Order {
	// Generate random order parameters
	arg := CreateOrderParams{
		ID: int32(util.RandomInt(1, 10000)), 
		UserID: pgtype.Int4{
			Int32: int32(account.ID),
			Valid: true,
		},
		Status: pgtype.Text{
			String: util.RandomOrderStatus(),
			Valid:  true,
		},
	}

	order, err := testQueries.CreateOrder(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, order)

	// Validate the returned order matches input
	require.Equal(t, arg.ID, order.ID)
	require.Equal(t, arg.UserID, order.UserID)
	require.Equal(t, arg.Status, order.Status)

	// Ensure system-generated fields are set
	require.NotZero(t, order.CreatedAt)

	return order
}

// TestCreateOrder tests the CreateOrder function
func TestCreateOrder(t *testing.T) {
	account := createRandomAccount(t)
	createRandomOrder(t, account)
}

// TestCreateOrderWithNullValues tests edge cases with nullable fields
func TestCreateOrderWithNullValues(t *testing.T) {
	// Test with null user_id
	arg1 := CreateOrderParams{
		ID: int32(util.RandomInt(1, 10000)),
		UserID: pgtype.Int4{
			Valid: false, // Null user_id
		},
		Status: pgtype.Text{
			String: "pending",
			Valid:  true,
		},
	}

	order1, err := testQueries.CreateOrder(context.Background(), arg1)
	require.NoError(t, err)
	require.NotEmpty(t, order1)
	require.False(t, order1.UserID.Valid) // Should be null
	require.Equal(t, "pending", order1.Status.String)

	// Test with null status
	account := createRandomAccount(t)
	arg2 := CreateOrderParams{
		ID: int32(util.RandomInt(1, 10000)),
		UserID: pgtype.Int4{
			Int32: int32(account.ID),
			Valid: true,
		},
		Status: pgtype.Text{
			Valid: false, // Null status
		},
	}

	order2, err := testQueries.CreateOrder(context.Background(), arg2)
	require.NoError(t, err)
	require.NotEmpty(t, order2)
	require.True(t, order2.UserID.Valid)
	require.False(t, order2.Status.Valid) // Should be null
}

// TestGetOrder tests retrieving an order by ID
func TestGetOrder(t *testing.T) {
	account := createRandomAccount(t)
	initialOrder := createRandomOrder(t, account)

	// Retrieve the order
	retrievedOrder, err := testQueries.GetOrder(context.Background(), initialOrder.ID)
	require.NoError(t, err)
	require.NotEmpty(t, retrievedOrder)

	// Validate all fields match
	require.Equal(t, initialOrder.ID, retrievedOrder.ID)
	require.Equal(t, initialOrder.UserID, retrievedOrder.UserID)
	require.Equal(t, initialOrder.Status, retrievedOrder.Status)
	require.WithinDuration(t, initialOrder.CreatedAt.Time, retrievedOrder.CreatedAt.Time, time.Second)
}

// TestGetOrderNotFound tests retrieving a non-existent order
func TestGetOrderNotFound(t *testing.T) {
	nonExistentID := int32(999999)
	order, err := testQueries.GetOrder(context.Background(), nonExistentID)
	require.Error(t, err)
	require.ErrorIs(t, err, pgx.ErrNoRows)
	require.Empty(t, order)
}

// TestDeleteOrder tests deleting an order
func TestDeleteOrder(t *testing.T) {
	account := createRandomAccount(t)
	targetOrder := createRandomOrder(t, account)

	// Delete the order
	err := testQueries.DeleteOrder(context.Background(), targetOrder.ID)
	require.NoError(t, err)

	// Verify it's deleted
	deletedOrder, err := testQueries.GetOrder(context.Background(), targetOrder.ID)
	require.Error(t, err)
	require.ErrorIs(t, err, pgx.ErrNoRows)
	require.Empty(t, deletedOrder)
}

// TestDeleteNonExistentOrder tests deleting an order that doesn't exist
func TestDeleteNonExistentOrder(t *testing.T) {
	nonExistentID := int32(999999)
	err := testQueries.DeleteOrder(context.Background(), nonExistentID)
	require.NoError(t, err) 
}

// TestListOrdersByUser tests retrieving orders for a specific user
func TestListOrdersByUser(t *testing.T) {
	// Create two users
	user1 := createRandomAccount(t)
	user2 := createRandomAccount(t)

	// Create orders for user1
	var user1Orders []Order
	for i := 0; i < 5; i++ {
		order := createRandomOrder(t, user1)
		user1Orders = append(user1Orders, order)
	}

	// Create orders for user2
	for i := 0; i < 3; i++ {
		createRandomOrder(t, user2)
	}

	// List orders for user1
	arg := pgtype.Int4{
		Int32: int32(user1.ID),
		Valid: true,
	}
	listedOrders, err := testQueries.ListOrdersByUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, listedOrders)

	// Verify we got exactly 5 orders for user1
	require.Len(t, listedOrders, 5)

	// Verify all returned orders belong to user1
	for _, order := range listedOrders {
		require.Equal(t, int32(user1.ID), order.UserID.Int32)
		require.True(t, order.UserID.Valid)
	}

	// Verify orders are sorted by created_at DESC (newest first)
	for i := 0; i < len(listedOrders)-1; i++ {
		require.True(t, listedOrders[i].CreatedAt.Time.After(listedOrders[i+1].CreatedAt.Time) ||
			listedOrders[i].CreatedAt.Time.Equal(listedOrders[i+1].CreatedAt.Time))
	}

	// Verify our created orders are in the list
	foundCount := 0
	for _, createdOrder := range user1Orders {
		for _, listedOrder := range listedOrders {
			if createdOrder.ID == listedOrder.ID {
				foundCount++
				break
			}
		}
	}
	require.Equal(t, 5, foundCount)
}

// TestListOrdersByUserWithNoOrders tests listing orders for a user with no orders
func TestListOrdersByUserWithNoOrders(t *testing.T) {
	user := createRandomAccount(t)
	arg := pgtype.Int4{
		Int32: int32(user.ID),
		Valid: true,
	}

	orders, err := testQueries.ListOrdersByUser(context.Background(), arg)
	require.NoError(t, err)
	require.Empty(t, orders)
}

// TestListOrdersByUserNullUserID tests listing orders with null user_id
func TestListOrdersByUserNullUserID(t *testing.T) {
	// Create some orders with null user_id
	for i := 0; i < 3; i++ {
		arg := CreateOrderParams{
			ID: int32(util.RandomInt(10001, 20000)),
			UserID: pgtype.Int4{
				Valid: false,
			},
			Status: pgtype.Text{
				String: "pending",
				Valid:  true,
			},
		}
		_, err := testQueries.CreateOrder(context.Background(), arg)
		require.NoError(t, err)
	}

	// Try to list orders with null user_id
	arg := pgtype.Int4{
		Valid: false,
	}
	orders, err := testQueries.ListOrdersByUser(context.Background(), arg)
	require.NoError(t, err)
	
	// Should return orders with null user_id
	for _, order := range orders {
		require.False(t, order.UserID.Valid)
	}
}

// TestUpdateOrderStatus tests updating an order's status
func TestUpdateOrderStatus(t *testing.T) {
	account := createRandomAccount(t)
	targetOrder := createRandomOrder(t, account)

	// Update the order status
	newStatus := "shipped"
	arg := UpdateOrderStatusParams{
		ID: targetOrder.ID,
		Status: pgtype.Text{
			String: newStatus,
			Valid:  true,
		},
	}

	err := testQueries.UpdateOrderStatus(context.Background(), arg)
	require.NoError(t, err)

	// Retrieve and verify the update
	updatedOrder, err := testQueries.GetOrder(context.Background(), targetOrder.ID)
	require.NoError(t, err)
	require.NotEmpty(t, updatedOrder)

	require.Equal(t, targetOrder.ID, updatedOrder.ID)
	require.Equal(t, targetOrder.UserID, updatedOrder.UserID)
	require.Equal(t, newStatus, updatedOrder.Status.String)
	require.True(t, updatedOrder.Status.Valid)
	require.WithinDuration(t, targetOrder.CreatedAt.Time, updatedOrder.CreatedAt.Time, time.Second)
}

// TestUpdateOrderStatusToNull tests updating status to null
func TestUpdateOrderStatusToNull(t *testing.T) {
	account := createRandomAccount(t)
	targetOrder := createRandomOrder(t, account)

	// Update status to null
	arg := UpdateOrderStatusParams{
		ID: targetOrder.ID,
		Status: pgtype.Text{
			Valid: false, // Null status
		},
	}

	err := testQueries.UpdateOrderStatus(context.Background(), arg)
	require.NoError(t, err)

	// Verify update
	updatedOrder, err := testQueries.GetOrder(context.Background(), targetOrder.ID)
	require.NoError(t, err)
	require.False(t, updatedOrder.Status.Valid) // Status should be null
}

// TestUpdateNonExistentOrder tests updating an order that doesn't exist
func TestUpdateNonExistentOrder(t *testing.T) {
	arg := UpdateOrderStatusParams{
		ID: int32(999999),
		Status: pgtype.Text{
			String: "cancelled",
			Valid:  true,
		},
	}

	err := testQueries.UpdateOrderStatus(context.Background(), arg)
	require.NoError(t, err) // Updating non-existent row should not error
}

// TestConcurrentOrderCreation tests creating orders concurrently
func TestConcurrentOrderCreation(t *testing.T) {
	account := createRandomAccount(t)
	
	// Number of concurrent goroutines
	const n = 10
	errs := make(chan error, n)
	
	for i := 0; i < n; i++ {
		go func(id int) {
			arg := CreateOrderParams{
				ID: int32(id + 1), // Ensure unique IDs
				UserID: pgtype.Int4{
					Int32: int32(account.ID),
					Valid: true,
				},
				Status: pgtype.Text{
					String: "pending",
					Valid:  true,
				},
			}
			_, err := testQueries.CreateOrder(context.Background(), arg)
			errs <- err
		}(i)
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)
	}
	
	// Verify all orders were created
	arg := pgtype.Int4{
		Int32: int32(account.ID),
		Valid: true,
	}
	orders, err := testQueries.ListOrdersByUser(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, orders, n)
}