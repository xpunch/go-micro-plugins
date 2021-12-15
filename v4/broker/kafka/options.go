package kafka

import (
	"context"
	"fmt"
	"strings"

	"github.com/Shopify/sarama"
	"github.com/pkg/errors"
	"go-micro.dev/v4/broker"
	"go-micro.dev/v4/codec"
	"go-micro.dev/v4/logger"
)

var (
	DefaultBrokerConfig  = sarama.NewConfig()
	DefaultClusterConfig = sarama.NewConfig()
)

type brokerConfigKey struct{}
type clusterConfigKey struct{}

func BrokerConfig(c *sarama.Config) broker.Option {
	return setBrokerOption(brokerConfigKey{}, c)
}

func ClusterConfig(c *sarama.Config) broker.Option {
	return setBrokerOption(clusterConfigKey{}, c)
}

type subscribeContextKey struct{}

// SubscribeContext set the context for broker.SubscribeOption
func SubscribeContext(ctx context.Context) broker.SubscribeOption {
	return setSubscribeOption(subscribeContextKey{}, ctx)
}

type subscribeConfigKey struct{}

func SubscribeConfig(c *sarama.Config) broker.SubscribeOption {
	return setSubscribeOption(subscribeConfigKey{}, c)
}

// consumerGroupHandler is the implementation of sarama.ConsumerGroupHandler
type consumerGroupHandler struct {
	handler broker.Handler
	subopts broker.SubscribeOptions
	kopts   broker.Options
	cg      sarama.ConsumerGroup
}

func (*consumerGroupHandler) Setup(sess sarama.ConsumerGroupSession) error {
	logger.Infof("[kafka] %v:%s:%d", sess.Claims(), sess.MemberID(), sess.GenerationID())
	return nil
}

func (*consumerGroupHandler) Cleanup(sess sarama.ConsumerGroupSession) error {
	logger.Infof("[kafka] %v:%s:%d", sess.Claims(), sess.MemberID(), sess.GenerationID())
	return nil
}

func (h *consumerGroupHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		var m broker.Message
		p := &publication{m: &m, t: msg.Topic, km: msg, cg: h.cg, sess: sess}
		eh := h.kopts.ErrorHandler

		if err := h.kopts.Codec.Unmarshal(msg.Value, &m); err != nil {
			p.err = err
			p.m.Body = msg.Value
			if eh != nil {
				eh(p)
			} else {
				logger.Errorf("[kafka]%s failed to unmarshal: %v", h.kopts.Codec.String(), err)
			}
			continue
		}
		if m.Header == nil {
			m.Header = make(map[string]string)
		}
		key := fmt.Sprintf("%s/%d/%d", msg.Topic, msg.Partition, msg.Offset)
		m.Header["key"] = key

		err := h.handler(p)
		if err == nil && h.subopts.AutoAck {
			sess.MarkMessage(msg, "")
		} else if err != nil {
			if errors.Is(err, codec.ErrInvalidMessage) ||
				strings.Contains(err.Error(), "cannot parse invalid wire-format data") ||
				strings.Contains(err.Error(), "unknown field") {
				logger.Errorf("[kafka]unmarshal error:%s: %v", key, err)
				sess.MarkMessage(msg, "")
				continue
			}
			p.err = err
			if eh != nil {
				if err = eh(p); err != nil {
					logger.Errorf("[kafka]subscriber error: %v", err)
					return err
				}
			} else {
				logger.Errorf("[kafka]subscriber error: %v", err)
				return errors.Wrapf(err, "offset: %d", msg.Offset)
			}
		}
	}
	return nil
}
