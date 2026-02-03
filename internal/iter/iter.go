package iter

import (
	"bytes"
	"iter"
	"slices"
)

// Concat joins multiple iter.Seq[T] into a single iter.Seq[T].
func Concat[E any](seqs ...iter.Seq[E]) iter.Seq[E] {
	return func(yield func(E) bool) {
		for _, seq := range seqs {
			if seq == nil {
				continue
			}
			for e := range seq {
				if !yield(e) {
					return
				}
			}
		}
	}
}

// Single returns an iter.Seq[T] that yields exactly one value.
func Single[E any](e E) iter.Seq[E] {
	return func(yield func(E) bool) {
		if !yield(e) {
			return
		}
	}
}

// Map transforms iter.Seq[T] into iter.Seq[U] using f.
func Map[T any, U any](seq iter.Seq[T], f func(T) U) iter.Seq[U] {
	return func(yield func(U) bool) {
		for e := range seq {
			if !yield(f(e)) {
				return
			}
		}
	}
}

// MapSlice converts a slice into iter.Seq[U] by mapping each element.
func MapSlice[T any, U any](s []T, f func(T) U) iter.Seq[U] {
	return Map(slices.Values(s), f)
}

// Filter returns an iter.Seq[T] containing only items matching pred.
func Filter[E any](seq iter.Seq[E], pred func(E) bool) iter.Seq[E] {
	return func(yield func(E) bool) {
		if seq == nil {
			return
		}
		for e := range seq {
			if !pred(e) {
				continue
			}
			if !yield(e) {
				return
			}
		}
	}
}

// JoinString joins values from iter.Seq[string] using sep.
func JoinString[E ~string](seq iter.Seq[E], sep string) E {
	var buf bytes.Buffer
	for part := range seq {
		buf.WriteString(string(part))
		buf.WriteString(sep)
	}
	buf.Truncate(max(buf.Len()-len(sep), 0))
	return E(buf.String())
}
