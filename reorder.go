package segvdl

import (
	"context"
	"github.com/pkg/errors"
	"io"
)

// Reorderer takes a stream of segments and emits
// them in ascending order
type Reorderer struct {
	Next int
	Emit func (Segment) error
	pool map[int]Segment
}

func (r *Reorderer) emit(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		s, ok := r.pool[r.Next]
		if !ok {
			return nil
		}
		delete(r.pool, r.Next)
		if err := r.Emit(s); err != nil {
			return errors.Wrapf(err, "emit error", s.Order)
		}
		r.Next++
	}
}

func (r *Reorderer) Reorder(ctx context.Context, in <-chan Segment) error {
	r.pool = make(map[int]Segment)
	// Don't leak segments
	defer func () {
		for s := range in {
			if s, ok := s.Data.(io.ReadCloser); ok {
				s.Close()
			}
		}
		for _, s := range r.pool {
			if s, ok := s.Data.(io.ReadCloser); ok {
				s.Close()
			}
		}
	}()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case s, ok := <-in:
			if !ok {
				return nil
			}
			// Add to pool
			r.pool[s.Order] = s
			// Emit segments
			if err := r.emit(ctx); err != nil {
				return err
			}
		}
	}
}
