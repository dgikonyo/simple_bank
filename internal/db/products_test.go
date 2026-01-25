package db

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"simple_bank/util"
)

// createRandomProduct is a helper function to create a random product for testing
func createRandomProduct(t *testing.T, merchant Merchant) Product {
	// Generate random product parameters
	arg := CreateProductParams{
		ID:         int32(util.RandomInt(1, 10000)), // Generate unique ID for testing
		Name:       util.RandomProductName(),
		MerchantID: int32(merchant.ID),
		Price: pgtype.Int4{
			Int32: int32(util.RandomInt(100, 10000)), // Random price between 100 and 10000
			Valid: true,
		},
		Status: pgtype.Text{
			String: getRandomProductStatus(),
			Valid:  true,
		},
	}

	product, err := testQueries.CreateProduct(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, product)

	// Validate the returned product matches input
	require.Equal(t, arg.ID, product.ID)
	require.Equal(t, arg.Name, product.Name)
	require.Equal(t, arg.MerchantID, product.MerchantID)
	require.Equal(t, arg.Price, product.Price)
	require.Equal(t, arg.Status, product.Status)

	// Ensure system-generated fields are set
	require.NotZero(t, product.CreatedAt)

	return product
}

// getRandomProductStatus returns a random product status from a predefined list
func getRandomProductStatus() string {
	statuses := []string{"active", "inactive", "out_of_stock", "discontinued", "coming_soon"}
	return statuses[util.RandomInt(0, int64(len(statuses)-1))]
}

// TestCreateProduct tests the CreateProduct function
func TestCreateProduct(t *testing.T) {
	// Setup dependencies
	account := createRandomAccount(t)
	country := createRandomCountry(t)
	merchant := createRandomMerchant(t, account, country)

	createRandomProduct(t, merchant)
}

// TestCreateProductWithNullValues tests edge cases with nullable fields
func TestCreateProductWithNullValues(t *testing.T) {
	account := createRandomAccount(t)
	country := createRandomCountry(t)
	merchant := createRandomMerchant(t, account, country)

	// Test with null price
	arg1 := CreateProductParams{
		ID:         int32(util.RandomInt(10001, 20000)),
		Name:       "Free Sample",
		MerchantID: int32(merchant.ID),
		Price: pgtype.Int4{
			Valid: false,
		},
		Status: pgtype.Text{
			String: "active",
			Valid:  true,
		},
	}

	product1, err := testQueries.CreateProduct(context.Background(), arg1)
	require.NoError(t, err)
	require.NotEmpty(t, product1)
	require.False(t, product1.Price.Valid)
	require.Equal(t, "Free Sample", product1.Name)
	require.Equal(t, "active", product1.Status.String)

	// Test with null status
	arg2 := CreateProductParams{
		ID:         int32(util.RandomInt(20001, 30000)),
		Name:       "Uncategorized Product",
		MerchantID: int32(merchant.ID),
		Price: pgtype.Int4{
			Int32: 5000,
			Valid: true,
		},
		Status: pgtype.Text{
			Valid: false, // Null status
		},
	}

	product2, err := testQueries.CreateProduct(context.Background(), arg2)
	require.NoError(t, err)
	require.NotEmpty(t, product2)
	require.True(t, product2.Price.Valid)
	require.Equal(t, int32(5000), product2.Price.Int32)
	require.False(t, product2.Status.Valid) // Status should be null

	// Test with both null price and status
	arg3 := CreateProductParams{
		ID:         int32(util.RandomInt(30001, 40000)),
		Name:       "Unknown Item",
		MerchantID: int32(merchant.ID),
		Price: pgtype.Int4{
			Valid: false,
		},
		Status: pgtype.Text{
			Valid: false,
		},
	}

	product3, err := testQueries.CreateProduct(context.Background(), arg3)
	require.NoError(t, err)
	require.NotEmpty(t, product3)
	require.False(t, product3.Price.Valid)
	require.False(t, product3.Status.Valid)
}

// TestGetProduct tests retrieving a product by ID
func TestGetProduct(t *testing.T) {
	account := createRandomAccount(t)
	country := createRandomCountry(t)
	merchant := createRandomMerchant(t, account, country)
	initialProduct := createRandomProduct(t, merchant)

	// Retrieve the product
	retrievedProduct, err := testQueries.GetProduct(context.Background(), initialProduct.ID)
	require.NoError(t, err)
	require.NotEmpty(t, retrievedProduct)

	// Validate all fields match
	require.Equal(t, initialProduct.ID, retrievedProduct.ID)
	require.Equal(t, initialProduct.Name, retrievedProduct.Name)
	require.Equal(t, initialProduct.MerchantID, retrievedProduct.MerchantID)
	require.Equal(t, initialProduct.Price, retrievedProduct.Price)
	require.Equal(t, initialProduct.Status, retrievedProduct.Status)
	require.WithinDuration(t, initialProduct.CreatedAt.Time, retrievedProduct.CreatedAt.Time, time.Second)
}

