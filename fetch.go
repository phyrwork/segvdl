package segvdl

import (
	"bytes"
	"context"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"io"
	"log"
)

type Fetcher struct {
	Workers int
	Client *Client
}

func (f *Fetcher) openSegmentsByPattern(ctx context.Context, pattern Pattern, cb func (Segment) error) error {
	i := 0
	for {
		// GET the next segment
		url := pattern.Get(i)
		rsp, err := f.Client.Get(ctx, url)
		if err != nil {
			return errors.Wrapf(err, "do get error", i)
		}
		switch rsp.StatusCode {
		case 200:
			// Emit segment
			if err := cb(Segment{i, rsp}); err != nil {
				return errors.Wrap(err, "[open] segment callback error")
			}
		case 400, 404, 500:
			// End of pool
			rsp.Body.Close()
			return nil
		default:
			// Probably an error
			rsp.Body.Close()
			return errors.Errorf("response error: unexpected status code %v", rsp.StatusCode)
		}
		i++
	}
}

func (f *Fetcher) bufferSegments(ctx context.Context, in <-chan Segment, cb func (Segment) error) error {
	// Don't leak segments
	defer func () {
		for s := range in {
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
			var b bytes.Buffer
			_, err := io.Copy(&b, s.Data)
			if s, ok := s.Data.(io.ReadCloser); ok {
				s.Close()
			}
			if err != nil {
				return errors.Wrapf(err, "copy error", s.Order)
			}
			if err := cb(Segment{s.Order, &b}); err != nil {
				return errors.Wrap(err, "[buffer] segment callback error")
			}
		}
	}
}

func (f *Fetcher) CopyFromPattern(ctx context.Context, w io.Writer, pattern Pattern) error {
	// Determine download parallelism
	workers := 1
	if f.Workers > 0 {
		workers = f.Workers
	}
	// Start the pipeline
	errg, ctx := errgroup.WithContext(ctx)
	// Find segments
	opened := make(chan Segment, workers)
	errg.Go(func () error {
		defer close(opened)
		return f.openSegmentsByPattern(ctx, pattern, func (s Segment) error {
			select {
			case <-ctx.Done():
				if s, ok := s.Data.(io.ReadCloser); ok {
					s.Close()
				}
				return ctx.Err()
			case opened <- s:
				log.Printf("[get] %d", s.Order)
				return nil
			}
		})
	})
	// Download segments
	buffered := make(chan Segment, workers)
	errg.Go(func () error {
		defer close(buffered)
		errg, ctx := errgroup.WithContext(ctx)
		// Download in parallel
		for i := 0; i < workers; i++ {
			errg.Go(func () error {
				return f.bufferSegments(ctx, opened, func (s Segment) error {
					select {
					case <-ctx.Done():
						if s, ok := s.Data.(io.ReadCloser); ok {
							s.Close()
						}
						return ctx.Err()
					case buffered <- s:
						log.Printf("[fetch] %d", s.Order)
						return nil
					}
				})
			})
		}
		return errg.Wait()
	})
	// Reorder segments (out)
	errg.Go(func () error {
		reorder := Reorderer{
			Next: 0,
			Emit: func (s Segment) error {
				_, err := io.Copy(w, s.Data)
				if s, ok := s.Data.(io.ReadCloser); ok {
					s.Close()
				}
				if err != nil {
					return errors.Wrap(err, "copy error")
				}
				log.Printf("[reorder] %d", s.Order)
				return nil
			},
		}
		return reorder.Reorder(ctx, buffered)
	})
	return errg.Wait()
}