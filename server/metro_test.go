package server

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	commonv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	infrav3 "github.com/paralus/paralus/proto/types/infrapb/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeMetroService implements service.MetroService with function fields so
// each test wires only the method it needs.
type fakeMetroService struct {
	createFn      func(ctx context.Context, metro *infrav3.Location) (*infrav3.Location, error)
	getByIdFn     func(ctx context.Context, id uuid.UUID) (*infrav3.Location, error)
	getByNameFn   func(ctx context.Context, name string) (*infrav3.Location, error)
	getIDByNameFn func(ctx context.Context, name string) (uuid.UUID, error)
	updateFn      func(ctx context.Context, metro *infrav3.Location) (*infrav3.Location, error)
	deleteFn      func(ctx context.Context, metro *infrav3.Location) (*infrav3.Location, error)
	listFn        func(ctx context.Context, partner string) (*infrav3.LocationList, error)
}

func (f *fakeMetroService) Create(ctx context.Context, metro *infrav3.Location) (*infrav3.Location, error) {
	return f.createFn(ctx, metro)
}
func (f *fakeMetroService) GetById(ctx context.Context, id uuid.UUID) (*infrav3.Location, error) {
	return f.getByIdFn(ctx, id)
}
func (f *fakeMetroService) GetByName(ctx context.Context, name string) (*infrav3.Location, error) {
	return f.getByNameFn(ctx, name)
}
func (f *fakeMetroService) GetIDByName(ctx context.Context, name string) (uuid.UUID, error) {
	return f.getIDByNameFn(ctx, name)
}
func (f *fakeMetroService) Update(ctx context.Context, metro *infrav3.Location) (*infrav3.Location, error) {
	return f.updateFn(ctx, metro)
}
func (f *fakeMetroService) Delete(ctx context.Context, metro *infrav3.Location) (*infrav3.Location, error) {
	return f.deleteFn(ctx, metro)
}
func (f *fakeMetroService) List(ctx context.Context, partner string) (*infrav3.LocationList, error) {
	return f.listFn(ctx, partner)
}

func TestLocationServer_CreateLocation(t *testing.T) {
	t.Run("returns the created location on success", func(t *testing.T) {
		want := &infrav3.Location{Metadata: &commonv3.Metadata{Name: "loc1"}}
		ms := &fakeMetroService{createFn: func(ctx context.Context, metro *infrav3.Location) (*infrav3.Location, error) {
			return want, nil
		}}
		s := NewLocationServer(ms)

		got, err := s.CreateLocation(context.Background(), &infrav3.Location{})
		require.NoError(t, err)
		assert.Same(t, want, got)
	})

	t.Run("propagates the service error as a nil response", func(t *testing.T) {
		want := errors.New("boom")
		ms := &fakeMetroService{createFn: func(ctx context.Context, metro *infrav3.Location) (*infrav3.Location, error) {
			return nil, want
		}}
		s := NewLocationServer(ms)

		got, err := s.CreateLocation(context.Background(), &infrav3.Location{})
		assert.Same(t, want, err)
		assert.Nil(t, got)
	})
}