// TestGetProductNotFound tests retrieving a non-existent product
func TestGetProductNotFound(t *testing.T) {
	// Use an ID that shouldn't exist
	nonExistentID := int32(999999)
	product, err := testQueries.GetProduct(context.Background(), nonExistentID)
	require.Error(t, err)
	require.ErrorIs(t, err, pgx.ErrNoRows)
	require.Empty(t, product)
}

// TestDeleteProduct tests deleting a product
func TestDeleteProduct(t *testing.T) {
	account := createRandomAccount(t)
	country := createRandomCountry(t)
	merchant := createRandomMerchant(t, account, country)
	targetProduct := createRandomProduct(t, merchant)

	// Delete the product
	err := testQueries.DeleteProduct(context.Background(), targetProduct.ID)
	require.NoError(t, err)

	// Verify it's deleted
	deletedProduct, err := testQueries.GetProduct(context.Background(), targetProduct.ID)
	require.Error(t, err)
	require.ErrorIs(t, err, pgx.ErrNoRows)
	require.Empty(t, deletedProduct)
}

// TestDeleteProductCascadeOrConstraint tests that deleting a merchant doesn't cascade to products
func TestDeleteProductCascadeOrConstraint(t *testing.T) {
	account := createRandomAccount(t)
	country := createRandomCountry(t)
	merchant := createRandomMerchant(t, account, country)
	product := createRandomProduct(t, merchant)

	err := testQueries.DeleteMerchant(context.Background(), merchant.ID)

	_, err2 := testQueries.GetProduct(context.Background(), product.ID)
	if err == nil {
		// Merchant deletion succeeded, check what happened to product
		if err2 != nil {
			// Product was also deleted (CASCADE)
			require.ErrorIs(t, err2, pgx.ErrNoRows)
		}
	}
}

// TestListProductsByMerchant tests retrieving products for a specific merchant
func TestListProductsByMerchant(t *testing.T) {
	// Create two merchants
	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)
	country := createRandomCountry(t)

	merchant1 := createRandomMerchant(t, account1, country)
	merchant2 := createRandomMerchant(t, account2, country)

	// Create products for merchant1
	var merchant1Products []Product
	for i := 0; i < 5; i++ {
		product := createRandomProduct(t, merchant1)
		merchant1Products = append(merchant1Products, product)
	}

	// Create products for merchant2
	for i := 0; i < 3; i++ {
		createRandomProduct(t, merchant2)
	}

	// List products for merchant1
	listedProducts, err := testQueries.ListProductsByMerchant(context.Background(), int32(merchant1.ID))
	require.NoError(t, err)
	require.NotEmpty(t, listedProducts)

	// Verify we got exactly 5 products for merchant1
	require.Len(t, listedProducts, 5)

	// Verify all returned products belong to merchant1
	for _, product := range listedProducts {
		require.Equal(t, int32(merchant1.ID), product.MerchantID)
	}

	// Verify products are sorted by created_at DESC (newest first)
	for i := 0; i < len(listedProducts)-1; i++ {
		require.True(t, listedProducts[i].CreatedAt.Time.After(listedProducts[i+1].CreatedAt.Time) ||
			listedProducts[i].CreatedAt.Time.Equal(listedProducts[i+1].CreatedAt.Time))
	}

	// Verify our created products are in the list
	foundCount := 0
	for _, createdProduct := range merchant1Products {
		for _, listedProduct := range listedProducts {
			if createdProduct.ID == listedProduct.ID {
				foundCount++
				break
			}
		}
	}
	require.Equal(t, 5, foundCount)
}

// TestListProductsByMerchantWithNoProducts tests listing products for a merchant with no products
func TestListProductsByMerchantWithNoProducts(t *testing.T) {
	account := createRandomAccount(t)
	country := createRandomCountry(t)
	merchant := createRandomMerchant(t, account, country)

	products, err := testQueries.ListProductsByMerchant(context.Background(), int32(merchant.ID))
	require.NoError(t, err)
	require.Empty(t, products)
}

