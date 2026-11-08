package abstraction

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 本文件覆盖 nats_queue.go 的错误/边界分支:
// - newNATSQueue 默认 URL 分支
// - ensureStream 的 "already exist" 竞态分支
// - Publish / Subscribe 的 ensureStream 错误返回
// - SubscribeSync 错误返回
// - Subscribe goroutine 的超时 / ctx 取消 / 坏 JSON / 缓冲满

// errJetStream 桩: 内嵌真实接口, 只覆盖被测试用到的两个方法.
type errJetStream struct {
	nats.JetStreamContext
	streamInfoErr error
	streamInfoOK  bool
	addStreamErr  error
	subSyncErr    error
}

func (e errJetStream) StreamInfo(stream string, opts ...nats.JSOpt) (*nats.StreamInfo, error) {
	if e.streamInfoOK {
		return &nats.StreamInfo{}, nil
	}
	return nil, e.streamInfoErr
}

func (e errJetStream) AddStream(cfg *nats.StreamConfig, opts ...nats.JSOpt) (*nats.StreamInfo, error) {
	return nil, e.addStreamErr
}

func (e errJetStream) SubscribeSync(subj string, opts ...nats.SubOpt) (*nats.Subscription, error) {
	return nil, e.subSyncErr
}

func TestNewNATSQueue_DefaultURLBranch(t *testing.T) {
	// cfg.NATSURL 为空 → 回落到 nats://127.0.0.1:4222.
	// 若本机没有服务则连接失败(尽快返回); 若有则成功, 关闭即可.
	q, err := newNATSQueue(MessageQueueConfig{})
	if err == nil {
		_ = q.Close()
	}
	// 两种情况都接受, 关键是默认 URL 分支被执行且不 panic
}

func TestNATSQueue_EnsureStream_AlreadyExistsRace(t *testing.T) {
	q := &natsQueue{js: errJetStream{
		streamInfoErr: errors.New("stream not found"),
		addStreamErr:  errors.New("nats: stream name already exists"),
	}}
	// AddStream 报 "already exists" → ensureStream 应返回 nil
	require.NoError(t, q.ensureStream("dup.topic"))
}

func TestNATSQueue_EnsureStream_OtherError(t *testing.T) {
	q := &natsQueue{js: errJetStream{
		streamInfoErr: errors.New("stream not found"),
		addStreamErr:  errors.New("permission denied"),
	}}
	err := q.ensureStream("t")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "permission denied")
}

func TestNATSQueue_Publish_EnsureStreamError(t *testing.T) {
	q := &natsQueue{js: errJetStream{
		streamInfoErr: errors.New("stream not found"),
		addStreamErr:  errors.New("permission denied"),
	}}
	err := q.Publish(context.Background(), "topic", Message{Payload: []byte("x")})
	require.Error(t, err)
}

func TestNATSQueue_Subscribe_EnsureStreamError(t *testing.T) {
	q := &natsQueue{js: errJetStream{
		streamInfoErr: errors.New("stream not found"),
		addStreamErr:  errors.New("permission denied"),
	}}
	_, err := q.Subscribe(context.Background(), "topic", "g1")
	require.Error(t, err)
}

func TestNATSQueue_Subscribe_SubscribeSyncError(t *testing.T) {
	q := &natsQueue{js: errJetStream{
		streamInfoOK: true,
		subSyncErr:   errors.New("sub create failed"),
	}}
	_, err := q.Subscribe(context.Background(), "topic", "g1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "sub create failed")
}

// ---- 以下使用嵌入式 nats-server 的真实 JetStream ----

func TestNATSQueue_Subscribe_CancelClosesChannel(t *testing.T) {
	_, url := startEmbeddedNATS(t)
	q, err := newNATSQueue(MessageQueueConfig{NATSURL: url})
	require.NoError(t, err)
	defer q.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // 立即取消
	ch, err := q.Subscribe(ctx, "cancel.topic", "g1")
	require.NoError(t, err)

	select {
	case _, ok := <-ch:
		assert.False(t, ok, "channel should be closed after ctx cancel")
	case <-time.After(3 * time.Second):
		t.Fatal("channel not closed")
	}
}

func TestNATSQueue_Subscribe_NextMsgTimeoutContinues(t *testing.T) {
	_, url := startEmbeddedNATS(t)
	q, err := newNATSQueue(MessageQueueConfig{NATSURL: url})
	require.NoError(t, err)
	defer q.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	ch, err := q.Subscribe(ctx, "empty.topic", "g1")
	require.NoError(t, err)

	// 无消息 → NextMsg 1s 超时 → continue; 期间 channel 保持打开
	select {
	case m, ok := <-ch:
		_ = m
		t.Fatalf("unexpected message before cancel (open=%v)", ok)
	case <-time.After(400 * time.Millisecond):
	}
	// 等待 ctx 超时后 channel 关闭
	select {
	case _, ok := <-ch:
		assert.False(t, ok)
	case <-time.After(3 * time.Second):
		t.Fatal("channel not closed after ctx timeout")
	}
}

func TestNATSQueue_Subscribe_SkipsBadJSON(t *testing.T) {
	_, url := startEmbeddedNATS(t)
	q, err := newNATSQueue(MessageQueueConfig{NATSURL: url})
	require.NoError(t, err)
	defer q.Close()

	topic := "badjson.topic"
	require.NoError(t, q.ensureStream(topic))
	// 直接注入原始非 JSON 字节(绕过 Publish 的 json.Marshal)
	_, err = q.js.Publish(topic, []byte("{not-json"))
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 700*time.Millisecond)
	defer cancel()
	ch, err := q.Subscribe(ctx, topic, "g1")
	require.NoError(t, err)

	// 坏消息被跳过: channel 不应收到消息, 之后随 ctx 超时关闭
	deadline := time.After(3 * time.Second)
	for {
		select {
		case m, ok := <-ch:
			if !ok {
				return // channel closed → 坏消息确实被跳过
			}
			t.Fatalf("bad json should be skipped, got %+v", m)
		case <-deadline:
			t.Fatal("channel did not close within ctx timeout")
		}
	}
}

func TestNATSQueue_Subscribe_SendBlockedCancels(t *testing.T) {
	_, url := startEmbeddedNATS(t)
	q, err := newNATSQueue(MessageQueueConfig{NATSURL: url})
	require.NoError(t, err)
	defer q.Close()

	topic := "full.topic"
	ctx, cancel := context.WithCancel(context.Background())
	ch, err := q.Subscribe(ctx, topic, "g1")
	require.NoError(t, err)

	// 灌入超过缓冲容量(100)的消息
	for i := 0; i < 150; i++ {
		require.NoError(t, q.Publish(ctx, topic, Message{ID: "m", Payload: []byte("x")}))
	}
	time.Sleep(800 * time.Millisecond) // 让 goroutine 读满缓冲并阻塞在发送
	cancel()
	// 先排空缓冲中的消息, 再确认 channel 关闭
	deadline := time.After(3 * time.Second)
	closed := false
	for {
		select {
		case _, ok := <-ch:
			if !ok {
				closed = true
			}
		case <-deadline:
			t.Fatal("channel not closed")
		}
		if closed {
			break
		}
	}
}