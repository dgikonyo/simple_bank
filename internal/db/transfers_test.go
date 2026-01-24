package db

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/jackc/pgx/v5"
	"simple_bank/util"
)

// createRandomTransfer is a helper function to create a random transfer for testing
func createRandomTransfer(t *testing.T) (Transfer, Account, Account) {
	// Create two random accounts for transfer
	fromAccount := createRandomAccount(t)
	toAccount := createRandomAccount(t)
	
	// Ensure accounts are different (important for transfers)
	require.NotEqual(t, fromAccount.ID, toAccount.ID)
	
	// Generate random transfer parameters
	arg := CreateTransferParams{
		FromAccountID: fromAccount.ID,
		ToAccountID:   toAccount.ID,
		Amount:        util.RandomMoney(), // Assume this generates positive amounts
	}

	transfer, err := testQueries.CreateTransfer(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, transfer)

	// Validate the returned transfer matches input
	require.Equal(t, arg.FromAccountID, transfer.FromAccountID)
	require.Equal(t, arg.ToAccountID, transfer.ToAccountID)
	require.Equal(t, arg.Amount, transfer.Amount)

	// Ensure system-generated fields are set
	require.NotZero(t, transfer.ID)
	require.NotZero(t, transfer.CreatedAt)

	return transfer, fromAccount, toAccount
}

// TestCreateTransfer tests basic transfer creation
func TestCreateTransfer(t *testing.T) {
	createRandomTransfer(t)
}

// TestCreateTransferSameAccount tests transferring to same account
func TestCreateTransferSameAccount(t *testing.T) {
	account := createRandomAccount(t)
	
	arg := CreateTransferParams{
		FromAccountID: account.ID,
		ToAccountID:   account.ID, // Same account
		Amount:        util.RandomMoney(),
	}

	transfer, err := testQueries.CreateTransfer(context.Background(), arg)
	
	if err != nil {
		require.Contains(t, err.Error(), "check", "Expected constraint violation")
	} else {
		require.NotEmpty(t, transfer)
		require.Equal(t, account.ID, transfer.FromAccountID)
		require.Equal(t, account.ID, transfer.ToAccountID)
	}
}

// TestCreateTransferWithZeroAmount tests transfers with zero amount
func TestCreateTransferWithZeroAmount(t *testing.T) {
	fromAccount := createRandomAccount(t)
	toAccount := createRandomAccount(t)
	
	arg := CreateTransferParams{
		FromAccountID: fromAccount.ID,
		ToAccountID:   toAccount.ID,
		Amount:        0, // Zero amount
	}

	transfer, err := testQueries.CreateTransfer(context.Background(), arg)
	if err != nil {
		require.Contains(t, err.Error(), "check", "Expected check constraint for positive amount")
	} else {
		require.NotEmpty(t, transfer)
		require.Equal(t, int64(0), transfer.Amount)
	}
}

// TestCreateTransferWithNegativeAmount tests transfers with negative amount
func TestCreateTransferWithNegativeAmount(t *testing.T) {
	fromAccount := createRandomAccount(t)
	toAccount := createRandomAccount(t)
	
	arg := CreateTransferParams{
		FromAccountID: fromAccount.ID,
		ToAccountID:   toAccount.ID,
		Amount:        -100, // Negative amount
	}

	_, err := testQueries.CreateTransfer(context.Background(), arg)
	// This should definitely fail - transfers should have positive amounts
	require.Error(t, err)
	require.Contains(t, err.Error(), "check", "Expected check constraint for positive amount")
}

// TestCreateTransferLargeAmount tests with very large amounts
func TestCreateTransferLargeAmount(t *testing.T) {
	fromAccount := createRandomAccount(t)
	toAccount := createRandomAccount(t)
	
	arg := CreateTransferParams{
		FromAccountID: fromAccount.ID,
		ToAccountID:   toAccount.ID,
		Amount:        1_000_000_000_000, // 1 trillion
	}

	transfer, err := testQueries.CreateTransfer(context.Background(), arg)
	// This tests integer limits and business logic
	if err != nil {
		require.Contains(t, err.Error(), "balance", "Expected insufficient balance error")
	} else {
		require.NotEmpty(t, transfer)
		require.Equal(t, int64(1_000_000_000_000), transfer.Amount)
	}
}

