package bot

import (
	"testing"
)

// TODO тесты для добавления деревьев

func Test_route_findTree(t *testing.T) {
	mockFunc := func(c Context) error { return nil }
	r := newRoute()
	r.addTree("/user/:type/:date/:schedule_id", mockFunc)
	r.addTree("/friends/action/:schedule_id", mockFunc)

	testCases := []struct {
		testName   string
		path       string
		expectFlag bool
	}{
		{
			testName:   "3 path params",
			path:       "/user/day/2024-01-01/foobar_id",
			expectFlag: true,
		},
		{
			testName:   "1 path param",
			path:       "/friends/action/foobar_id",
			expectFlag: true,
		},
		{
			testName:   "path not exist",
			path:       "/user/foobar",
			expectFlag: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			ctx := &nativeContext{
				params: map[string]string{},
			}
			_, ok := r.findTree(ctx, tc.path)
			if ok != tc.expectFlag {
				t.Errorf("not equal flag: expect %t got %t on path %s", ok, tc.expectFlag, tc.path)
			}
		})
	}
}

func Benchmark_route_findTree_3Params(b *testing.B) {
	mockFunc := func(c Context) error { return nil }
	r := newRoute()
	r.addTree("/user/:type/:date/:schedule_id", mockFunc)

	ctx := &nativeContext{
		params: map[string]string{},
	}

	for i := 0; i < b.N; i++ {
		_, ok := r.findTree(ctx, "/user/day/2024-01-01/5608b53c-8568-4fbe-b72d-c4f278b3f4b6")
		if !ok {
			b.Errorf("handler not found")
		}
		if ctx.Param("type") != "day" || ctx.Param("date") != "2024-01-01" || ctx.Param("schedule_id") != "5608b53c-8568-4fbe-b72d-c4f278b3f4b6" {
			b.Errorf("at least one of path args is invalid or missed")
		}
	}
}
func Benchmark_route_findTree(b *testing.B) {
	mockFunc := func(c Context) error { return nil }
	r := newRoute()
	r.addTree("/friends/action/:schedule_id", mockFunc)

	ctx := &nativeContext{
		params: map[string]string{},
	}

	for i := 0; i < b.N; i++ {
		_, ok := r.findTree(ctx, "/friends/action/5608b53c-8568-4fbe-b72d-c4f278b3f4b6")
		if !ok {
			b.Errorf("handler not found")
		}
		if ctx.Param("schedule_id") != "5608b53c-8568-4fbe-b72d-c4f278b3f4b6" {
			b.Errorf("schedule id is missed")
		}
	}
}
