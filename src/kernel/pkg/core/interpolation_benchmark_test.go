package core

import (
	"testing"
)

func BenchmarkInterpolation(b *testing.B) {
	input := map[string]interface{}{
		"email":   "{{state.user_email}}",
		"name":    "{{state.user_name}}",
		"address": "{{state.user_address}}",
		"profile": map[string]interface{}{
			"id":   "{{state.user_id}}",
			"type": "regular",
		},
	}
	hotState := map[string]interface{}{
		"state.user_email":   "alice@example.com",
		"state.user_name":    "Alice Smith",
		"state.user_address": "123 Wonderland Ave",
		"state.user_id":      12345,
	}

	b.Run("Legacy", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = ResolveInterpolation(input, hotState)
		}
	})

	b.Run("Fast", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = FastResolveInterpolation(input, hotState)
		}
	})
}
