package gocache

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name string
		conf RedisConfig
		want *Redis
	}{
		{
			name: "Test 1: Successful New",
			conf: RedisConfig{
				Address:  "localhost",
				Port:     6379,
				Password: "",
			},
			want: &Redis{
				Client: redis.NewClient(&redis.Options{
					Addr:     "localhost:6379",
					Password: "",
				}),
			},
		},
		{
			name: "Test 2: Successful New with password",
			conf: RedisConfig{
				Address:  "localhost",
				Port:     6379,
				Password: "password",
			},
			want: &Redis{
				Client: redis.NewClient(&redis.Options{
					Addr:     "localhost:6379",
					Password: "password",
				}),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := New(tt.conf)
			assert.Equal(t, got.Client.Options().Addr, tt.want.Client.Options().Addr)
			assert.Equal(t, got.Client.Options().Password, tt.want.Client.Options().Password)
		})
	}
}

func TestRedisGet(t *testing.T) {
	db, mock := redismock.NewClientMock()

	tests := []struct {
		name    string
		key     string
		value   interface{}
		mock    func()
		wantErr bool
	}{
		{
			name:  "Test 1: Successful Get",
			key:   "key1",
			value: "value1",
			mock: func() {
				mock.ExpectGet("key1").SetVal("\"value1\"")
			},
			wantErr: false,
		},
		{
			name:  "Test 2: Key does not exist",
			key:   "key2",
			value: nil,
			mock: func() {
				mock.ExpectGet("key2").RedisNil()
			},
			wantErr: true,
		},
		{
			name:  "Test 3: Redis error",
			key:   "key3",
			value: nil,
			mock: func() {
				mock.ExpectGet("key3").SetErr(redis.ErrClosed)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			c := &Redis{Client: db}
			var result interface{}
			err := c.Get(context.Background(), tt.key, &result)
			if (err != nil) != tt.wantErr {
				t.Errorf("Redis.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				assert.Equal(t, result, tt.value)
			}
		})
	}
}

func TestRedisPut(t *testing.T) {
	db, mock := redismock.NewClientMock()

	tests := []struct {
		name       string
		key        string
		value      interface{}
		expiration time.Duration
		mock       func()
		wantErr    bool
	}{
		{
			name:       "Test 1: Successful Put",
			key:        "key1",
			value:      "value1",
			expiration: 0,
			mock: func() {
				mock.ExpectSet("key1", "\"value1\"", 0).SetVal("OK")
			},
			wantErr: false,
		},
		{
			name:       "Test 2: JSON Marshal error",
			key:        "key2",
			value:      make(chan int),
			expiration: 0,
			mock:       func() {},
			wantErr:    true,
		},
		{
			name:       "Test 3: Redis error",
			key:        "key3",
			value:      "value3",
			expiration: 0,
			mock: func() {
				mock.ExpectSet("key3", "\"value3\"", 0).SetErr(redis.ErrClosed)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			c := &Redis{Client: db}
			err := c.Put(context.Background(), tt.key, tt.value, tt.expiration)
			if (err != nil) != tt.wantErr {
				t.Errorf("Redis.Put() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestRedisHGetAll(t *testing.T) {
	db, mock := redismock.NewClientMock()

	tests := []struct {
		name    string
		key     string
		value   map[string]string
		mock    func()
		wantErr bool
	}{
		{
			name:  "Test 1: Successful HGetAll",
			key:   "key1",
			value: map[string]string{"field1": "value1", "field2": "value2"},
			mock: func() {
				mock.ExpectHGetAll("key1").SetVal(map[string]string{"field1": "value1", "field2": "value2"})
			},
			wantErr: false,
		},
		{
			name:  "Test 2: Key does not exist",
			key:   "key2",
			value: map[string]string{},
			mock: func() {
				mock.ExpectHGetAll("key2").SetVal(map[string]string{})
			},
			wantErr: true,
		},
		{
			name:  "Test 3: Redis nil",
			key:   "key3",
			value: nil,
			mock: func() {
				mock.ExpectHGetAll("key3").SetErr(redis.Nil)
			},
			wantErr: true,
		},
		{
			name:  "Test 4: Redis error",
			key:   "key4",
			value: nil,
			mock: func() {
				mock.ExpectHGetAll("key4").SetErr(redis.ErrClosed)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			c := &Redis{Client: db}
			result, err := c.HGetAll(context.Background(), tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("Redis.HGetAll() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				assert.Equal(t, result, tt.value)
			}
		})
	}
}

func TestRedisHSet(t *testing.T) {
	db, mock := redismock.NewClientMock()

	tests := []struct {
		name    string
		key     string
		value   map[string]interface{}
		mock    func()
		wantErr bool
	}{
		{
			name:  "Test 1: Successful HSet",
			key:   "key1",
			value: map[string]interface{}{"field1": "value1", "field2": "value2"},
			mock: func() {
				mock.ExpectHSet("key1", map[string]interface{}{"field1": "value1", "field2": "value2"}).SetVal(2)
			},
			wantErr: false,
		},
		{
			name:  "Test 2: Redis error",
			key:   "key2",
			value: map[string]interface{}{"field1": "value1"},
			mock: func() {
				mock.ExpectHSet("key2", map[string]interface{}{"field1": "value1"}).SetErr(redis.ErrClosed)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			c := &Redis{Client: db}
			err := c.HSet(context.Background(), tt.key, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Redis.HSet() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestRedisExpire(t *testing.T) {
	db, mock := redismock.NewClientMock()

	tests := []struct {
		name       string
		key        string
		expiration time.Duration
		mock       func()
		wantErr    bool
	}{
		{
			name:       "Test 1: Successful Expire",
			key:        "key1",
			expiration: time.Minute,
			mock: func() {
				mock.ExpectExpire("key1", time.Minute).SetVal(true)
			},
			wantErr: false,
		},
		{
			name:       "Test 2: Key does not exist",
			key:        "key2",
			expiration: time.Minute,
			mock: func() {
				mock.ExpectExpire("key2", time.Minute).SetVal(false)
			},
			wantErr: false,
		},
		{
			name:       "Test 3: Redis error",
			key:        "key3",
			expiration: time.Minute,
			mock: func() {
				mock.ExpectExpire("key3", time.Minute).SetErr(redis.ErrClosed)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			c := &Redis{Client: db}
			err := c.Expire(context.Background(), tt.key, tt.expiration)
			if (err != nil) != tt.wantErr {
				t.Errorf("Redis.Expire() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestRedisDelete(t *testing.T) {
	db, mock := redismock.NewClientMock()

	tests := []struct {
		name    string
		keys    []string
		mock    func()
		want    int64
		wantErr bool
	}{
		{
			name: "Test 1: Successful Delete",
			keys: []string{"key1", "key2"},
			mock: func() {
				mock.ExpectDel("key1", "key2").SetVal(2)
			},
			want:    2,
			wantErr: false,
		},
		{
			name: "Test 2: No keys exist",
			keys: []string{"key3", "key4"},
			mock: func() {
				mock.ExpectDel("key3", "key4").SetVal(0)
			},
			want:    0,
			wantErr: false,
		},
		{
			name: "Test 3: Redis error",
			keys: []string{"key5"},
			mock: func() {
				mock.ExpectDel("key5").SetErr(redis.ErrClosed)
			},
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			c := &Redis{Client: db}
			got, err := c.Delete(context.Background(), tt.keys...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Redis.Delete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Redis.Delete() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRedisExists(t *testing.T) {
	db, mock := redismock.NewClientMock()

	tests := []struct {
		name    string
		keys    []string
		mock    func()
		want    bool
		wantErr bool
	}{
		{
			name: "Test 1: Keys exist",
			keys: []string{"key1", "key2"},
			mock: func() {
				mock.ExpectExists("key1", "key2").SetVal(2)
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "Test 2: No keys exist",
			keys: []string{"key3", "key4"},
			mock: func() {
				mock.ExpectExists("key3", "key4").SetVal(0)
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "Test 3: Redis error",
			keys: []string{"key5"},
			mock: func() {
				mock.ExpectExists("key5").SetErr(redis.ErrClosed)
			},
			want:    false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			c := &Redis{Client: db}
			got, err := c.Exists(context.Background(), tt.keys...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Redis.Exists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Redis.Exists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRedisIncrement(t *testing.T) {
	db, mock := redismock.NewClientMock()

	tests := []struct {
		name    string
		key     string
		value   int64
		mock    func()
		want    int64
		wantErr bool
	}{
		{
			name:  "Test 1: Successful Increment",
			key:   "key1",
			value: 1,
			mock: func() {
				mock.ExpectIncrBy("key1", 1).SetVal(1)
			},
			want:    1,
			wantErr: false,
		},
		{
			name:  "Test 2: Redis error",
			key:   "key2",
			value: 1,
			mock: func() {
				mock.ExpectIncrBy("key2", 1).SetErr(redis.ErrClosed)
			},
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			c := &Redis{Client: db}
			got, err := c.Increment(context.Background(), tt.key, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Redis.Increment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Redis.Increment() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRedisDecrement(t *testing.T) {
	db, mock := redismock.NewClientMock()

	tests := []struct {
		name    string
		key     string
		value   int64
		mock    func()
		want    int64
		wantErr bool
	}{
		{
			name:  "Test 1: Successful Decrement",
			key:   "key1",
			value: 1,
			mock: func() {
				mock.ExpectDecrBy("key1", 1).SetVal(0)
			},
			want:    0,
			wantErr: false,
		},
		{
			name:  "Test 2: Redis error",
			key:   "key2",
			value: 1,
			mock: func() {
				mock.ExpectDecrBy("key2", 1).SetErr(redis.ErrClosed)
			},
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			c := &Redis{Client: db}
			got, err := c.Decrement(context.Background(), tt.key, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Redis.Decrement() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Redis.Decrement() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRedisKeys(t *testing.T) {
	db, mock := redismock.NewClientMock()

	tests := []struct {
		name    string
		pattern string
		value   []string
		mock    func()
		wantErr bool
	}{
		{
			name:    "Test 1: Successful Keys",
			pattern: "key*",
			value:   []string{"key1", "key2"},
			mock: func() {
				mock.ExpectKeys("key*").SetVal([]string{"key1", "key2"})
			},
			wantErr: false,
		},
		{
			name:    "Test 2: No keys match pattern",
			pattern: "nonexistent*",
			value:   []string{},
			mock: func() {
				mock.ExpectKeys("nonexistent*").SetVal([]string{})
			},
			wantErr: false,
		},
		{
			name:    "Test 3: Redis error",
			pattern: "key*",
			value:   nil,
			mock: func() {
				mock.ExpectKeys("key*").SetErr(redis.ErrClosed)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			c := &Redis{Client: db}
			result, err := c.Keys(context.Background(), tt.pattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("Redis.Keys() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				assert.Equal(t, result, tt.value)
			}
		})
	}
}

func TestRedisTTL(t *testing.T) {
	db, mock := redismock.NewClientMock()

	tests := []struct {
		name    string
		key     string
		value   time.Duration
		mock    func()
		wantErr bool
	}{
		{
			name:  "Test 1: Successful TTL",
			key:   "key1",
			value: time.Minute,
			mock: func() {
				mock.ExpectTTL("key1").SetVal(time.Minute)
			},
			wantErr: false,
		},
		{
			name:  "Test 2: Key does not exist",
			key:   "key2",
			value: -2 * time.Second,
			mock: func() {
				mock.ExpectTTL("key2").SetVal(-2 * time.Second)
			},
			wantErr: false,
		},
		{
			name:  "Test 3: Redis error",
			key:   "key3",
			value: 0,
			mock: func() {
				mock.ExpectTTL("key3").SetErr(redis.ErrClosed)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			c := &Redis{Client: db}
			result, err := c.TTL(context.Background(), tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("Redis.TTL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				assert.Equal(t, result, tt.value)
			}
		})
	}
}

func TestRedisPing(t *testing.T) {
	db, mock := redismock.NewClientMock()

	tests := []struct {
		name    string
		mock    func()
		wantErr bool
	}{
		{
			name: "Test 1: Successful Ping",
			mock: func() {
				mock.ExpectPing().SetVal("PONG")
			},
			wantErr: false,
		},
		{
			name: "Test 2: Redis error",
			mock: func() {
				mock.ExpectPing().SetErr(redis.ErrClosed)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			c := &Redis{Client: db}
			err := c.Ping(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("Redis.Ping() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
