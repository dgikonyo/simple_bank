package db

import (
    "context"
    "testing"

    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgtype"
    "github.com/stretchr/testify/require"
    "simple_bank/util"
)

func createRandomCountry(t *testing.T) Country {
    arg := CreateCountryParams{
        Code: int32(util.RandomInt(1, 1000)),
        Name: pgtype.Text{
            String: util.RandomOwner(),
            Valid:  true,
        },
        ContinentName: pgtype.Text{
            String: util.RandomContinent(), 
            Valid:  true,
        },
    }

    country, err := testQueries.CreateCountry(context.Background(), arg)
    require.NoError(t, err)
    require.NotEmpty(t, country)

    require.Equal(t, arg.Code, country.Code)
    require.Equal(t, arg.Name.String, country.Name.String)
    require.Equal(t, arg.ContinentName.String, country.ContinentName.String)

    return country
}

func TestCreateCountry(t *testing.T) {
    createRandomCountry(t)
}

func TestGetCountry(t *testing.T) {
    country1 := createRandomCountry(t)
    country2, err := testQueries.GetCountry(context.Background(), country1.Code)

    require.NoError(t, err)
    require.NotEmpty(t, country2)

    require.Equal(t, country1.Code, country2.Code)
    require.Equal(t, country1.Name.String, country2.Name.String)
    require.Equal(t, country1.ContinentName.String, country2.ContinentName.String)
}

func TestUpdateCountry(t *testing.T) {
    country1 := createRandomCountry(t)

    arg := UpdateCountryParams{
        Code: country1.Code,
        Name: pgtype.Text{
            String: "Updated Name",
            Valid:  true,
        },
        ContinentName: country1.ContinentName,
    }

    err := testQueries.UpdateCountry(context.Background(), arg)
    require.NoError(t, err)

    updatedCountry, err := testQueries.GetCountry(context.Background(), country1.Code)
    require.NoError(t, err)
    require.NotEmpty(t, updatedCountry)

    require.Equal(t, country1.Code, updatedCountry.Code)
    require.Equal(t, arg.Name.String, updatedCountry.Name.String)
    require.Equal(t, country1.ContinentName.String, updatedCountry.ContinentName.String)
}

func TestDeleteCountry(t *testing.T) {
    country1 := createRandomCountry(t)
    err := testQueries.DeleteCountry(context.Background(), country1.Code)
    require.NoError(t, err)

    country2, err := testQueries.GetCountry(context.Background(), country1.Code)
    require.Error(t, err)
    require.EqualError(t, err, pgx.ErrNoRows.Error())
    require.Empty(t, country2)
}

func TestListCountries(t *testing.T) {
    countries, err := testQueries.ListCountries(context.Background())
    require.NoError(t, err)
    require.NotEmpty(t, countries)

    for _, country := range countries {
        require.NotEmpty(t, country)
    }
}