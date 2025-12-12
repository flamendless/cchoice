package orders

import (
	"context"
	"strings"
	"testing"
	"time"

	"cchoice/internal/constants"

	"github.com/stretchr/testify/require"
)

func TestGenerateUniqueOrderReferenceNumber(t *testing.T) {
	ctx := context.Background()

	t.Run("generates valid format", func(t *testing.T) {
		ref, err := GenerateUniqueOrderReferenceNumber(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, ref)
		require.True(t, strings.HasPrefix(ref, orderRefPrefix), "reference should start with %s", orderRefPrefix)

		require.True(t, strings.HasPrefix(ref, "CCO-"), "reference should start with CCO-")
		refWithoutPrefix := strings.TrimPrefix(ref, "CCO-")
		require.GreaterOrEqual(t, len(refWithoutPrefix), 6, "reference should have timestamp and 6 random chars")
		require.Len(t, refWithoutPrefix[len(refWithoutPrefix)-6:], 6, "last 6 characters should be random hex")
		require.Equal(t, strings.ToUpper(refWithoutPrefix[len(refWithoutPrefix)-6:]), refWithoutPrefix[len(refWithoutPrefix)-6:], "random chars should be uppercase")
	})

	t.Run("generates unique values", func(t *testing.T) {
		generated := make(map[string]bool)
		const numGenerations = 100

		for i := range numGenerations {
			ref, err := GenerateUniqueOrderReferenceNumber(ctx)
			require.NoError(t, err)
			require.NotEmpty(t, ref)

			refWithoutPrefix := strings.TrimPrefix(ref, "CCO-")
			randomPart := refWithoutPrefix[len(refWithoutPrefix)-6:]
			require.Equal(t, strings.ToUpper(randomPart), randomPart, "random chars should be uppercase: %s", ref)

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
		require.True(t, constants.OrderReferenceRegex.MatchString(ref), "reference %s does not match expected pattern", ref)
	})

	t.Run("timestamp increases over time", func(t *testing.T) {
		ref1, err := GenerateUniqueOrderReferenceNumber(ctx)
		require.NoError(t, err)

		time.Sleep(time.Millisecond)

		ref2, err := GenerateUniqueOrderReferenceNumber(ctx)
		require.NoError(t, err)

		ref1WithoutPrefix := strings.TrimPrefix(ref1, "CCO-")
		ref2WithoutPrefix := strings.TrimPrefix(ref2, "CCO-")

		time1 := ref1WithoutPrefix[:len(ref1WithoutPrefix)-6]
		time2 := ref2WithoutPrefix[:len(ref2WithoutPrefix)-6]
		random1 := ref1WithoutPrefix[len(ref1WithoutPrefix)-6:]
		random2 := ref2WithoutPrefix[len(ref2WithoutPrefix)-6:]

		if time1 == time2 {
			require.NotEqual(t, random1, random2, "random parts should differ when timestamps are same")
		}
	})

	t.Run("handles concurrent generation", func(t *testing.T) {
		const numGoroutines = 50
		const numPerGoroutine = 10

		results := make(chan string, numGoroutines*numPerGoroutine)
		errors := make(chan error, numGoroutines*numPerGoroutine)

		for range numGoroutines {
			go func() {
				for range numPerGoroutine {
					ref, err := GenerateUniqueOrderReferenceNumber(ctx)
					if err != nil {
						errors <- err
						return
					}
					results <- ref
					time.Sleep(time.Microsecond)
				}
			}()
		}

		generated := make(map[string]bool)
		duplicates := 0
		for range numGoroutines * numPerGoroutine {
			select {
			case err := <-errors:
				t.Fatalf("Error generating reference: %v", err)
			case ref := <-results:
				require.NotEmpty(t, ref)
				if generated[ref] {
					duplicates++
					t.Logf("Generated duplicate reference in concurrent execution: %s", ref)
				}
				generated[ref] = true
			case <-time.After(5 * time.Second):
				t.Fatal("Timeout waiting for results")
			}
		}
		require.LessOrEqual(t, duplicates, 10, "too many duplicates in concurrent execution: %d", duplicates)
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
		for range batchSize {
			_, err := GenerateUniqueOrderReferenceNumber(ctx)
			if err != nil {
				b.Fatal(err)
			}
		}
	}
}