func TestLocationServer_GetLocation(t *testing.T) {
	t.Run("returns the location found by name without falling back to GetById", func(t *testing.T) {
		want := &infrav3.Location{Metadata: &commonv3.Metadata{Name: "loc1"}}
		ms := &fakeMetroService{getByNameFn: func(ctx context.Context, name string) (*infrav3.Location, error) {
			assert.Equal(t, "loc1", name)
			return want, nil
		}}
		s := NewLocationServer(ms)

		got, err := s.GetLocation(context.Background(), &infrav3.Location{Metadata: &commonv3.Metadata{Name: "loc1"}})
		require.NoError(t, err)
		assert.Same(t, want, got)
	})

	t.Run("falls back to GetById when GetByName fails and the metadata Id is a valid UUID", func(t *testing.T) {
		id := uuid.New()
		want := &infrav3.Location{Metadata: &commonv3.Metadata{Id: id.String()}}
		ms := &fakeMetroService{
			getByNameFn: func(ctx context.Context, name string) (*infrav3.Location, error) {
				return nil, errors.New("not found by name")
			},
			getByIdFn: func(ctx context.Context, gotID uuid.UUID) (*infrav3.Location, error) {
				assert.Equal(t, id, gotID)
				return want, nil
			},
		}
		s := NewLocationServer(ms)

		got, err := s.GetLocation(context.Background(), &infrav3.Location{Metadata: &commonv3.Metadata{Id: id.String()}})
		require.NoError(t, err)
		assert.Same(t, want, got)
	})

	t.Run("fails when GetByName fails and the metadata Id is not a valid UUID", func(t *testing.T) {
		ms := &fakeMetroService{getByNameFn: func(ctx context.Context, name string) (*infrav3.Location, error) {
			return nil, errors.New("not found by name")
		}}
		s := NewLocationServer(ms)

		_, err := s.GetLocation(context.Background(), &infrav3.Location{Metadata: &commonv3.Metadata{Id: "not-a-uuid"}})
		assert.Error(t, err)
	})

	t.Run("propagates a GetById error after a valid-UUID fallback", func(t *testing.T) {
		id := uuid.New()
		want := errors.New("boom")
		ms := &fakeMetroService{
			getByNameFn: func(ctx context.Context, name string) (*infrav3.Location, error) {
				return nil, errors.New("not found by name")
			},
			getByIdFn: func(ctx context.Context, gotID uuid.UUID) (*infrav3.Location, error) {
				return nil, want
			},
		}
		s := NewLocationServer(ms)

		_, err := s.GetLocation(context.Background(), &infrav3.Location{Metadata: &commonv3.Metadata{Id: id.String()}})
		assert.Same(t, want, err)
	})
}

func TestLocationServer_DeleteLocation(t *testing.T) {
	t.Run("returns the deleted location on success", func(t *testing.T) {
		want := &infrav3.Location{}
		ms := &fakeMetroService{deleteFn: func(ctx context.Context, metro *infrav3.Location) (*infrav3.Location, error) {
			return want, nil
		}}
		s := NewLocationServer(ms)

		got, err := s.DeleteLocation(context.Background(), &infrav3.Location{})
		require.NoError(t, err)
		assert.Same(t, want, got)
	})

	t.Run("propagates the service error as a nil response", func(t *testing.T) {
		want := errors.New("boom")
		ms := &fakeMetroService{deleteFn: func(ctx context.Context, metro *infrav3.Location) (*infrav3.Location, error) {
			return nil, want
		}}
		s := NewLocationServer(ms)

		got, err := s.DeleteLocation(context.Background(), &infrav3.Location{})
		assert.Same(t, want, err)
		assert.Nil(t, got)
	})
}

func TestLocationServer_UpdateLocation(t *testing.T) {
	want := &infrav3.Location{}
	ms := &fakeMetroService{updateFn: func(ctx context.Context, metro *infrav3.Location) (*infrav3.Location, error) {
		return want, nil
	}}
	s := NewLocationServer(ms)

	got, err := s.UpdateLocation(context.Background(), &infrav3.Location{})
	require.NoError(t, err)
	assert.Same(t, want, got)
}

func TestLocationServer_GetLocations(t *testing.T) {
	t.Run("lists locations scoped to the metadata partner", func(t *testing.T) {
		want := &infrav3.LocationList{}
		var gotPartner string
		ms := &fakeMetroService{listFn: func(ctx context.Context, partner string) (*infrav3.LocationList, error) {
			gotPartner = partner
			return want, nil
		}}
		s := NewLocationServer(ms)

		got, err := s.GetLocations(context.Background(), &infrav3.Location{Metadata: &commonv3.Metadata{Partner: "p1"}})
		require.NoError(t, err)
		assert.Same(t, want, got)
		assert.Equal(t, "p1", gotPartner)
	})

	t.Run("propagates the service error as a nil response", func(t *testing.T) {
		want := errors.New("boom")
		ms := &fakeMetroService{listFn: func(ctx context.Context, partner string) (*infrav3.LocationList, error) {
			return nil, want
		}}
		s := NewLocationServer(ms)

		got, err := s.GetLocations(context.Background(), &infrav3.Location{Metadata: &commonv3.Metadata{}})
		assert.Same(t, want, err)
		assert.Nil(t, got)
	})
}