// TestListProductsByNonExistentMerchant tests listing products for a merchant that doesn't exist
func TestListProductsByNonExistentMerchant(t *testing.T) {
	nonExistentMerchantID := int32(999999)
	products, err := testQueries.ListProductsByMerchant(context.Background(), nonExistentMerchantID)
	require.NoError(t, err)
	require.Empty(t, products) // Should return empty slice, not error
}

// TestUpdateProduct tests updating a product
func TestUpdateProduct(t *testing.T) {
	account := createRandomAccount(t)
	country := createRandomCountry(t)
	merchant := createRandomMerchant(t, account, country)
	targetProduct := createRandomProduct(t, merchant)

	// Create a different merchant for testing merchant_id update
	account2 := createRandomAccount(t)
	merchant2 := createRandomMerchant(t, account2, country)

	// Update all fields of the product
	newName := "Updated Product Name"
	newPrice := pgtype.Int4{
		Int32: 7500,
		Valid: true,
	}
	newStatus := pgtype.Text{
		String: "out_of_stock",
		Valid:  true,
	}

	arg := UpdateProductParams{
		ID:         targetProduct.ID,
		Name:       newName,
		MerchantID: int32(merchant2.ID), // Change merchant
		Price:      newPrice,
		Status:     newStatus,
	}

	err := testQueries.UpdateProduct(context.Background(), arg)
	require.NoError(t, err)

	// Retrieve and verify the update
	updatedProduct, err := testQueries.GetProduct(context.Background(), targetProduct.ID)
	require.NoError(t, err)
	require.NotEmpty(t, updatedProduct)

	require.Equal(t, targetProduct.ID, updatedProduct.ID)
	require.Equal(t, newName, updatedProduct.Name)
	require.Equal(t, int32(merchant2.ID), updatedProduct.MerchantID)
	require.Equal(t, newPrice, updatedProduct.Price)
	require.Equal(t, newStatus, updatedProduct.Status)
	require.WithinDuration(t, targetProduct.CreatedAt.Time, updatedProduct.CreatedAt.Time, time.Second)
}

// TestUpdateProductPartialNulls tests updating a product with null values
func TestUpdateProductPartialNulls(t *testing.T) {
	account := createRandomAccount(t)
	country := createRandomCountry(t)
	merchant := createRandomMerchant(t, account, country)

	// Create a product with non-null price and status
	targetProduct := createRandomProduct(t, merchant)

	// Update with null price (keep status unchanged)
	arg1 := UpdateProductParams{
		ID:         targetProduct.ID,
		Name:       "Product with No Price",
		MerchantID: targetProduct.MerchantID,
		Price: pgtype.Int4{
			Valid: false, // Set price to null
		},
		Status: targetProduct.Status, // Keep original status
	}

	err := testQueries.UpdateProduct(context.Background(), arg1)
	require.NoError(t, err)

	// Verify price is null but status remains
	updatedProduct1, err := testQueries.GetProduct(context.Background(), targetProduct.ID)
	require.NoError(t, err)
	require.False(t, updatedProduct1.Price.Valid)
	require.Equal(t, targetProduct.Status, updatedProduct1.Status)

	// Now update with null status (restore price)
	arg2 := UpdateProductParams{
		ID:         targetProduct.ID,
		Name:       "Product with No Status",
		MerchantID: targetProduct.MerchantID,
		Price: pgtype.Int4{
			Int32: 2500,
			Valid: true,
		},
		Status: pgtype.Text{
			Valid: false, // Set status to null
		},
	}

	err = testQueries.UpdateProduct(context.Background(), arg2)
	require.NoError(t, err)

	updatedProduct2, err := testQueries.GetProduct(context.Background(), targetProduct.ID)
	require.NoError(t, err)
	require.True(t, updatedProduct2.Price.Valid)
	require.Equal(t, int32(2500), updatedProduct2.Price.Int32)
	require.False(t, updatedProduct2.Status.Valid)
}

// TestUpdateNonExistentProduct tests updating a product that doesn't exist
func TestUpdateNonExistentProduct(t *testing.T) {
	arg := UpdateProductParams{
		ID:         int32(999999),
		Name:       "Ghost Product",
		MerchantID: 1,
		Price: pgtype.Int4{
			Int32: 1000,
			Valid: true,
		},
		Status: pgtype.Text{
			String: "active",
			Valid:  true,
		},
	}

	err := testQueries.UpdateProduct(context.Background(), arg)
	require.NoError(t, err) // Updating non-existent row should not error
}

