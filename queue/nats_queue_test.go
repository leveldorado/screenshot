package queue

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

const ScreenshotTestNATSAddressEnvVariable = "SCREENSHOT_TEST_NATS"

func TestNATS_Subscribe(t *testing.T) {
	address := os.Getenv(ScreenshotTestNATSAddressEnvVariable)
	n, err := NewNATS(address, 10, time.Second)
	require.NoError(t, err)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	topic := uuid.New().String()
	group := uuid.New().String()
	groupSub, err := n.GroupSubscribe(ctx, topic, group)
	require.NoError(t, err)
	reply := uuid.New().String()
	replySub, err := n.Subscribe(ctx, reply)
	require.NoError(t, err)
	reqData := map[string]string{"TEST": "OK"}
	require.NoError(t, n.Publish(ctx, topic, reply, reqData))
	groupMsg := <-groupSub
	msgData := map[string]string{}
	require.NoError(t, json.Unmarshal(groupMsg.Data, &msgData))
	require.Equal(t, reqData, msgData)
	replyData := map[string]string{"OK": "TEST"}
	require.NoError(t, n.Reply(ctx, groupMsg.Reply, replyData))
	replyMsg := <-replySub
	receivedReplyData := map[string]string{}
	require.NoError(t, json.Unmarshal(replyMsg.Data, &receivedReplyData))
	require.Equal(t, replyData, receivedReplyData)
}
