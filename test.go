package test

import (
	"context"
)

type TestLoader struct {
	innerLoader *dataloader.Loader[int32, Test]
}

func NewTestLoader() *TestLoader {
	return &TestLoader{}
}

func (l *TestLoader) getInnerLoader() *dataloader.Loader[int32, Test] {
	if l.innerLoader == nil {
		l.innerLoader = dataloader.NewBatchedLoader(
			func(ctx context.Context, keys []int32) []*dataloader.Result[Test] {
				testMap, err := l.findItemsMap(ctx, keys)

				result := make([]*dataloader.Result[Test], len(keys))
				for i, key := range keys {
					if err != nil {
						result[i] = &dataloader.Result[Test]{Error: err}
						continue
					}

					if test, ok := testMap[key]; ok {
						result[i] = &dataloader.Result[Test]{Data: test}
					} else {
						result[i] = &dataloader.Result[Test]{Error: pgx.ErrNoRows}
					}
				}
				return result
			},
		)
	}
	return l.innerLoader
}

func (l *TestLoader) findItemsMap(ctx context.Context, keys []int32) (map[int32]Test, error) {
	res := make(map[int32]Test, len(keys))

	query := `SELECT * FROM test.test WHERE id = ANY($1)`
	rows, err := db.Query(ctx, query, keys)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var result Test
		err := rows.Scan(
			&result.ID,
			&result.Status,
			&result.ChangedStatus,
			&result.Email,
			&result.Img,
			&result.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		res[result.ID] = result
	}
	return res, nil
}

func (l *TestLoader) Load(ctx context.Context, testKey int32) (Test, error) {
	return l.getInnerLoader().Load(ctx, testKey)()
}