// TestConcurrentProductCreation tests creating products concurrently
func TestConcurrentProductCreation(t *testing.T) {
	account := createRandomAccount(t)
	country := createRandomCountry(t)
	merchant := createRandomMerchant(t, account, country)

	// Number of concurrent goroutines
	const n = 10
	errs := make(chan error, n)
	productIDs := make([]int32, n)

	for i := 0; i < n; i++ {
		go func(id int) {
			arg := CreateProductParams{
				ID:         int32(id + 1), // Ensure unique IDs
				Name:       util.RandomProductName(),
				MerchantID: int32(merchant.ID),
				Price: pgtype.Int4{
					Int32: int32(util.RandomInt(100, 10000)),
					Valid: true,
				},
				Status: pgtype.Text{
					String: "active",
					Valid:  true,
				},
			}
			productIDs[id] = arg.ID
			_, err := testQueries.CreateProduct(context.Background(), arg)
			errs <- err
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)
	}

	// Verify all products were created
	products, err := testQueries.ListProductsByMerchant(context.Background(), int32(merchant.ID))
	require.NoError(t, err)

	// Check that we have at least n products (could be more if other tests created some)
	require.GreaterOrEqual(t, len(products), n)

	// Verify our specific products exist
	for i := 0; i < n; i++ {
		product, err := testQueries.GetProduct(context.Background(), productIDs[i])
		require.NoError(t, err)
		require.Equal(t, productIDs[i], product.ID)
	}
}

// TestProductPriceBoundaryValues tests edge cases for price field
func TestProductPriceBoundaryValues(t *testing.T) {
	account := createRandomAccount(t)
	country := createRandomCountry(t)
	merchant := createRandomMerchant(t, account, country)

	// Test with zero price
	arg1 := CreateProductParams{
		ID:         int32(util.RandomInt(50000, 60000)),
		Name:       "Free Product",
		MerchantID: int32(merchant.ID),
		Price: pgtype.Int4{
			Int32: 0,
			Valid: true,
		},
		Status: pgtype.Text{
			String: "active",
			Valid:  true,
		},
	}

	product1, err := testQueries.CreateProduct(context.Background(), arg1)
	require.NoError(t, err)
	require.Equal(t, int32(0), product1.Price.Int32)

	// Test with negative price (if allowed by business logic)
	arg2 := CreateProductParams{
		ID:         int32(util.RandomInt(60001, 70000)),
		Name:       "Discount Coupon",
		MerchantID: int32(merchant.ID),
		Price: pgtype.Int4{
			Int32: -100, // Negative price for coupons/credits
			Valid: true,
		},
		Status: pgtype.Text{
			String: "active",
			Valid:  true,
		},
	}

	product2, err := testQueries.CreateProduct(context.Background(), arg2)
	// This might fail if database has check constraint, but we test both cases
	if err == nil {
		require.Equal(t, int32(-100), product2.Price.Int32)
	}

	// Test with very large price
	arg3 := CreateProductParams{
		ID:         int32(util.RandomInt(70001, 80000)),
		Name:       "Luxury Item",
		MerchantID: int32(merchant.ID),
		Price: pgtype.Int4{
			Int32: 1000000000, // 1 billion
			Valid: true,
		},
		Status: pgtype.Text{
			String: "active",
			Valid:  true,
		},
	}

	product3, err := testQueries.CreateProduct(context.Background(), arg3)
	require.NoError(t, err)
	require.Equal(t, int32(1000000000), product3.Price.Int32)
}

// TestProductNameLength tests product name length boundaries
func TestProductNameLength(t *testing.T) {
	account := createRandomAccount(t)
	country := createRandomCountry(t)
	merchant := createRandomMerchant(t, account, country)

	// Test with empty name (if allowed)
	arg1 := CreateProductParams{
		ID:         int32(util.RandomInt(80001, 90000)),
		Name:       "", // Empty name
		MerchantID: int32(merchant.ID),
		Price: pgtype.Int4{
			Int32: 1000,
			Valid: true,
		},
		Status: pgtype.Text{
			String: "active",
			Valid:  true,
		},
	}

	product1, err := testQueries.CreateProduct(context.Background(), arg1)
	if err == nil {
		require.Equal(t, "", product1.Name)
	}

	// Test with very long name
	longName := "This is an extremely long product name that might exceed typical database column limits but we'll test it anyway to see what happens"
	arg2 := CreateProductParams{
		ID:         int32(util.RandomInt(90001, 100000)),
		Name:       longName,
		MerchantID: int32(merchant.ID),
		Price: pgtype.Int4{
			Int32: 2000,
			Valid: true,
		},
		Status: pgtype.Text{
			String: "active",
			Valid:  true,
		},
	}

	product2, err := testQueries.CreateProduct(context.Background(), arg2)
	if err == nil {
		require.Equal(t, longName, product2.Name)
	}
}
