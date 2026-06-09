package monitor

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestServiceStackLabels(t *testing.T) {
	ctx := context.Background()
	mgr := NewManager(NewSQLiteRepo(openTestDB(t)))

	svc, err := mgr.CreateService(ctx, CreateServiceInput{
		ProjectID: "proj1", Nome: "Front", Provider: "netlify", Camada: "frontend",
		Kind: KindHTTP, Target: "https://app.example.test",
	})
	require.NoError(t, err)
	require.Equal(t, "netlify", svc.Provider)
	require.Equal(t, "frontend", svc.Camada)

	// An invalid camada is rejected.
	_, err = mgr.CreateService(ctx, CreateServiceInput{
		ProjectID: "proj1", Nome: "X", Camada: "banco",
		Kind: KindHTTP, Target: "https://x.example.test",
	})
	require.ErrorIs(t, err, ErrInvalid)

	// Update relabels and round-trips through Get.
	prov, camada := "railway", "backend"
	up, err := mgr.UpdateService(ctx, svc.ID, UpdateServiceInput{Provider: &prov, Camada: &camada})
	require.NoError(t, err)
	require.Equal(t, "railway", up.Provider)
	require.Equal(t, "backend", up.Camada)

	got, err := mgr.GetService(ctx, svc.ID)
	require.NoError(t, err)
	require.Equal(t, "railway", got.Provider)
	require.Equal(t, "backend", got.Camada)
}
