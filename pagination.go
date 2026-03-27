package intercom

import "context"

// ListOptions specifies cursor-based pagination parameters.
type ListOptions struct {
	PerPage       int    `url:"per_page,omitempty"`
	StartingAfter string `url:"starting_after,omitempty"`
}

// ScrollOptions specifies scroll-based pagination parameters (used by Companies).
type ScrollOptions struct {
	ScrollParam string `url:"scroll_param,omitempty"`
}

// CursorPages represents the pagination metadata in list responses.
type CursorPages struct {
	Type       string             `json:"type"`
	Page       int                `json:"page"`
	PerPage    int                `json:"per_page"`
	TotalPages int                `json:"total_pages"`
	Next       *StartingAfterPage `json:"next,omitempty"`
}

// StartingAfterPage contains the cursor for the next page.
type StartingAfterPage struct {
	PerPage       int    `json:"per_page,omitempty"`
	StartingAfter string `json:"starting_after,omitempty"`
}

// PagedResult is a generic response for paginated list endpoints.
type PagedResult[T any] struct {
	Type       string      `json:"type"`
	Data       []T         `json:"data"`
	TotalCount int         `json:"total_count"`
	Pages      CursorPages `json:"pages"`
}

// PageFetcher is a function that fetches a single page of results.
type PageFetcher[T any] func(ctx context.Context, opts *ListOptions) (*PagedResult[T], error)

// Iter provides lazy iteration over paginated results.
type Iter[T any] struct {
	ctx     context.Context
	fetcher PageFetcher[T]
	opts    *ListOptions

	page    *PagedResult[T]
	index   int
	current T
	err     error
	done    bool
}

// NewIter creates a new paginated iterator. The fetcher function is called
// to retrieve each page of results as needed.
func NewIter[T any](ctx context.Context, opts *ListOptions, fetcher PageFetcher[T]) *Iter[T] {
	if opts == nil {
		opts = &ListOptions{}
	}
	optsCopy := *opts
	return &Iter[T]{
		ctx:     ctx,
		fetcher: fetcher,
		opts:    &optsCopy,
		index:   -1,
	}
}

// Next advances the iterator to the next item. It returns false when there
// are no more items or an error occurs. Use Current() to get the item and
// Err() to check for errors.
func (it *Iter[T]) Next() bool {
	if it.done {
		return false
	}

	// Try to advance within the current page
	if it.page != nil {
		it.index++
		if it.index < len(it.page.Data) {
			it.current = it.page.Data[it.index]
			return true
		}

		// Current page exhausted — check if there's a next page
		if it.page.Pages.Next == nil || it.page.Pages.Next.StartingAfter == "" {
			it.done = true
			return false
		}
		it.opts.StartingAfter = it.page.Pages.Next.StartingAfter
	}

	// Fetch the next (or first) page
	page, err := it.fetcher(it.ctx, it.opts)
	if err != nil {
		it.err = err
		it.done = true
		return false
	}

	it.page = page
	it.index = 0

	if len(page.Data) == 0 {
		it.done = true
		return false
	}

	it.current = page.Data[0]
	return true
}

// Current returns the most recent item yielded by Next().
func (it *Iter[T]) Current() T {
	return it.current
}

// Err returns the first error encountered during iteration.
func (it *Iter[T]) Err() error {
	return it.err
}

// Collect drains the iterator and returns all items as a slice. If an error
// occurs mid-pagination, Collect returns the items collected so far along with
// the error. For zero results it returns a non-nil empty slice.
func (it *Iter[T]) Collect() ([]T, error) {
	var items []T
	for it.Next() {
		items = append(items, it.Current())
	}
	if items == nil {
		items = []T{}
	}
	return items, it.Err()
}

// ForEach calls fn for each item yielded by the iterator. If fn returns a
// non-nil error, iteration stops and that error is returned. If a page fetch
// fails, the fetch error is returned. Callers can distinguish the two by
// checking iter.Err() after ForEach returns: it is non-nil only for fetch errors.
func (it *Iter[T]) ForEach(fn func(T) error) error {
	for it.Next() {
		if err := fn(it.Current()); err != nil {
			return err
		}
	}
	return it.Err()
}

// PageResponse returns the raw paged result for the current page.
func (it *Iter[T]) PageResponse() *PagedResult[T] {
	return it.page
}
