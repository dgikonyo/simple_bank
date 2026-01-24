package db

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
	"simple_bank/util"
)

func createRandomMerchant(t *testing.T, account Account, country Country) Merchant {
	arg := CreateMerchantParams {
		MerchantName: util.RandomOwner(),
		CountryCode: int32(country.Code),
		AdminID: int32(account.ID),
	}

	merchant, err := testQueries.CreateMerchant(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, merchant)

	require.Equal(t, arg.AdminID, merchant.AdminID)
	require.Equal(t, arg.CountryCode, merchant.CountryCode)
	require.Equal(t, arg.MerchantName, merchant.MerchantName)

	require.NotZero(t, merchant.AdminID)
	require.NotZero(t, merchant.CreatedAt)

	return merchant
}

func TestCreateMerchant(t *testing.T) {
	account := createRandomAccount(t)
	country := createRandomCountry(t)
	createRandomMerchant(t, account, country)
}

/**
	here we have created a merchant, 
	and are testing if that merchant exists in the database
*/
func TestGetMerchant(t *testing.T) {
	account := createRandomAccount(t)
	country := createRandomCountry(t)
	initial_merchant := createRandomMerchant(t, account, country)
	existing_merchant, err := testQueries.GetMerchant(context.Background(), initial_merchant.ID)

	require.NoError(t, err)
	require.NotEmpty(t, existing_merchant)

	require.Equal(t, initial_merchant.ID, existing_merchant.ID)
	require.Equal(t, initial_merchant.MerchantName, existing_merchant.MerchantName)
	require.Equal(t, initial_merchant.AdminID, existing_merchant.AdminID)
	require.Equal(t, initial_merchant.CountryCode, existing_merchant.CountryCode)
	require.WithinDuration(t, initial_merchant.CreatedAt.Time, existing_merchant.CreatedAt.Time, time.Second)
}

func TestDeleteMerchant(t *testing.T) {
	account := createRandomAccount(t)
	country := createRandomCountry(t)
	target_merchant := createRandomMerchant(t, account, country)
	
	err := testQueries.DeleteMerchant(context.Background(), target_merchant.ID)
	require.NoError(t, err)

	deleted_merchant, err := testQueries.GetMerchant(context.Background(), target_merchant.ID)
	require.Error(t, err)
	require.EqualError(t, err, pgx.ErrNoRows.Error())
	require.Empty(t, deleted_merchant)
}

func TestListMerchant(t *testing.T) {
	account := createRandomAccount(t)
	country := createRandomCountry(t)

	merchants := make([]Merchant, 5)
	for i := 0; i < 5; i++ {
		merchants[i] = createRandomMerchant(t, account, country)
	}

	listed_merchants, err := testQueries.ListMerchants(context.Background())
	require.NoError(t, err)
    require.NotEmpty(t, listed_merchants)

	// verify that the merchants are in the list
	foundCount := 0
	for _, created := range merchants {
		for _, listed_merchant := range listed_merchants {
			if created.ID == listed_merchant.ID {
				foundCount++
				break
			}
		}
	}

	require.GreaterOrEqual(t, foundCount, 5)
}

func TestUpdateMerchant(t *testing.T){
	account := createRandomAccount(t)
	country := createRandomCountry(t)
	target_merchant := createRandomMerchant(t, account, country)

	arg := UpdateMerchantParams{
		ID: target_merchant.ID,
		MerchantName: "Gikonyo Merchants",
		CountryCode: target_merchant.CountryCode,
		AdminID: target_merchant.AdminID,
	}

	err := testQueries.UpdateMerchant(context.Background(), arg)
	require.NoError(t, err)

	updated_merchant, err := testQueries.GetMerchant(context.Background(), target_merchant.ID)
	require.NoError(t, err)
	require.NotEmpty(t, updated_merchant)

	require.Equal(t, target_merchant.CountryCode, updated_merchant.CountryCode)
	require.Equal(t, target_merchant.AdminID, updated_merchant.AdminID)
}