// TestGetTransfer tests retrieving a transfer by ID
func TestGetTransfer(t *testing.T) {
	initialTransfer, fromAccount, toAccount := createRandomTransfer(t)

	retrievedTransfer, err := testQueries.GetTransfer(context.Background(), initialTransfer.ID)
	require.NoError(t, err)
	require.NotEmpty(t, retrievedTransfer)

	require.Equal(t, initialTransfer.ID, retrievedTransfer.ID)
	require.Equal(t, initialTransfer.FromAccountID, retrievedTransfer.FromAccountID)
	require.Equal(t, initialTransfer.ToAccountID, retrievedTransfer.ToAccountID)
	require.Equal(t, initialTransfer.Amount, retrievedTransfer.Amount)
	require.WithinDuration(t, initialTransfer.CreatedAt.Time, retrievedTransfer.CreatedAt.Time, time.Second)
	
	require.Equal(t, fromAccount.ID, retrievedTransfer.FromAccountID)
	require.Equal(t, toAccount.ID, retrievedTransfer.ToAccountID)
}

// TestGetTransferNotFound tests retrieving a non-existent transfer
func TestGetTransferNotFound(t *testing.T) {
	nonExistentID := int64(999999)
	transfer, err := testQueries.GetTransfer(context.Background(), nonExistentID)
	require.Error(t, err)
	require.ErrorIs(t, err, pgx.ErrNoRows)
	require.Empty(t, transfer)
}

// TestDeleteTransfer tests deleting a transfer
func TestDeleteTransfer(t *testing.T) {
	transfer, _, _ := createRandomTransfer(t)

	// Delete the transfer
	err := testQueries.DeleteTransfer(context.Background(), transfer.ID)
	require.NoError(t, err)

	// Verify it's deleted
	deletedTransfer, err := testQueries.GetTransfer(context.Background(), transfer.ID)
	require.Error(t, err)
	require.ErrorIs(t, err, pgx.ErrNoRows)
	require.Empty(t, deletedTransfer)
}

// TestDeleteTransferCascadeOrConstraint tests what happens when an account is deleted
func TestDeleteTransferCascadeOrConstraint(t *testing.T) {
	transfer, fromAccount, _ := createRandomTransfer(t)

	err := testQueries.DeleteAccount(context.Background(), fromAccount.ID)

	if err == nil {
		_, err2 := testQueries.GetTransfer(context.Background(), transfer.ID)
		if err2 != nil {
			// Transfer was deleted (CASCADE)
			require.ErrorIs(t, err2, pgx.ErrNoRows)
		}
	} else {
		// Account deletion failed (RESTRICT)
		require.Contains(t, err.Error(), "foreign", "Expected foreign key constraint violation")
	}
}

// TestListTransfers tests retrieving all transfers
func TestListTransfers(t *testing.T) {
	transfers := make([]Transfer, 5)
	for i := 0; i < 5; i++ {
		transfer, _, _ := createRandomTransfer(t)
		transfers[i] = transfer
	}

	// List all transfers
	listedTransfers, err := testQueries.ListTransfers(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, listedTransfers)

	// We should have at least 5 transfers 
	require.GreaterOrEqual(t, len(listedTransfers), 5)

	// Verify transfers are sorted by created_at DESC
	for i := 0; i < len(listedTransfers)-1; i++ {
		require.True(t, 
			listedTransfers[i].CreatedAt.Time.After(listedTransfers[i+1].CreatedAt.Time) ||
			listedTransfers[i].CreatedAt.Time.Equal(listedTransfers[i+1].CreatedAt.Time),
			"Transfers not sorted by created_at DESC",
		)
	}

	// Verify our created transfers are in the list
	foundCount := 0
	for _, createdTransfer := range transfers {
		for _, listedTransfer := range listedTransfers {
			if createdTransfer.ID == listedTransfer.ID {
				foundCount++
				break
			}
		}
	}
	require.Equal(t, 5, foundCount, "Not all created transfers found in list")
}

// TestListTransfersEmpty tests listing transfers when none exist
func TestListTransfersEmpty(t *testing.T) {
	listedTransfers, err := testQueries.ListTransfers(context.Background())
	require.NoError(t, err)
	require.Empty(t, listedTransfers)
}

// TestTransferIntegrity tests foreign key constraints
func TestTransferIntegrity(t *testing.T) {
	arg1 := CreateTransferParams{
		FromAccountID: 999999, // Non-existent account
		ToAccountID:   createRandomAccount(t).ID,
		Amount:        100,
	}

	_, err := testQueries.CreateTransfer(context.Background(), arg1)
	require.Error(t, err)
	require.Contains(t, err.Error(), "foreign", "Expected foreign key constraint violation for from_account_id")

	// Test with non-existent to_account_id
	arg2 := CreateTransferParams{
		FromAccountID: createRandomAccount(t).ID,
		ToAccountID:   999999, // Non-existent account
		Amount:        100,
	}

	_, err = testQueries.CreateTransfer(context.Background(), arg2)
	require.Error(t, err)
	require.Contains(t, err.Error(), "foreign", "Expected foreign key constraint violation for to_account_id")
}

