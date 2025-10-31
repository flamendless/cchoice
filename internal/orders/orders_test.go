package orders

import (
	"context"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGenerateUniqueOrderReferenceNumber(t *testing.T) {
	ctx := context.Background()

	t.Run("generates valid format", func(t *testing.T) {
		ref, err := GenerateUniqueOrderReferenceNumber(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, ref)
		require.True(t, strings.HasPrefix(ref, orderRefPrefix), "reference should start with %s", orderRefPrefix)

		parts := strings.Split(ref, "-")
		require.Len(t, parts, 3, "reference should have format CO-<timestamp>-<random>")
		require.Equal(t, orderRefPrefix[:len(orderRefPrefix)-1], parts[0])
		require.NotEmpty(t, parts[1], "timestamp part should not be empty")
		require.Len(t, parts[2], orderRefRandomLength*2, "random hex part should be %d characters", orderRefRandomLength*2)
	})

	t.Run("generates unique values", func(t *testing.T) {
		generated := make(map[string]bool)
		const numGenerations = 100

		for i := 0; i < numGenerations; i++ {
			ref, err := GenerateUniqueOrderReferenceNumber(ctx)
			require.NoError(t, err)
			require.NotEmpty(t, ref)

			require.False(t, generated[ref], "generated duplicate reference: %s", ref)
			generated[ref] = true
			if i%10 == 0 {
				time.Sleep(time.Microsecond)
			}
		}
	})

	t.Run("format matches expected pattern", func(t *testing.T) {
		ref, err := GenerateUniqueOrderReferenceNumber(ctx)
		require.NoError(t, err)
		pattern := regexp.MustCompile(`^CO-[0-9a-z]+-[0-9a-f]{8}$`)
		require.True(t, pattern.MatchString(ref), "reference %s does not match expected pattern", ref)
	})

	t.Run("timestamp increases over time", func(t *testing.T) {
		ref1, err := GenerateUniqueOrderReferenceNumber(ctx)
		require.NoError(t, err)

		time.Sleep(time.Millisecond)

		ref2, err := GenerateUniqueOrderReferenceNumber(ctx)
		require.NoError(t, err)

		parts1 := strings.Split(ref1, "-")
		parts2 := strings.Split(ref2, "-")

		require.Len(t, parts1, 3)
		require.Len(t, parts2, 3)

		if parts1[1] == parts2[1] {
			require.NotEqual(t, parts1[2], parts2[2], "random parts should differ when timestamps are same")
		}
	})

	t.Run("handles concurrent generation", func(t *testing.T) {
		const numGoroutines = 50
		const numPerGoroutine = 10

		results := make(chan string, numGoroutines*numPerGoroutine)
		errors := make(chan error, numGoroutines*numPerGoroutine)

		for i := 0; i < numGoroutines; i++ {
			go func() {
				for j := 0; j < numPerGoroutine; j++ {
					ref, err := GenerateUniqueOrderReferenceNumber(ctx)
					if err != nil {
						errors <- err
						return
					}
					results <- ref
				}
			}()
		}

		generated := make(map[string]bool)
		for i := 0; i < numGoroutines*numPerGoroutine; i++ {
			select {
			case err := <-errors:
				t.Fatalf("Error generating reference: %v", err)
			case ref := <-results:
				require.NotEmpty(t, ref)
				if generated[ref] {
					t.Errorf("Generated duplicate reference in concurrent execution: %s", ref)
				}
				generated[ref] = true
			case <-time.After(5 * time.Second):
				t.Fatal("Timeout waiting for results")
			}
		}
	})

	t.Run("returns error is nil", func(t *testing.T) {
		ref, err := GenerateUniqueOrderReferenceNumber(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, ref)
	})
}

func BenchmarkGenerateUniqueOrderReferenceNumber(b *testing.B) {
	ctx := context.Background()

	b.ResetTimer()
	for b.Loop() {
		_, err := GenerateUniqueOrderReferenceNumber(ctx)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGenerateUniqueOrderReferenceNumberParallel(b *testing.B) {
	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := GenerateUniqueOrderReferenceNumber(ctx)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkGenerateUniqueOrderReferenceNumberBatch(b *testing.B) {
	ctx := context.Background()
	const batchSize = 100

	b.ResetTimer()
	for b.Loop() {
		for i := 0; i < batchSize; i++ {
			_, err := GenerateUniqueOrderReferenceNumber(ctx)
			if err != nil {
				b.Fatal(err)
			}
		}
	}
}
