package profile_runner

import (
	"testing"
	"tsw_controller_app/tswapi"

	"github.com/stretchr/testify/assert"
)

type mockApi struct {
	tswapi.TSWAPI
	activeCab tswapi.TSWAPIActiveCab
}

func (api *mockApi) GetActiveCab() (tswapi.TSWAPIActiveCab, error) {
	return api.activeCab, nil
}

func TestApiController_FormatControlName(t *testing.T) {
	t.Run("no placeholder", func(t *testing.T) {
		api := &mockApi{}
		controller := NewAPIController(api)
		formatted, err := controller.formatControlName("Control_F")
		assert.Nil(t, err, "Expected no error when one cab is active")
		assert.Equal(t, formatted, "Control_F")
	})

	t.Run("no active cab", func(t *testing.T) {
		api := &mockApi{}
		controller := NewAPIController(api)
		_, err := controller.formatControlName("Control_{SIDE}")
		assert.Error(t, err, "Expected formatControlName to return an error when no side is active")
	})

	t.Run("front cab", func(t *testing.T) {
		api := &mockApi{
			activeCab: tswapi.TSWAPIActiveCab{
				Front: true,
			},
		}
		controller := NewAPIController(api)
		formatted, err := controller.formatControlName("Control_{SIDE}")
		assert.Nil(t, err, "Expected no error when one cab is active")
		assert.Equal(t, formatted, "Control_F")
	})

	t.Run("back cab", func(t *testing.T) {
		api := &mockApi{
			activeCab: tswapi.TSWAPIActiveCab{
				Back: true,
			},
		}
		controller := NewAPIController(api)
		formatted, err := controller.formatControlName("Control_{SIDE}")
		assert.Nil(t, err, "Expected no error when one cab is active")
		assert.Equal(t, formatted, "Control_B")
	})

	t.Run("front cab - custom placeholder", func(t *testing.T) {
		api := &mockApi{
			activeCab: tswapi.TSWAPIActiveCab{
				Front: true,
			},
		}
		controller := NewAPIController(api)
		formatted, err := controller.formatControlName("Control_{SIDE:1:2}")
		assert.Nil(t, err, "Expected no error when one cab is active")
		assert.Equal(t, formatted, "Control_1")
	})

	t.Run("back cab - custom placeholder", func(t *testing.T) {
		api := &mockApi{
			activeCab: tswapi.TSWAPIActiveCab{
				Back: true,
			},
		}
		controller := NewAPIController(api)
		formatted, err := controller.formatControlName("Control_{SIDE:1:2}")
		assert.Nil(t, err, "Expected no error when one cab is active")
		assert.Equal(t, formatted, "Control_2")
	})
}