// TestConcurrentTransfers tests concurrent transfer creation
func TestConcurrentTransfers(t *testing.T) {
	fromAccount := createRandomAccount(t)
	toAccount := createRandomAccount(t)
	
	// Number of concurrent goroutines
	const n = 10
	errs := make(chan error, n)
	
	for i := 0; i < n; i++ {
		go func(id int) {
			arg := CreateTransferParams{
				FromAccountID: fromAccount.ID,
				ToAccountID:   toAccount.ID,
				Amount:        int64(10 + id), // Different amounts
			}
			_, err := testQueries.CreateTransfer(context.Background(), arg)
			errs <- err
		}(i)
	}
	
	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)
	}
	
	// Verify all transfers were created
	listedTransfers, err := testQueries.ListTransfers(context.Background())
	require.NoError(t, err)
	
	// Count transfers between these two accounts
	count := 0
	for _, transfer := range listedTransfers {
		if transfer.FromAccountID == fromAccount.ID && 
		   transfer.ToAccountID == toAccount.ID {
			count++
		}
	}
	require.GreaterOrEqual(t, count, n)
}

// TestTransferAmountPrecision tests handling of amount precision
func TestTransferAmountPrecision(t *testing.T) {
	fromAccount := createRandomAccount(t)
	toAccount := createRandomAccount(t)
	
	testCases := []struct {
		name   string
		amount int64
		shouldFail bool
	}{
		{"Single cent", 1, false},
		{"Dollar amount", 100, false},
		{"Large amount", 9999999999, false},
		{"Zero amount", 0, true}, // Business rule: likely invalid
		{"Negative amount", -1, true},
		{"Very large amount", 9223372036854775807, false}, // Max int64
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			arg := CreateTransferParams{
				FromAccountID: fromAccount.ID,
				ToAccountID:   toAccount.ID,
				Amount:        tc.amount,
			}
			
			transfer, err := testQueries.CreateTransfer(context.Background(), arg)
			
			if tc.shouldFail {
				require.Error(t, err)
				// Check for specific constraint violations
				if tc.amount < 0 {
					require.Contains(t, err.Error(), "check", "Expected check constraint for positive amount")
				}
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, transfer)
				require.Equal(t, tc.amount, transfer.Amount)
				
				// Clean up for next test
				err = testQueries.DeleteTransfer(context.Background(), transfer.ID)
				require.NoError(t, err)
			}
		})
	}
}

// TestTransferAuditTrail tests that created_at is properly set
func TestTransferAuditTrail(t *testing.T) {
	transfer, _, _ := createRandomTransfer(t)
	
	// Verify created_at is recent (within last minute)
	require.WithinDuration(t, time.Now(), transfer.CreatedAt.Time, time.Minute)
	
	// Verify created_at doesn't change on subsequent retrieval
	retrievedTransfer, err := testQueries.GetTransfer(context.Background(), transfer.ID)
	require.NoError(t, err)
	require.Equal(t, transfer.CreatedAt.Time.Unix(), retrievedTransfer.CreatedAt.Time.Unix())
}

// TestTransferIsolation tests that transfers don't interfere with each other
func TestTransferIsolation(t *testing.T) {
	// Create two independent transfers
	transfer1, acc1, acc2 := createRandomTransfer(t)
	transfer2, acc3, acc4 := createRandomTransfer(t)
	
	// Ensure all transfers are distinct
	require.NotEqual(t, transfer1.ID, transfer2.ID)
	require.NotEqual(t, acc1.ID, acc3.ID) // Different accounts
	require.NotEqual(t, acc2.ID, acc4.ID)
	
	// Verify each transfer can be retrieved independently
	retrieved1, err := testQueries.GetTransfer(context.Background(), transfer1.ID)
	require.NoError(t, err)
	require.Equal(t, transfer1.FromAccountID, retrieved1.FromAccountID)
	
	retrieved2, err := testQueries.GetTransfer(context.Background(), transfer2.ID)
	require.NoError(t, err)
	require.Equal(t, transfer2.FromAccountID, retrieved2.FromAccountID)
	
	// Delete one transfer, other should remain
	err = testQueries.DeleteTransfer(context.Background(), transfer1.ID)
	require.NoError(t, err)
	
	_, err = testQueries.GetTransfer(context.Background(), transfer1.ID)
	require.Error(t, err)
	
	_, err = testQueries.GetTransfer(context.Background(), transfer2.ID)
	require.NoError(t, err)
}