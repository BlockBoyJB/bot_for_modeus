package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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

func Test_Bot_handle(t *testing.T) {
	// TODO tests for state and tree
	mockFunc := func(c Context) error { return nil }
	b := &Bot{
		routers: newRouter(),
		storage: newMemoryStorage(),
	}
	b.Command("/hello", mockFunc)
	b.Callback("/some_callback", mockFunc)

	g := b.Group()
	g.Command("/group_cmd", mockFunc)
	g.Message("hello world", mockFunc)

	g2 := g.Group()
	g2.Callback("/internal_callback", mockFunc)

	testCases := []struct {
		testName   string
		update     tgbotapi.Update
		expectFlag bool
	}{
		{
			testName: "cmd hello",
			update: tgbotapi.Update{
				Message: &tgbotapi.Message{
					Text: "/hello",
					Entities: []tgbotapi.MessageEntity{
						{
							Type:   "bot_command",
							Offset: 0,
						},
					},
				},
			},
			expectFlag: true,
		},
		{
			testName: "callback",
			update: tgbotapi.Update{
				CallbackQuery: &tgbotapi.CallbackQuery{
					Data: "/some_callback",
				},
			},
			expectFlag: true,
		},
		{
			testName: "test cmd group",
			update: tgbotapi.Update{
				Message: &tgbotapi.Message{
					Text: "/group_cmd",
					Entities: []tgbotapi.MessageEntity{
						{
							Type:   "bot_command",
							Offset: 0,
						},
					},
				},
			},
			expectFlag: true,
		},
		{
			testName: "group message",
			update: tgbotapi.Update{
				Message: &tgbotapi.Message{
					Text: "hello world",
				},
			},
			expectFlag: true,
		},
		{
			testName: "group in group handler",
			update: tgbotapi.Update{
				CallbackQuery: &tgbotapi.CallbackQuery{
					Data: "/internal_callback",
				},
			},
			expectFlag: true,
		},
		{
			testName: "not found message",
			update: tgbotapi.Update{
				Message: &tgbotapi.Message{
					From: &tgbotapi.User{
						ID: 1,
					},
					Text: "some text that not exist",
				},
			},
			expectFlag: false,
		},
		{
			testName: "not found callback",
			update: tgbotapi.Update{
				CallbackQuery: &tgbotapi.CallbackQuery{
					From: &tgbotapi.User{
						ID: 1,
					},
					Data: "not-found-callback",
				},
			},
			expectFlag: false,
		},
		{
			testName: "not found cmd",
			update: tgbotapi.Update{
				Message: &tgbotapi.Message{
					From: &tgbotapi.User{
						ID: 1,
					},
					Text: "/cmd_not_exist",
					Entities: []tgbotapi.MessageEntity{
						{
							Type:   "bot_command",
							Offset: 0,
						},
					},
				},
			},
			expectFlag: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			ctx := &nativeContext{
				bot:    b,
				update: tc.update,
				params: map[string]string{},
			}
			_, ok := b.handle(ctx, tc.update)
			if tc.expectFlag != ok {
				t.Errorf("not equal, expect %t got %t", tc.expectFlag, ok)
			}
		})
	}
}
