package example_test

import (
	"net/http/httptest"
	"testing"

	"github.com/antonydenyer/block-builder-mempool/example"
	"github.com/antonydenyer/block-builder-mempool/testbed"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun/dbfixture"
)

func TestUserHandler(t *testing.T) {
	_, app := testbed.StartApp(t)
	defer app.Stop()

	fixture := loadFixture(t, app)
	testUser := fixture.MustRow("User.test").(*example.User)

	{
		req := testbed.NewRequest("GET", "/api/users", nil)
		resp := httptest.NewRecorder()

		err := app.Router().ServeHTTPError(resp, req)

		require.NoError(t, err)
		require.Contains(t, resp.Body.String(), testUser.Name)
	}

	{
		req := testbed.NewRequest("GET", "/api/users/1", nil)
		resp := httptest.NewRecorder()

		err := app.Router().ServeHTTPError(resp, req)

		require.NoError(t, err)
		require.Contains(t, resp.Body.String(), testUser.Name)
	}
}

func loadFixture(t *testing.T, app *app.App) *dbfixture.Fixture {
	db := app.DB()
	db.RegisterModel((*example.User)(nil), (*example.Org)(nil))

	fixture := dbfixture.New(db, dbfixture.WithRecreateTables())
	err := fixture.Load(app.Context(), app.FS(), "fixture/fixture.yml")
	require.NoError(t, err)

	return fixture
}
