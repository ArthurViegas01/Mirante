package cv

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/lumni/mirante/internal/platform/tenant"
)

func ctxFor(uid string) context.Context {
	return tenant.WithUserID(context.Background(), uid)
}

// TestUserIsolation: each user has their own master CV; one user's profile,
// skills, experience and education never leak into another's.
func TestUserIsolation(t *testing.T) {
	svc := newService(t)
	ctxA := ctxFor("user-a")
	ctxB := ctxFor("user-b")

	_, err := svc.SaveCV(ctxA, CVInput{
		Nome: "Arthur", Titulo: "Backend Dev", TituloAlvo: "Staff",
		Skills:      []string{"Go", "Docker"},
		Experiences: []ExperienceInput{{Empresa: "Acme", Cargo: "Eng"}},
		Education:   []EducationInput{{Instituicao: "UFRGS", Curso: "CC"}},
	})
	require.NoError(t, err)

	// B has nothing yet — A's CV does not bleed across users.
	pb, err := svc.GetProfile(ctxB)
	require.NoError(t, err)
	require.Equal(t, "", pb.Nome)
	require.Empty(t, pb.Skills)
	require.Empty(t, pb.Experiences)
	require.Empty(t, pb.Education)

	// B saves a distinct CV; A's stays intact.
	_, err = svc.SaveCV(ctxB, CVInput{Nome: "Beatriz", Titulo: "Designer", Skills: []string{"Figma"}})
	require.NoError(t, err)

	pa, err := svc.GetProfile(ctxA)
	require.NoError(t, err)
	require.Equal(t, "Arthur", pa.Nome)
	require.ElementsMatch(t, []string{"Docker", "Go"}, pa.Skills)
	require.Len(t, pa.Experiences, 1)
	require.Len(t, pa.Education, 1)

	pb2, err := svc.GetProfile(ctxB)
	require.NoError(t, err)
	require.Equal(t, "Beatriz", pb2.Nome)
	require.ElementsMatch(t, []string{"Figma"}, pb2.Skills)
	require.Empty(t, pb2.Experiences)
}
